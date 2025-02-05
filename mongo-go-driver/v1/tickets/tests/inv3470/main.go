package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	k := myInt(1)

	v := struct {
		T myType `json:"t"`
		I *myInt `json:"i"`
	}{
		I: &k,
	}

	err := json.Unmarshal([]byte(`{"t": null, "i": null}`), &v)
	if err != nil {
		panic(err)
	}

	fmt.Println(v.I)
}

type myType struct {
	A string `json:"a"`
}

func (t *myType) UnmarshalJSON(b []byte) error {
	fmt.Println("myType")
	return nil
}

type myInt int64

func (m *myInt) UnmarshalJSON(b []byte) error {
	fmt.Println("myInt")
	return nil
}
