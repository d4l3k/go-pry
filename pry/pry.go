package pry

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"go/ast"

	"github.com/mattn/go-colorable"
	gotty "github.com/mattn/go-tty"
	"github.com/mgutz/ansi"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	out io.Writer = os.Stdout
	tty *gotty.TTY
)

var historyFile = ".go-pry_history"

func historyPath() (string, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return path.Join(dir, historyFile), nil
}

func loadHistory() []string {
	path, err := historyPath()
	if err != nil {
		log.Printf("Error finding user home dir: %s", err)
		return nil
	}
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	var history []string
	if err := json.Unmarshal(body, &history); err != nil {
		log.Printf("Error reading history file! %s", err)
		return nil
	}
	return history
}

func saveHistory(history *[]string) {
	body, err := json.Marshal(history)
	if err != nil {
		log.Printf("Err marshalling history: %s", err)
	}
	path, err := historyPath()
	if err != nil {
		log.Printf("Error finding user home dir: %s", err)
		return
	}
	if err := ioutil.WriteFile(path, body, 0755); err != nil {
		log.Printf("Error writing history: %s", err)
	}
}

// Pry does nothing. It only exists so running code without go-pry doesn't throw an error.
func Pry(v ...interface{}) {
}

// Apply drops into a pry shell in the location required.
func Apply(scope *Scope) {
	if runtime.GOOS == "windows" {
		out = colorable.NewColorableStdout()
	}
	var err error
	tty, err = gotty.Open()
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	if scope.Files == nil {
		scope.Files = map[string]*ast.File{}
	}

	_, filePathRaw, lineNum, _ := runtime.Caller(1)
	filePath := filepath.Dir(filePathRaw) + "/." + filepath.Base(filePathRaw) + "pry"

	if err := scope.ConfigureTypes(filePath, lineNum); err != nil {
		panic(err)
	}

	displayFilePosition(filePathRaw, filePath, lineNum)

	history := loadHistory()
	defer saveHistory(&history)

	currentPos := len(history)

	line := ""
	count := 0
	index := 0
	r := rune(0)
	for {
		prompt := fmt.Sprintf("[%d] go-pry> ", currentPos)
		fmt.Fprintf(out, "\r\033[K%s%s \033[0J\033[%dD", prompt, Highlight(line), len(line)-index+1)

		promptWidth := len(prompt) + index
		displaySuggestions(scope, line, index, promptWidth)

		bPrev := r

		r = 0
		for r == 0 {
			r, err = tty.ReadRune()
			if err != nil {
				panic(err)
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
					if len(history) < currentPos {
						currentPos = len(history)
					}
					if len(history) == currentPos {
						line = ""
					} else {
						line = history[currentPos]
					}
					index = len(line)
				case 65: // Up
					currentPos--
					if currentPos < 0 {
						currentPos = 0
					}
					if len(history) > 0 {
						line = history[currentPos]
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
				return
			}
			resp, err := scope.InterpretString(line)
			if err != nil {
				fmt.Fprintln(out, "Error: ", err, resp)
			} else {
				respStr := Highlight(fmt.Sprintf("%#v", resp))
				fmt.Fprintf(out, "=> %s\n", respStr)
			}
			history = append(history, line)
			count++
			currentPos = count
			line = ""
			index = 0
		case 4: // Ctrl-D
			fmt.Fprintln(out)
			return
		}
	}
}

func displayFilePosition(filePathRaw, filePath string, lineNum int) {
	fmt.Fprintf(out, "\nFrom %s @ line %d :\n\n", filePathRaw, lineNum)
	file, err := ioutil.ReadFile(filePath)
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
func displaySuggestions(scope *Scope, line string, index, promptWidth int) {
	// Suggestions
	suggestions, err := scope.SuggestionsGoCode(line, index)
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
