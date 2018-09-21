// +build linux darwin

package pry

import (
	"io"
	"os"

	gotty "github.com/mattn/go-tty"
)

func openTTY() (io.Writer, genericTTY) {
	tty, err := gotty.Open()
	if err != nil {
		panic(err)
	}
	return os.Stdout, tty
}
