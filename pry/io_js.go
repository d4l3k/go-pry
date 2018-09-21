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
	path = filepath.Base(path)

	r, w := io.Pipe()
	var respCB js.Callback
	respCB = js.NewCallback(func(args []js.Value) {
		defer respCB.Release()

		var textCB js.Callback
		textCB = js.NewCallback(func(args []js.Value) {
			defer textCB.Release()

			w.Write([]byte(args[0].String()))
			w.Close()
		})
		args[0].Call("text").Call("then", textCB)
	})
	js.Global().Call("fetch", path).Call("then", respCB)
	return ioutil.ReadAll(r)
}

func loadHistory() []string {
	hist := js.Global().Get("localStorage").Get("history")
	if hist.Type() == js.TypeUndefined {
		return nil
	}
	var history []string
	if err := json.Unmarshal([]byte(hist.String()), &history); err != nil {
		panic(err)
	}
	return history
}

func saveHistory(history *[]string) {
	bytes, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}
	js.Global().Get("localStorage").Set("history", string(bytes))
}
