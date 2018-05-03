package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
)

const out = "fuzz/corpus/"

var (
	exampleRegexpQuotes = regexp.MustCompile("InterpretString\\((\"(.*)\"|`(.*)`)\\)")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	files, err := filepath.Glob("**/*.go")
	if err != nil {
		return err
	}
	for _, fpath := range files {
		body, err := ioutil.ReadFile(fpath)
		if err != nil {
			return err
		}
		matches := exampleRegexpQuotes.FindAllSubmatch(body, -1)
		for _, match := range matches {
			expr := match[2]
		}
	}
	return nil
}
