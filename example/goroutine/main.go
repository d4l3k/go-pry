package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/d4l3k/go-pry/pry"
)

func main() {
	a := 5

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b := "toast"
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		pry.Pry()
		_ = b
	})
	_ = a

	log.Fatal(http.ListenAndServe(":8080", nil))
}
