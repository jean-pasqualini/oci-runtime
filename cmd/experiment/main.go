package main

import "fmt"

type Dummy struct {
	Color string
}

func debug(v struct{}) {
	fmt.Println(v)
}

type Magic = struct {
}

func main() {
	fmt.Println("hello")
}
