package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"
)

func main() {
	decoder := json.NewDecoder(strings.NewReader("[]"))

	var data []any

	err := decoder.Decode(&data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("a: %T, %d\n", data, unsafe.Sizeof(data))
	fmt.Printf("%d\n", unsafe.Sizeof([]string{}))

	sslice := make([]string, 1)
	fmt.Printf("%d\n", unsafe.Sizeof(sslice))

	fmt.Println(data == nil, data)
}
