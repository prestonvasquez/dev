package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type encodetest struct {
	Field1String  string
	Field1Int64   int64
	Field1Float64 float64
	Field2String  string
	Field2Int64   int64
	Field2Float64 float64
	Field3String  string
	Field3Int64   int64
	Field3Float64 float64
	Field4String  string
	Field4Int64   int64
	Field4Float64 float64
}

type nestedtest1 struct {
	Nested nestedtest2
}

type nestedtest2 struct {
	Nested nestedtest3
}

type nestedtest3 struct {
	Nested nestedtest4
}

type nestedtest4 struct {
	Nested nestedtest5
}

type nestedtest5 struct {
	Nested nestedtest6
}

type nestedtest6 struct {
	Nested nestedtest7
}

type nestedtest7 struct {
	Nested nestedtest8
}

type nestedtest8 struct {
	Nested nestedtest9
}

type nestedtest9 struct {
	Nested nestedtest10
}

type nestedtest10 struct {
	Nested nestedtest11
}

type nestedtest11 struct {
	Nested encodetest
}

var encodetestInstance = encodetest{
	Field1String:  "foo",
	Field1Int64:   1,
	Field1Float64: 3.0,
	Field2String:  "bar",
	Field2Int64:   2,
	Field2Float64: 3.1,
	Field3String:  "baz",
	Field3Int64:   3,
	Field3Float64: 3.14,
	Field4String:  "qux",
	Field4Int64:   4,
	Field4Float64: 3.141,
}

var nestedInstance = nestedtest1{
	Nested: nestedtest2{
		Nested: nestedtest3{
			Nested: nestedtest4{
				Nested: nestedtest5{
					Nested: nestedtest6{
						Nested: nestedtest7{
							Nested: nestedtest8{
								Nested: nestedtest9{
									Nested: nestedtest10{
										Nested: nestedtest11{
											Nested: encodetestInstance,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func main() {
	// Create the CPU profile file.
	f, err := os.Create("bson_cpu.pprof")
	if err != nil {
		log.Fatalf("could not create CPU profile: %v", err)
	}
	defer f.Close()

	// Start CPU profiling.
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatalf("could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()

	// Duration for which the profiling will run.
	profileDuration := 30 * time.Second
	log.Printf("Profiling BSON marshaling for %v...", profileDuration)
	endTime := time.Now().Add(profileDuration)

	// Run a loop that repeatedly marshals the instances.
	// Both the flat struct (encodetestInstance) and the deeply nested struct (nestedInstance) are profiled.
	for time.Now().Before(endTime) {
		if _, err := bson.Marshal(encodetestInstance); err != nil {
			log.Printf("Error marshaling encodetestInstance: %v", err)
		}
		if _, err := bson.Marshal(nestedInstance); err != nil {
			log.Printf("Error marshaling nestedInstance: %v", err)
		}
	}

	log.Println("Profiling complete. CPU profile written to bson_cpu.pprof")
}
