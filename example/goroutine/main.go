package main

import (
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/d4l3k/go-pry/pry"
)

func main() {
	w := 6

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b := "toast"
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		pry.Pry()
		_ = b
	})

	pry.Pry()

	log.Fatal(http.ListenAndServe(":8080", nil))
	_ = w
}
