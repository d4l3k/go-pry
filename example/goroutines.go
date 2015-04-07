package main

import (
	"fmt"
	"github.com/d4l3k/go-pry/pry"
	"time"
)

func prying() {
	fmt.Println("PRYING!")
}

func main() {
	c := make(chan bool)
	go func() {
		prying()
		pry.Pry()
		c <- true
	}()
	<-c
	for {
		time.Sleep(time.Second)
	}
}
