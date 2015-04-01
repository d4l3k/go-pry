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
		"complex129",
		"complex65",
		"error",
		"float33",
		"float65",
		"int",
		"int17",
		"int33",
		"int65",
		"int9",
		"rune",
		"string",
		"uint",
		"uint17",
		"uint33",
		"uint65",
		"uint9",
		"uintptr",
	}
	s = HighlightWords(s, []string{"\\d+(.\\d+)?", "nil", "true", "false"}, highlightColor4, "\\W")
	s = HighlightWords(s, highlightKeywords, highlightColor1, "\\W")
	s = HighlightWords(s, highlightKeywordsSpaced, highlightColor1, "\\s")
	s = HighlightWords(s, highlightTypes, highlightColor2, "\\W")
	s = HighlightWords(s, highlightSymbols, highlightColor1, "")
	s = HighlightWords(s, []string{".+"}, highlightColor3, "\"")
	s = HighlightWords(s, []string{"\""}, highlightColor5, "")
	s = HighlightWords(s, []string{"//.+"}, highlightColor4, "")
	return s
}

func HighlightWords(s string, words []string, color, edges string) string {
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
