package pry

import (
	"reflect"
	"testing"
)

// Literals
func TestStringLiteral(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("\"Hello!\"")
	if err != nil {
		t.Error(err)
	}
	if out != "Hello!" {
		t.Error("Expected Hello!")
	}
}
func TestIntLiteral(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("-1234")
	if err != nil {
		t.Error(err)
	}
	if out != -1234 {
		t.Error("Expected -1234")
	}
}
func TestCharLiteral(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("'a'")
	if err != nil {
		t.Error(err)
	}
	if out != 'a' {
		t.Errorf("Expected 'a' got %#v.", out)
	}
}
func TestArrayLiteral(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("[]int{1,2,3,4}")
	if err != nil {
		t.Error(err)
	}
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestMapLiteral(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("map[string]int{\"duck\": 5,\n \"banana\": -123,\n}")
	if err != nil {
		t.Error(err)
	}
	expected := map[string]int{
		"duck":   5,
		"banana": -123,
	}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestTypeCast(t *testing.T) {
	scope := NewScope()
	scope.Set("a", -1234.0)
	out, err := scope.InterpretString("int(a)")
	if err != nil {
		t.Error(err)
	}
	expected := -1234
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// Selectors and Ident
func TestBasicIdent(t *testing.T) {
	scope := NewScope()
	scope.Set("a", 5)
	out, err := scope.InterpretString("a")
	if err != nil {
		t.Error(err)
	}
	expected := 5
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestMissingBasicIdent(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("a")
	if err == nil || out != nil {
		t.Error("Found non-existant ident.")
	}
}
func TestMapIdent(t *testing.T) {
	scope := NewScope()
	scope.Set("a", map[string]int{
		"B": 10,
	})
	out, err := scope.InterpretString("a[\"B\"]")
	if err != nil {
		t.Error(err)
	}
	expected := 10
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestMissingMapIdent(t *testing.T) {
	scope := NewScope()
	scope.Set("a", map[string]int{})

	out, err := scope.InterpretString("a[\"b\"]")
	if err != nil {
		t.Error(err)
	}
	if out != 0 {
		t.Error("Found non-existant ident.")
	}
}
func TestArrIdent(t *testing.T) {
	scope := NewScope()
	scope.Set("a", []int{1, 2, 3})

	out, err := scope.InterpretString("a[1]")
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestMissingArrIdent(t *testing.T) {
	scope := NewScope()
	scope.Set("a", []int{1})

	out, err := scope.InterpretString("a[1]")
	if err == nil || out != nil {
		t.Error("Should have thrown out of range error")
	}
}

func TestSlice(t *testing.T) {
	scope := NewScope()
	scope.Set("a", []int{1, 2, 3, 4})

	out, err := scope.InterpretString("a[1:3]")
	if err != nil {
		t.Error(err)
	}
	expected := []int{2, 3}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// Structs
type testStruct struct {
	A    int
	C, D string
}

func (a testStruct) B() int {
	return a.A
}

func TestSelector(t *testing.T) {
	scope := NewScope()
	scope.Set("a", testStruct{A: 1})

	out, err := scope.InterpretString("a.A")
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestSelectorFunc(t *testing.T) {
	scope := NewScope()
	scope.Set("a", testStruct{A: 1})

	out, err := scope.InterpretString("a.B()")
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// Basic Math
func TestBasicMath(t *testing.T) {
	scope := NewScope()
	pairs := map[string]interface{}{
		"2*3":        6,
		"2.0 * 3.0":  6.0,
		"10 / 2":     5,
		"10.0 / 2.0": 5.0,
		"1 + 2":      3,
		"1.0 + 2.0":  3.0,
	}
	for k, expected := range pairs {
		out, err := scope.InterpretString(k)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(expected, out) {
			t.Errorf("Expected %#v got %#v.", expected, out)
		}
	}
}

func TestParens(t *testing.T) {
	scope := NewScope()
	scope.Set("a", 5)

	out, err := scope.InterpretString("((10) * (a))")
	if err != nil {
		t.Error(err)
	}
	expected := 50
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// Test Make
func TestMakeSlice(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("make([]int, 1, 10)")
	if err != nil {
		t.Error(err)
	}
	expected := make([]int, 1, 10)
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestMakeChan(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("make(chan int, 10)")
	if err != nil {
		t.Error(err)
	}
	expected := make(chan int, 10)
	if reflect.TypeOf(expected) != reflect.TypeOf(out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestMakeUnknown(t *testing.T) {
	scope := NewScope()
	out, err := scope.InterpretString("make(int)")
	if err == nil || out != nil {
		t.Error("Should have thrown error.")
	}
}

func TestAppend(t *testing.T) {
	scope := NewScope()
	scope.Set("a", []int{1})

	out, err := scope.InterpretString("append(a, 2, 3)")
	if err != nil {
		t.Error(err)
	}
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// TODO Assignment
func TestDeclareAssign(t *testing.T) {
	scope := NewScope()
	scope.Set("a", []int{1})

	out, err := scope.InterpretString("b := 2")
	if err != nil {
		t.Error(err)
	}
	expected := 2
	out, _ = scope.Get("b")
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestAssign(t *testing.T) {
	scope := NewScope()
	scope.Set("a", 1)

	out, err := scope.InterpretString("a = 2")
	if err != nil {
		t.Error(err)
	}
	expected := 2
	out, _ = scope.Get("a")
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// Statements

func TestFuncDeclAndCall(t *testing.T) {
	scope := NewScope()

	out, err := scope.InterpretString("a := func(){ return 5 }")
	if err != nil {
		t.Error(err)
	}
	out, err = scope.InterpretString("a()")
	if err != nil {
		t.Error(err)
	}
	expected := 5
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// TODO Packages

// TODO References
