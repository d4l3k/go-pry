package main

import (
	"os"
	"path/filepath"

	"github.com/d4l3k/go-pry/pry"
)

func main() {
	a := filepath.Base("/asdf/asdf")
	pry.Pry()
	os.Setenv("foo", "bar")
	_ = a
}
