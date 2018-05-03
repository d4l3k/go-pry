package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const out = "fuzz/corpus/"

var (
	exampleRegexpQuotes = regexp.MustCompile("(?s)InterpretString\\(`(.*?)`\\)")
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
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	for _, fpath := range files {
		body, err := ioutil.ReadFile(fpath)
		if err != nil {
			return err
		}

		for {
			match := exampleRegexpQuotes.FindSubmatchIndex(body)
			if match == nil {
				break
			}

			expr := bytes.TrimSpace(body[match[2]:match[3]])
			hash := sha1.Sum(expr)
			file := hex.EncodeToString(hash[:])
			if err := ioutil.WriteFile(filepath.Join(out, file), expr, 0644); err != nil {
				return err
			}

			body = body[match[1]:]
		}
	}
	return nil
}
