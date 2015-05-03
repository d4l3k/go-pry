package main

import "github.com/d4l3k/go-pry/pry"
import "fmt"

func main() {
	for i := 0; i < 10; i++ {
		pry.Pry()
	}
	fmt.Println("DUCK")
}
