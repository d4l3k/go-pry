package pry

import (
	"io/ioutil"
	"regexp"
	"testing"
)

// Make sure the highlighter doesn't change the code.
func TestHighlightSafe(t *testing.T) {
	t.Parallel()

	fileBytes, err := ioutil.ReadFile("../example/file/file.go")
	if err != nil {
		t.Error(err)
	}
	fileStr := (string)(fileBytes)
	highlight := Highlight(fileStr)

	r, err := regexp.Compile("\\x1b\\[(.*?)m")
	if err != nil {
		t.Error(err)
	}

	// Strip Bash control sequences
	s := r.ReplaceAllLiteralString(highlight, "")

	if s != fileStr {
		t.Error("Highlighting has changed the code!")
	}
}
