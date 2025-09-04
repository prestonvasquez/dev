package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("expected 1 timeout argument")
	}

	timeoutArg := os.Args[1]

	timeout, err := strconv.Atoi(timeoutArg)
	if err != nil {
		log.Fatalf("could not convert timeout argument to integer: %v", err)
	}

	blockTime := (3.0 / 20.0) * float64(timeout)
	maxAwaitTime := float64(timeout) / 2.0
	buffer := (maxAwaitTime - blockTime) / 70.0
	getMoreBound := maxAwaitTime - blockTime + buffer

	fmt.Printf("blockTime: %dms\n", int(blockTime))
	fmt.Printf("maxAwaitTime: %dms\n", int(maxAwaitTime))
	fmt.Printf("getMoreBound: %dms\n", int(getMoreBound))
}
