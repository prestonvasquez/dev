package main

import "fmt"

type Parent struct {
	Sort, Hint any
}

//type Child1 struct {
//	Hint any
//}
//
//type Child2 struct {
//	Sort, Hint any
//}

func main() {
	print("string")
	print(1)
}

type printType interface {
	int | string
}

func print[T printType](t T) {
	fmt.Println(t)
}
