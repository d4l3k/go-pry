package main

import (
	"github.com/d4l3k/go-pry/pry"

	"log"
)

// Note: This file has some gibberish to test for highlighting and other edge cases.

/*
	Block Quote
*/

func X() bool {
	return true
}

type Banana struct {
	Name string
	Cake []int
}

func (b Banana) Ly() string {
	return b.Name + "ly"
}

func main() {
	a := 1
	b := Banana{"Jeoffry", []int{1, 2, 3}}
	m := []int{1234}
	_ = m

	testMake := make(chan int, 1)
	testMap := map[int]interface{}{
		1: 2,
		3: "asdf",
		5: []interface{}{
			1, "asdf",
		},
	}
	_ = testMap
	go func() {
		_ = 1 + 1*1/1%1
	}()

	if d := X(); d {
		log.Println(d)
		for i, j := range []int{1} {
			k := 1
			log.Println(i, j, k)
			// Example comment
			pry.Pry()
		}
	}
	log.Println("Test", a, b, main, testMake)
}
