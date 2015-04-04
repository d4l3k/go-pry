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

	"github.com/mgutz/ansi"
)

// Pry does nothing. It only exists so running code without go-pry doesn't throw an error.
func Pry(v ...interface{}) {
}

// Apply drops into a pry shell in the location required.
func Apply(v Scope) {
	for _, v := range v {
		p, isPackage := v.(Package)
		if isPackage {
			p.Process()
		}
	}
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// restore the echoing state when exiting
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	_, filePathRaw, lineNum, _ := runtime.Caller(1)
	filePath := filepath.Dir(filePathRaw) + "/." + filepath.Base(filePathRaw) + "pry"
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
	maxLength := len(fmt.Sprint(end))
	for i := start; i < end; i++ {
		caret := "  "
		if i == lineNum {
			caret = "=>"
		}
		numStr := fmt.Sprint(i + 1)
		if len(numStr) < maxLength {
			numStr = " " + numStr
		}
		num := ansi.Color(numStr, "blue+b")
		line := strings.Replace(lines[i], "\t", "  ", -1)
		fmt.Printf(" %s %s: %s\n", caret, num, Highlight(line))
	}
	fmt.Println()

	history := []string{}
	currentPos := 0

	line := ""
	count := 0
	b := make([]byte, 1)
	for {
		fmt.Printf("\r\033[K[%d] go-pry> %s \033[1D", currentPos, Highlight(line))
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
					continue
				case 65: // Up
					currentPos--
					if currentPos < 0 {
						currentPos = 0
					}
					if len(history) > 0 {
						line = history[currentPos]
					}
					continue
				case 67: // Right
					continue
				case 68: // Left
					continue
				}
			}
			line += string(b)
		case 127: // Backspace
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		case 27: // ? These two happen on key press
		case 9: //TAB
			if len(line) == 0 {
				fmt.Println()
				for k := range v {
					fmt.Print(k + " ")
				}
				fmt.Println()
			} else if line[len(line)-1] == '.' {
				val, present := v[line[:len(line)-1]]
				if present {
					typeOf := reflect.TypeOf(val)
					fmt.Println()
					methods := make([]string, typeOf.NumMethod())
					for i := range methods {
						methods[i] = typeOf.Method(i).Name + "("
					}
					fields := make([]string, typeOf.NumField())
					for i := range fields {
						fields[i] = typeOf.Field(i).Name
					}
					fmt.Println(typeOf.Name() + ": " + strings.Join(fields, " ") + " " + strings.Join(methods, " "))
				}
			}
		case 10: //ENTER
			fmt.Println()
			if len(line) == 0 {
				continue
			}
			if line == "continue" || line == "exit" {
				return
			}
			resp, err := InterpretString(v, line)
			if err != nil {
				fmt.Println("Error: ", err)
			} else {
				respStr := Highlight(fmt.Sprintf("%#v", resp))
				fmt.Printf("=> %s\n", respStr)
			}
			history = append(history, line)
			count++
			currentPos = count
			line = ""
		}
	}
}
