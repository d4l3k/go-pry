// +build js

package pry

import (
	"io"
	"log"
	"syscall/js"
)

var tty = newWASMTTY()

func newWASMTTY() *wasmTTY {

	r, w := io.Pipe()
	t := &wasmTTY{
		term: js.Global().Get("term"),
		r:    r,
	}
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		data := args[0].String()
		w.Write([]byte(data))
		return nil
	})
	t.term.Call("on", "data", cb)

	return t
}

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	log.SetOutput(tty)
}

func openTTY() (io.Writer, genericTTY) {
	return tty, tty
}

type wasmTTY struct {
	term js.Value
	r    io.Reader
}

func (t *wasmTTY) Write(buf []byte) (int, error) {
	t.term.Call("write", string(buf))
	return len(buf), nil
}

func (t *wasmTTY) ReadRune() (rune, error) {
	var buf [1]byte
	if _, err := t.r.Read(buf[:]); err != nil {
		return 0, err
	}
	return rune(buf[0]), nil
}

func (t *wasmTTY) Size() (int, int, error) {
	return t.term.Get("cols").Int(), t.term.Get("rows").Int(), nil
}

func (t *wasmTTY) Close() error {
	return nil
}
