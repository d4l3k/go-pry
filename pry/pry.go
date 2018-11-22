package pry

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"go/ast"

	"github.com/mgutz/ansi"
)

// Pry does nothing. It only exists so running code without go-pry doesn't throw an error.
func Pry(v ...interface{}) {
}

// Apply drops into a pry shell in the location required.
func Apply(scope *Scope) {
	out, tty := openTTY()
	defer tty.Close()

	_, filePathRaw, lineNum, _ := runtime.Caller(1)
	filePath := filepath.Dir(filePathRaw) + "/." + filepath.Base(filePathRaw) + "pry"

	if err := apply(scope, out, tty, filePath, filePathRaw, lineNum); err != nil {
		panic(err)
	}
}

type genericTTY interface {
	ReadRune() (rune, error)
	Size() (int, int, error)
	Close() error
}

func apply(
	scope *Scope,
	out io.Writer,
	tty genericTTY,
	filePath, filePathRaw string,
	lineNum int,
) error {
	if scope.Files == nil {
		scope.Files = map[string]*ast.File{}
	}

	if err := scope.ConfigureTypes(filePath, lineNum); err != nil {
		return err
	}

	displayFilePosition(out, filePathRaw, filePath, lineNum)

	history, err := NewHistory()
	if err != nil {
		fmt.Errorf("Failed to initiliaze history %+v", err)
	}
	if err := history.Load(); err != nil {
		fmt.Errorf("Failed to load the history %+v", err)
	}

	currentPos := history.Len()

	line := ""
	count := history.Len()
	index := 0
	r := rune(0)
	for {
		prompt := fmt.Sprintf("[%d] go-pry> ", currentPos)
		fmt.Fprintf(out, "\r\033[K%s%s \033[0J\033[%dD", prompt, Highlight(line), len(line)-index+1)

		promptWidth := len(prompt) + index
		displaySuggestions(scope, out, tty, line, index, promptWidth)

		bPrev := r

		r = 0
		for r == 0 {
			var err error
			r, err = tty.ReadRune()
			if err != nil {
				return err
			}
		}
		switch r {
		default:
			if bPrev == 27 && r == 91 {
				continue
			} else if bPrev == 91 {
				switch r {
				case 66: // Down
					currentPos++
					if history.Len() < currentPos {
						currentPos = history.Len()
					}
					if history.Len() == currentPos {
						line = ""
					} else {
						line = history.Records[currentPos]
					}
					index = len(line)
				case 65: // Up
					currentPos--
					if currentPos < 0 {
						currentPos = 0
					}
					if history.Len() > 0 {
						line = history.Records[currentPos]
					}
					index = len(line)
				case 67: // Right
					index++
					if index > len(line) {
						index = len(line)
					}
				case 68: // Left
					index--
					if index < 0 {
						index = 0
					}
				}
				continue
			} else if bPrev == 51 && r == 126 { // DELETE
				if len(line) > 0 && index < len(line) {
					line = line[:index] + line[index+1:]
				}
				if index > len(line) {
					index = len(line)
				}
				continue
			}
			line = line[:index] + string(r) + line[index:]
			index++
		case 127, '\b': // Backspace
			if len(line) > 0 && index > 0 {
				line = line[:index-1] + line[index:]
				index--
			}
			if index > len(line) {
				index = len(line)
			}
		case 27: // ? This happens on key press
		case 9: //TAB
		case 10, 13: //ENTER
			fmt.Fprintln(out, "\033[100000C\033[0J")
			if len(line) == 0 {
				continue
			}
			if line == "continue" || line == "exit" {
				return nil
			}
			resp, err := scope.InterpretString(line)
			if err != nil {
				fmt.Fprintln(out, "Error: ", err, resp)
			} else {
				respStr := Highlight(fmt.Sprintf("%#v", resp))
				fmt.Fprintf(out, "=> %s\n", respStr)
			}
			history.Add(line)
			err = history.Save()
			if err != nil {
				fmt.Fprintln(out, "Error: ", err)
			}

			count++
			currentPos = count
			line = ""
			index = 0
		case 4: // Ctrl-D
			fmt.Fprintln(out)
			return nil
		}
	}
}

func displayFilePosition(
	out io.Writer, filePathRaw, filePath string, lineNum int,
) {
	fmt.Fprintf(out, "\nFrom %s @ line %d :\n\n", filePathRaw, lineNum)
	file, err := readFile(filePath)
	if err != nil {
		fmt.Fprintln(out, err)
	}
	lines := strings.Split((string)(file), "\n")
	lineNum--
	start := lineNum - 5
	if start < 0 {
		start = 0
	}
	end := lineNum + 6
	if end > len(lines) {
		end = len(lines)
	}
	maxLen := len(fmt.Sprint(end))
	for i := start; i < end; i++ {
		caret := "  "
		if i == lineNum {
			caret = "=>"
		}
		numStr := fmt.Sprint(i + 1)
		if len(numStr) < maxLen {
			numStr = " " + numStr
		}
		num := ansi.Color(numStr, "blue+b")
		highlightedLine := Highlight(strings.Replace(lines[i], "\t", "  ", -1))
		fmt.Fprintf(out, " %s %s: %s\n", caret, num, highlightedLine)
	}
	fmt.Fprintln(out)
}

// displaySuggestions renders the live autocomplete from GoCode.
func displaySuggestions(
	scope *Scope,
	out io.Writer,
	tty genericTTY,
	line string,
	index, promptWidth int,
) {
	var err error
	var suggestions []string
	if runtime.GOOS == "js" {
		suggestions, err = scope.SuggestionsPry(line, index)
	} else {
		suggestions, err = scope.SuggestionsGoCode(line, index)
	}
	if err != nil {
		suggestions = []string{"ERR", err.Error()}
	}

	maxLength := 0
	if len(suggestions) > 10 {
		suggestions = suggestions[:10]
	}
	for _, term := range suggestions {
		if len(term) > maxLength {
			maxLength = len(term)
		}
	}
	termWidth, _, _ := tty.Size()
	for _, term := range suggestions {
		paddedTerm := term
		for len(paddedTerm) < maxLength {
			paddedTerm += " "
		}
		var leftPadding string
		for i := 0; i < promptWidth; i++ {
			leftPadding += " "
		}
		if promptWidth > termWidth {
			return
		} else if len(paddedTerm)+promptWidth > termWidth {
			paddedTerm = paddedTerm[:termWidth-promptWidth]
		}
		fmt.Fprintf(out, "\n%s%s\033[%dD", leftPadding, ansi.Color(paddedTerm, "white+b:magenta"), len(paddedTerm))
	}
	if len(suggestions) > 0 {
		fmt.Fprintf(out, "\033[%dA", len(suggestions))
	}
}
