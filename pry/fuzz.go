package pry

// Fuzz is used for go-fuzz testing.
func Fuzz(data []byte) int {
	s := NewScope()
	val, err := s.InterpretString(string(data))
	if err != nil {
		if val != nil {
			panic("val != nil on error")
		}
		return 0
	}
	return 1
}
