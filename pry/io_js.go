// +build js

package pry

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"path/filepath"
	"syscall/js"
)

func readFile(path string) ([]byte, error) {
	path = filepath.Join("bundles", filepath.Base(path))

	r, w := io.Pipe()
	var respCB js.Func
	respCB = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer respCB.Release()

		var textCB js.Func
		textCB = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer textCB.Release()

			w.Write([]byte(args[0].String()))
			w.Close()

			return nil
		})
		args[0].Call("text").Call("then", textCB)

		return nil
	})
	js.Global().Call("fetch", path).Call("then", respCB)
	return ioutil.ReadAll(r)
}

type browserHistory struct {
	Records []string
}

// NewHistory constructs browserHistory instance
func NewHistory() (*browserHistory, error) {

	// FIXME:
	// when localStorage is full, can be return an error

	return &browserHistory{}, nil
}

// Load unmarshal localStorage data into history's records
func (bh *browserHistory) Load() error {
	hist := js.Global().Get("localStorage").Get("history")
	if hist.Type() == js.TypeUndefined {
		return nil // nothing to unmarashal
	}
	var records []string
	if err := json.Unmarshal([]byte(hist.String()), &records); err != nil {
		return err
	}
	bh.Records = records

	return nil
}

// Save saves marshaled history's records into localStorage
func (bh browserHistory) Save() error {
	bytes, err := json.Marshal(bh.Records)
	if err != nil {
		return err
	}
	js.Global().Get("localStorage").Set("history", string(bytes))

	return nil
}

// Len returns amount of records in history
func (bh browserHistory) Len() int { return len(bh.Records) }

// Add appends record into history's records
func (bh *browserHistory) Add(record string) {
	bh.Records = append(bh.Records, record)
}
