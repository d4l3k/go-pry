package main

import (
	"fmt"
	"github.com/d4l3k/go-pry/pry"
	"time"
)

func main() {
	c := make(chan bool)
	go func() {
		pry.Pry()
		fmt.Println("PRYING!")
		c <- true
	}()
	<-c
	for {
		time.Sleep(time.Second)
	}
}
