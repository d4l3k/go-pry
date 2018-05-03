package pry

import "fmt"

// Fuzz is used for go-fuzz testing.
func Fuzz(data []byte) int {
	s := NewScope()
	val, err := s.InterpretString(string(data))
	if err != nil {
		if val != nil {
			panic(fmt.Sprintf("%#v != nil on error: %+v", val, err))
		}
		return 0
	}
	return 1
}
