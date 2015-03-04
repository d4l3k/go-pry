package pry

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

func Pry(v ...interface{}) {
}

func Apply(v map[string]interface{}) {
	fmt.Println(v)
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	// restore the echoing state when exiting
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

	line := ""
	count := 1
	var b []byte = make([]byte, 1)
	for {
		fmt.Printf("\r[%d] go-pry> %s", count, line)
		os.Stdin.Read(b)
		switch b[0] {
		default:
			line += string(b)
		case 127: // Backspace
			if len(line) > 0 {
				line = line[:len(line)-1]
			}
		case 9: //TAB
			if len(line) > 0 && line[len(line)-1] == '.' {
				val, present := v[line[:len(line)-1]]
				if present {
					typeOf := reflect.TypeOf(val)
					fmt.Println()
					methods := make([]string, typeOf.NumMethod())
					for i, _ := range methods {
						methods[i] = typeOf.Method(i).Name
					}
					fmt.Println(typeOf.Name() + ": " + strings.Join(methods, " "))
				}
			}
		case 10: //ENTER
			fmt.Println()
			val, present := v[line]
			if present {
				fmt.Println(val)
			} else if line != "" {
				fmt.Println("ERROR: Invalid input!")
			}
			count += 1
			line = ""
		}
	}
}
