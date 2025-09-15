package main

import "fmt"

// go run --tags medium ./cmd/experiment --root abc run --color chocolate cid
func main() {
	var s string = "a"
	if true {
		s := "b"
		fmt.Println(s)
	}
	fmt.Println(s)
}
