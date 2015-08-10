package pry

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	if scope.Files == nil {
		scope.Files = map[string]*ast.File{}
	}
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// restore the echoing state when exiting
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	_, filePathRaw, lineNum, _ := runtime.Caller(1)
	filePath := filepath.Dir(filePathRaw) + "/." + filepath.Base(filePathRaw) + "pry"

	err := scope.ConfigureTypes(filePath, lineNum)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nFrom %s @ line %d :\n\n", filePathRaw, lineNum)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
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
		fmt.Printf(" %s %s: %s\n", caret, num, highlightedLine)
	}
	fmt.Println()

	history := []string{}
	currentPos := 0

	line := ""
	count := 0
	index := 0
	b := make([]byte, 1)
	for {
		prompt := fmt.Sprintf("[%d] go-pry> ", currentPos)
		fmt.Printf("\r\033[K%s%s \033[0J\033[%dD", prompt, Highlight(line), len(line)-index+1)
		promptWidth := len(prompt) + index

		// Suggestions
		// TODO(d4l3k): Refactor out
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
		for _, term := range suggestions {
			paddedTerm := term
			for len(paddedTerm) < maxLength {
				paddedTerm += " "
			}
			leftPadding := ""
			for i := 0; i < promptWidth; i++ {
				leftPadding += " "
			}
			fmt.Printf("\n%s%s\033[%dD", leftPadding, ansi.Color(paddedTerm, "white+b:magenta"), len(paddedTerm))
		}
		if len(suggestions) > 0 {
			fmt.Printf("\033[%dA", len(suggestions))
		}

		bPrev := b[0]
		os.Stdin.Read(b)
		switch b[0] {
		default:
			if bPrev == 27 && b[0] == 91 {
				continue
			} else if bPrev == 91 {
				switch b[0] {
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
					continue
				case 65: // Up
					currentPos--
					if currentPos < 0 {
						currentPos = 0
					}
					if len(history) > 0 {
						line = history[currentPos]
					}
					index = len(line)
					continue
				case 67: // Right
					index++
					if index > len(line) {
						index = len(line)
					}
					continue
				case 68: // Left
					index--
					if index < 0 {
						index = 0
					}
					continue
				}
			} else if bPrev == 51 && b[0] == 126 { // DELETE
				line = line[:index-1] + line[index:]
				if len(line) > 0 && index < len(line) {
					line = line[:index] + line[index+1:]
				}
				if index > len(line) {
					index = len(line)
				}
				continue
			}
			line = line[:index] + string(b) + line[index:]
			index++
		case 127: // Backspace
			if len(line) > 0 {
				line = line[:index-1] + line[index:]
				index--
			}
			if index > len(line) {
				index = len(line)
			}
		case 27: // ? This happens on key press
		case 9: //TAB
		case 10: //ENTER
			fmt.Println()
			if len(line) == 0 {
				continue
			}
			if line == "continue" || line == "exit" {
				return
			}
			resp, err := scope.InterpretString(line)
			if err != nil {
				fmt.Println("Error: ", err, resp)
			} else {
				respStr := Highlight(fmt.Sprintf("%#v", resp))
				fmt.Printf("=> %s\033[0J\n", respStr)
			}
			history = append(history, line)
			count++
			currentPos = count
			line = ""
			index = 0
		}
	}
}

// Type returns the reflect.Type of v. Used so packages don't need to import reflect.
func Type(v interface{}) reflect.Type {
	return reflect.TypeOf(v)
}
