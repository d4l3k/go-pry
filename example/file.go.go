package main

import (
	"../pry"

	"log"
)

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
	if d := X(); d {
		log.Println(d)
		for i, j := range []int{1} {
			k := 1
			log.Println(i, j, k)
			// Example comment
			pry.Apply(map[string]interface{}{ "X": X, "main": main, "a": a, "b": b, "m": m, "d": d, "i": i, "j": j, "k": k, "log": pry.Package{Name: "log", Functions: map[string]interface{}{"Ltime": log.Ltime, "LstdFlags": log.LstdFlags, "Prefix": log.Prefix, "Fatalf": log.Fatalf, "Fatalln": log.Fatalln, "Panicf": log.Panicf, "Llongfile": log.Llongfile, "Lshortfile": log.Lshortfile, "SetPrefix": log.SetPrefix, "Print": log.Print, "Fatal": log.Fatal, "Panic": log.Panic, "Panicln": log.Panicln, "Ldate": log.Ldate, "New": log.New, "Flags": log.Flags, "SetFlags": log.SetFlags, "Println": log.Println, "Printf": log.Printf, "Lmicroseconds": log.Lmicroseconds, "Logger": log.Logger{}, "SetOutput": log.SetOutput, }}, })

		}
	}
	log.Println("Test", a, b, main)
}
