package pry

import (
	"github.com/mgutz/ansi"

	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
)

func Pry(v ...interface{}) {
}

func Apply(v map[string]interface{}) {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// restore the echoing state when exiting
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	_, filePathRaw, lineNum, _ := runtime.Caller(1)
	filePath := strings.TrimSuffix(filePathRaw, ".go")
	fmt.Printf("\nFrom %s @ line %d :\n\n", filePath, lineNum)
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}
	lines := strings.Split((string)(file), "\n")
	lineNum -= 1
	start := lineNum - 5
	if start < 0 {
		start = 0
	}
	end := lineNum + 5
	if end > len(lines) {
		end = len(lines)
	}
	maxLength := len(fmt.Sprint(end))
	for i := start; i <= end; i++ {
		caret := "  "
		if i == lineNum {
			caret = "=>"
		}
		numStr := fmt.Sprint(i)
		if len(numStr) < maxLength {
			numStr = " " + numStr
		}
		num := ansi.Color(numStr, "blue+b")
		line := strings.Replace(lines[i], "\t", "  ", -1)
		fmt.Printf(" %s %s: %s\n", caret, num, Highlight(line))
	}
	fmt.Println()

	line := ""
	count := 1
	var b []byte = make([]byte, 1)
	for {
		fmt.Printf("\r[%d] go-pry> %s \033[1D", count, line)
		os.Stdin.Read(b)
		switch b[0] {
		default:
			line += string(b)
		case 127: // Backspace
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		case 27: // ? These two happen on key press
		case 91: // ?
		case 65: // Up
		case 66: // Down
		case 67: // Right
		case 68: // Left
		case 9: //TAB
			if len(line) == 0 {
				fmt.Println()
				for k, _ := range v {
					fmt.Print(k + " ")
				}
				fmt.Println()
			} else if line[len(line)-1] == '.' {
				val, present := v[line[:len(line)-1]]
				if present {
					typeOf := reflect.TypeOf(val)
					fmt.Println()
					methods := make([]string, typeOf.NumMethod())
					for i, _ := range methods {
						methods[i] = typeOf.Method(i).Name + "("
					}
					fields := make([]string, typeOf.NumField())
					for i, _ := range fields {
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
			count += 1
			line = ""
		}
	}
}
