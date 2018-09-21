// +build windows

package pry

import (
	"io"

	colorable "github.com/mattn/go-colorable"
	gotty "github.com/mattn/go-tty"
)

func openTTY() (io.Writer, genericTTY) {
	tty, err := gotty.Open()
	if err != nil {
		panic(err)
	}
	return colorable.NewColorableStdout(), tty
}
