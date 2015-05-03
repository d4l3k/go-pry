package main

import (
	"fmt"
	"unsafe"
)

func a() {
	fmt.Println("A")
}

func b() {
	fmt.Println("B")
}

func main() {
	funcRef := &a
	ptr := unsafe.Pointer(a)
}
