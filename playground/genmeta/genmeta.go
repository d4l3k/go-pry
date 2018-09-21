package main

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/d4l3k/go-pry/pry"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	dir, err := parser.ParseDir(token.NewFileSet(), wd, nil, 0)
	if err != nil {
		return err
	}
	imp := pry.JSImporter{
		Dir: dir,
	}
	spew.Dump(imp)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(imp); err != nil {
		return err
	}
	if err := ioutil.WriteFile("meta.go", []byte(
		`package main
import "github.com/d4l3k/go-pry/pry"
func init(){
	pry.InternalSetImports(`+"`"+buf.String()+"`"+`)
}`,
	), 0644); err != nil {
		return err
	}
	return nil
}
