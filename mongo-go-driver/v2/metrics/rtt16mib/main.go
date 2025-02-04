package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	// We'll aim for ~16 MiB of actual data in the document.
	// The max BSON doc size is 16 MB (~16.77 MiB), so you may need to reduce if you get an error.
	DocumentSize = 16 * 1024 * 1024
)

// Flags
var (
	iface      = flag.String("iface", "lo", "Network interface to capture on")
	mongoURI   = flag.String("uri", "mongodb://localhost:27017", "MongoDB URI")
	dbName     = flag.String("db", "testdb", "Database name")
	collName   = flag.String("coll", "testcoll", "Collection name")
	docID      = flag.String("id", "large_doc", "Document _id to upsert/fetch")
	portFilter = flag.Int("port", 27017, "MongoDB port (for BPF capture filter)")
)

func main() {
	flag.Parse()

	// 1. Connect to MongoDB
	clientOpts := options.Client().ApplyURI(*mongoURI)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// 2. Upsert a large document (~16 MiB)
	coll := client.Database(*dbName).Collection(*collName)

	bigData := make([]byte, DocumentSize)
	_, _ = rand.Read(bigData) // fill with random bytes, or just leave zeroed
	doc := bson.M{
		"_id":  *docID,
		"data": bigData,
	}
	upsert := true
	_, err = coll.ReplaceOne(
		context.Background(),
		bson.M{"_id": *docID},
		doc,
		&options.ReplaceOptions{Upsert: &upsert},
	)
	if err != nil {
		log.Fatalf("Failed to upsert large doc: %v", err)
	}

	// 3. Start a packet capture on the given interface
	// NOTE: you generally need root or CAP_NET_RAW privileges to do this.
	handle, err := pcap.OpenLive(*iface, 65535, false, pcap.BlockForever)
	if err != nil {
		log.Fatalf("pcap OpenLive failed: %v", err)
	}
	defer handle.Close()

	// Set a BPF filter to see only traffic on the MongoDB port
	filter := fmt.Sprintf("tcp and port %d", *portFilter)
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatalf("Failed to set BPF filter: %v", err)
	}
	log.Printf("Capturing on interface %q with filter %q\n", *iface, filter)

	// We'll parse the server->client traffic to count:
	// - how many reply messages (wire-protocol)
	// - total "body" bytes (excluding the 16-byte header).
	var (
		captureWg      sync.WaitGroup
		totalBodyBytes int64
		totalReplies   int64
		mu             sync.Mutex
	)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	captureWg.Add(1)

	// A buffer to do naive reassembly of server->client data
	// keyed by 4-tuple (srcIP, srcPort, dstIP, dstPort) if you want multiple connections.
	// But for simplicity, assume we only have one direct connection to local Mongo.
	var serverToClientBuffer []byte

	go func() {
		defer captureWg.Done()
		for packet := range packetSource.Packets() {
			// We only care about TCP layer
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer == nil {
				continue
			}
			tcp, _ := tcpLayer.(*layers.TCP)
			if tcp == nil {
				continue
			}

			// If the packet is from server->client, i.e. source port = 27017 by default
			// (or the user-specified port).
			if int(tcp.SrcPort) == *portFilter {
				// Append payload to our buffer
				mu.Lock()
				serverToClientBuffer = append(serverToClientBuffer, tcp.Payload...)
				// Try to parse complete wire-protocol messages in a loop
				for {
					if len(serverToClientBuffer) < 4 {
						// Not enough to read length
						break
					}
					// The first 4 bytes = total message length (int32, little-endian).
					msgLen := int32(binary.LittleEndian.Uint32(serverToClientBuffer[:4]))
					if msgLen < 16 {
						// Invalid or incomplete
						break
					}
					if len(serverToClientBuffer) < int(msgLen) {
						// We don't have the full message yet
						break
					}
					// We have a complete wire-protocol message of length msgLen
					// Skip the 16-byte header. The body is (msgLen - 16).
					bodySize := int64(msgLen - 16)
					totalBodyBytes += bodySize
					totalReplies++

					// Remove this message from the buffer
					serverToClientBuffer = serverToClientBuffer[msgLen:]
				}
				mu.Unlock()
			}
		}
	}()

	// 4. Perform the Find to fetch the large doc. We'll measure how long it takes.
	start := time.Now()
	var result bson.M
	err = coll.FindOne(context.Background(), bson.M{"_id": *docID}).Decode(&result)
	if err != nil {
		log.Fatalf("FindOne error: %v", err)
	}
	elapsed := time.Since(start)

	// 5. We're done reading. Give a short grace period for final packets (e.g. ACK/FIN).
	time.Sleep(500 * time.Millisecond)

	// Stop the packet source by closing the handle (which will cause the packet loop to end).
	handle.Close()
	captureWg.Wait()

	// 6. Print the final stats
	sec := elapsed.Seconds()
	mbps := float64(DocumentSize) / (1024.0 * 1024.0) / sec

	mu.Lock()
	bodyBytes := totalBodyBytes
	replies := totalReplies
	mu.Unlock()

	fmt.Printf("\n--- Mongo Fetch Stats ---\n")
	fmt.Printf("Fetched ~%d bytes of document data (application-level) in %.3f seconds\n", DocumentSize, sec)
	fmt.Printf("Effective throughput (doc-size based): %.2f MiB/s\n", mbps)

	fmt.Printf("\n--- Wire Protocol Capture ---\n")
	fmt.Printf("Total wire-protocol replies: %d\n", replies)
	fmt.Printf("Sum of message bodies (excluding 16-byte header): %d bytes\n", bodyBytes)
	if sec > 0 {
		wireMbps := float64(bodyBytes) / (1024.0 * 1024.0) / sec
		fmt.Printf("Throughput (body-bytes only): %.2f MiB/s\n", wireMbps)
	} else {
		fmt.Printf("Throughput (body-bytes only): (instant?)\n")
	}
}
