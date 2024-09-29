package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(`[]`)))

	var data []any

	err := decoder.Decode(&data)
	if err != nil {
		panic(err)
	}

	fmt.Println(data == nil, data)
}
