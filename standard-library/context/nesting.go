//go:build r1

package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	//to1 := 5 * time.Minute
	//to2 := 10 * time.Minute
	//
	var to2 time.Duration

	//ctx, cancel := context.WithTimeout(context.Background(), to1)
	//defer cancel()
	ctx := context.Background()

	ctx2, cancel := context.WithTimeout(ctx, to2)
	defer cancel()

	dl, _ := ctx2.Deadline()
	fmt.Println(time.Until(dl), ctx2.Err())
}
