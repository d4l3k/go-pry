package pry

import (
	"github.com/mgutz/ansi"

	"regexp"
	"strings"
)

const highlightColor1 = "white+b"
const highlightColor2 = "green+b"
const highlightColor3 = "red"
const highlightColor4 = "blue+b"
const highlightColor5 = "red+b"

// Highlight highlights a string of go code for outputting to bash.
func Highlight(s string) string {
	highlightSymbols := []string{"==", "!=", ":=", "="}
	highlightKeywords := []string{
		"for", "defer", "func", "struct", "switch", "case",
		"interface", "if", "range", "bool", "type", "package", "import",
		"make", "append",
	}
	highlightKeywordsSpaced := []string{"go"}
	highlightTypes := []string{
		"byte",
		"complex128",
		"complex64",
		"error",
		"float",
		"float32",
		"float64",
		"int",
		"int16",
		"int32",
		"int64",
		"int8",
		"rune",
		"string",
		"uint",
		"uint16",
		"uint32",
		"uint64",
		"uint8",
		"uintptr",
	}
	s = highlightWords(s, []string{"-?(0[xX])?\\d+((\\.|e-?)\\d+)*", "nil", "true", "false"}, highlightColor4, "\\W")
	s = highlightWords(s, highlightKeywords, highlightColor1, "\\W")
	s = highlightWords(s, highlightKeywordsSpaced, highlightColor1, "\\s")
	s = highlightWords(s, highlightTypes, highlightColor2, "\\W")
	s = highlightWords(s, highlightSymbols, highlightColor1, "")
	s = highlightWords(s, []string{".+"}, highlightColor3, "\"")
	s = highlightWords(s, []string{"\""}, highlightColor5, "")
	s = highlightWords(s, []string{"//.+"}, highlightColor4, "")
	return s
}

func highlightWords(s string, words []string, color, edges string) string {
	lE := len(edges) - strings.Count(edges, "\\")
	s = " " + s + " "
	for _, word := range words {
		r, _ := regexp.Compile(edges + word + edges)
		s = (string)(r.ReplaceAllFunc(([]byte)(s), func(b []byte) []byte {
			bStr := string(b)
			return []byte(bStr[0:lE] + ansi.Color(bStr[lE:len(bStr)-lE], color) + bStr[len(bStr)-lE:])
		}))
	}
	if s[0] == ' ' {
		s = s[1:]
	}
	if s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}
