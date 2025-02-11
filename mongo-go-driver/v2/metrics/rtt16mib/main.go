package main

import (
	"context"
	"crypto/rand"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// We'll aim for ~16 MiB of actual data in the document.
// The max BSON doc size is 16 MB (~16.77 MiB), so you may need to reduce if you get an error.
const docSize = 16*1024*1024 - 100
const dbName = "rtt16mibtest"
const collName = "simple"

type connState int

const (
	stateIdle connState = iota
	stateWaitingReply
)

type roundTripConn struct {
	address string
	net.Conn
	mu             sync.Mutex
	roundTripCount int
	state          connState
	byteCountList  []int
	rttDursMS      []float64
	debug          bool
}

func (rtc *roundTripConn) reset() {
	rtc.mu.Lock()
	defer rtc.mu.Unlock()

	rtc.roundTripCount = 0
	rtc.byteCountList = []int{}
	rtc.rttDursMS = []float64{}
}

func (rtc *roundTripConn) Write(p []byte) (int, error) {
	n, err := rtc.Conn.Write(p)
	if err != nil {
		return n, err
	}

	rtc.mu.Lock()
	defer rtc.mu.Unlock()
	if rtc.state == stateIdle && n > 0 {
		rtc.state = stateWaitingReply
	}

	return n, nil
}

func (rtc *roundTripConn) Read(p []byte) (int, error) {
	start := time.Now()
	n, err := rtc.Conn.Read(p)
	if n > 0 {
		dur := time.Since(start)
		rtc.mu.Lock()
		if rtc.debug {
			if rtc.state == stateWaitingReply {
				rtc.roundTripCount++
				rtc.byteCountList = append(rtc.byteCountList, n)
				rtc.rttDursMS = append(rtc.rttDursMS, float64(dur.Nanoseconds()))
			}
		}
		rtc.mu.Unlock()
	}

	return n, err
}

type debugDialer struct {
	conns []*roundTripConn
}

var _ options.ContextDialer = &debugDialer{}

func (d *debugDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := net.DialTimeout(network, address, 10*time.Second)
	if err != nil {
		return nil, err
	}

	rtc := &roundTripConn{Conn: c, state: stateIdle, address: address}
	d.conns = append(d.conns, rtc)

	return rtc, nil
}

func (d *debugDialer) resetRtcs() {
	for _, rtc := range d.conns {
		rtc.reset()
	}
}

func (d *debugDialer) debugOn() {
	for _, rtc := range d.conns {
		rtc.mu.Lock()
		rtc.debug = true
		rtc.mu.Unlock()
	}
}

func (d *debugDialer) debugOff() {
	for _, rtc := range d.conns {
		rtc.mu.Lock()
		rtc.debug = false
		rtc.mu.Unlock()
	}
}

func main() {
	customerDialer := &debugDialer{}
	clientOpts := options.Client().
		ApplyURI(os.Getenv("MONGODB_URI")).
		SetDialer(customerDialer).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	log.Println("established client")

	// Make sure the connection is established (handshake, etc.).
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("ping error: %v", err)
	}

	log.Println("finished ping")

	coll := client.Database(dbName).Collection(collName)
	_ = coll.Drop(context.Background())

	buf := make([]byte, docSize)
	_, _ = rand.Read(buf)

	doc := bson.D{{Key: "data", Value: buf}}
	if _, err := coll.InsertOne(context.Background(), doc); err != nil {
		log.Fatalf("failed to insert doc: %v", err)
	}

	log.Println("inserted doc")

	customerDialer.debugOn()

	res := coll.FindOne(context.Background(), bson.D{})
	if err := res.Err(); err != nil {
		log.Fatalf("failed to find one: %v", err)
	}

	customerDialer.debugOff()

	log.Println("found one")

	for _, rtc := range customerDialer.conns {
		if len(rtc.byteCountList) == 0 {
			continue
		}

		rtc.mu.Lock()
		// Skip the first value in bytes and dur as it corresponds to the header.
		log.Printf("Address: %s, recv cycles: %v, non-header byte count: %v, non-header total time (ms): %v\n",
			rtc.address, rtc.roundTripCount, sum(rtc.byteCountList[1:]), sum(rtc.rttDursMS[1:])/1000000)
		rtc.mu.Unlock()
	}
}

// average calculates the mean of a slice of float64 numbers.
func average(data []float64) float64 {
	if len(data) == 0 {
		return 0 // Handle empty slice to avoid division by zero
	}
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

func sum[T int | float64](data []T) T {
	var s T
	for _, n := range data {
		s += n
	}
	return s
}
