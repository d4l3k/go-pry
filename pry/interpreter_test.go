package pry

import (
	"fmt"
	"reflect"
	"testing"
)

// Literals
func TestStringLiteral(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("-1234")
	if err != nil {
		t.Error(err)
	}
	if out != -1234 {
		t.Error("Expected -1234")
	}
}
func TestHexIntLiteral(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("0xC123")
	if err != nil {
		t.Error(err)
	}
	expected := 0xC123
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestOctalIntLiteral(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("03272")
	if err != nil {
		t.Error(err)
	}
	expected := 03272
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestCharLiteral(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

func TestFixedArrayLiteral(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString(`[4]int{1,2,3,4}`)
	if err != nil {
		t.Error(err)
	}
	expected := [4]int{1, 2, 3, 4}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestFixedArray(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString(`
		var a [3]int
		a[2]
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 0
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestFixedArraySet(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString(`
		var a [3]int
		b := &a
		a[2] = 1
		b[2]
	`)
	if err != nil {
		t.Errorf("%+v", err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestArraySet(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString(`
		a := []int{1,2,3,4}
		a[2] = 1
		a[2]
	`)
	if err != nil {
		t.Errorf("%+v", err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestMapLiteral(t *testing.T) {
	t.Parallel()

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

func TestMapLiteralInterface(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("map[string]interface{}{\"duck\": 5,\n \"banana\": -123,\n}")
	if err != nil {
		t.Error(err)
	}
	expected := map[string]interface{}{
		"duck":   5,
		"banana": -123,
	}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestTypeCast(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("a")
	if err == nil || out != nil {
		t.Error("Found non-existant ident.")
	}
}
func TestMapIdent(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	scope := NewScope()
	scope.Set("a", []int{1})

	out, err := scope.InterpretString("a[1]")
	if err == nil || out != nil {
		t.Error("Should have thrown out of range error")
	}
}

func TestSlice(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestMathShifting(t *testing.T) {
	t.Parallel()

	types := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr",
	}
	cases := []struct {
		l      int
		op     string
		r, out int
	}{
		{3, "%", 2, 1},
		{7, "&", 2, 2},
		{6, "|", 2, 6},
		{6, "^", 2, 4},
		{2, "<<", 2, 8},
		{8, ">>", 2, 2},
		{6, "&^", 4, 2},
	}
	scope := NewScope()
	for _, typ := range types {
		for _, td := range cases {
			query := fmt.Sprintf("%s(%d) %s %s(%d)", typ, td.l, td.op, typ, td.r)
			outI, err := scope.InterpretString(query)
			if err != nil {
				t.Error(err)
			}
			out := interfaceToInt(outI)
			if !reflect.DeepEqual(td.out, out) {
				t.Errorf("Expected %#v = %#v got %#v.", query, td.out, out)
			}
		}
	}
}

func TestMathBasic(t *testing.T) {
	t.Parallel()

	types := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"uintptr",
		"float32", "float64",
	}
	cases := []struct {
		l      int
		op     string
		r, out int
	}{
		{3, "+", 2, 5},
		{3, "-", 2, 1},
		{3, "*", 2, 6},
		{4, "/", 2, 2},
		{4, ">", 3, 1},
		{3, ">", 4, -1},
		{4, ">=", 3, 1},
		{3, ">=", 4, -1},
		{4, "<", 3, -1},
		{3, "<", 4, 1},
		{4, "<=", 3, -1},
		{3, "<=", 4, 1},
		{3, "==", 3, 1},
		{3, "==", 4, -1},
		{3, "!=", 3, -1},
		{3, "!=", 4, 1},
	}
	scope := NewScope()
	for _, typ := range types {
		for _, td := range cases {
			query := fmt.Sprintf("%s(%d) %s %s(%d)", typ, td.l, td.op, typ, td.r)
			outI, err := scope.InterpretString(query)
			if err != nil {
				t.Error(err)
			}
			out := interfaceToInt(outI)
			if !reflect.DeepEqual(td.out, out) {
				t.Errorf("Expected %#v = %#v got %#v.", query, td.out, out)
			}
		}
	}
}

func TestBoolConds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		l      bool
		op     string
		r, out bool
	}{
		{true, "&&", true, true},
		{true, "&&", false, false},
		{false, "&&", true, false},
		{false, "&&", false, false},
		{true, "||", true, true},
		{true, "||", false, true},
		{false, "||", true, true},
		{false, "||", false, false},
	}
	scope := NewScope()
	for _, td := range cases {
		query := fmt.Sprintf("%#v %s %#v", td.l, td.op, td.r)
		outI, err := scope.InterpretString(query)
		if err != nil {
			t.Error(err)
		}
		out := outI.(bool)
		if !reflect.DeepEqual(td.out, out) {
			t.Errorf("Expected %#v = %#v got %#v.", query, td.out, out)
		}
	}
}

func interfaceToInt(i interface{}) int {
	switch v := i.(type) {
	case int:
		return int(v)
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case uintptr:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return -1
	}
	return 0
}

func TestStringConcat(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	scope.Set("a", 5)

	out, err := scope.InterpretString("\"hello\" + \"foo\"")
	if err != nil {
		t.Error(err)
	}
	expected := "hellofoo"
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestParens(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestMakeChanInterface(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("make(chan interface{}, 10)")
	if err != nil {
		t.Error(err)
	}
	expected := make(chan interface{}, 10)
	if reflect.TypeOf(expected) != reflect.TypeOf(out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestMakeUnknown(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	out, err := scope.InterpretString("make(int)")
	if err == nil || out != nil {
		t.Error("Should have thrown error.")
	}
}

func TestAppend(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	scope.Set("a", []int{1})

	_, err := scope.InterpretString("a = append(a, 2, 3)")
	if err != nil {
		t.Error(err)
	}
	expected := []int{1, 2, 3}
	outV, found := scope.Get("a")
	if !found {
		t.Errorf("failed to find \"a\"")
	}
	out := outV.([]int)
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestDeclareAssignVar(t *testing.T) {
	t.Parallel()

	scope := NewScope()
	scope.Set("a", []int{1})

	out, err := scope.InterpretString("var a, b int = 2, 3")
	if err != nil {
		t.Error(err)
	}
	testData := []struct {
		v    string
		want int
	}{
		{"a", 2},
		{"b", 3},
	}
	for _, td := range testData {
		out, _ = scope.Get(td.v)
		if !reflect.DeepEqual(td.want, out) {
			t.Errorf("Expected %#v got %#v.", td.want, out)
		}
	}
}

func TestDeclareAssign(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

// Channels

func TestChannel(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString("a := make(chan int, 10); a <- 1; a <- 2; []int{<-a, <-a}")
	if err != nil {
		t.Error(err)
	}
	expected := []int{1, 2}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestChannelSendFail(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	_, out := scope.InterpretString("a := make(chan int); a <- 1")
	expected := ErrChanSendFailed
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected err %#v got %#v.", expected, out)
	}
}

func TestChannelRecvFail(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	_, out := scope.InterpretString("a := make(chan int); close(a); <-a")
	expected := ErrChanRecvFailed
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected err %#v got %#v.", expected, out)
	}
}

// Control structures

func TestFor(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString("a := 1; for i := 0; i < 5; i++ { a++ }; a")
	if err != nil {
		t.Error(err)
	}
	expected := 6
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}

	out, err = scope.InterpretString("a := 1; for i := 5; i > 0; i-- { a++ }; a")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestForBreak(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	_, err := scope.InterpretString("for { break }")
	if err != nil {
		t.Error(err)
	}
}

func TestForContinue(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	for i:=0; i < 1; i++ {
		a = 1
		continue
		a = 2
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestForRangeArray(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString("a := 1; for i, c := range []int{1,2,3} {a=a+i+c}; a")
	if err != nil {
		t.Error(err)
	}
	expected := 1 + 0 + 1 + 2 + 1 + 2 + 3
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestForRangeMap(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString("a := 1; for i, c := range map[int]int{0: 1, 1: 2, 2: 3} {a=a+i+c}; a")
	if err != nil {
		t.Error(err)
	}
	expected := 1 + 0 + 1 + 2 + 1 + 2 + 3
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSelectDefault(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	c := make(chan int)
	select {
	case b := <-c:
		a = b
	default:
		a = 1
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSelect(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	c := make(chan int, 10)
	c <- 2
	select {
	case b := <-c:
		a = b
	default:
		a = 1
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSelectMultiCase(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	c := make(chan int, 10)
	e := make(chan int, 10)
	c <- 2
	a := 0
	select {
	case d := <-e:
		a = d
	case b := <-c:
		a = b
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitch(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 10
	out := 0
	switch a {
	case 10:
		out = 1
	default:
		out = 2
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitchDefault(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	out := 0
	switch a {
	case 10:
		out = 1
	default:
		out = 2
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitchBool(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	out := 0
	switch {
	case true:
		out = 1
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitchType(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	out := 0
	var t interface{}
	t = 10
	switch t.(type){
	case int:
		out = 1
	case bool:
		out = 2
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitchTypeUse(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	out := 0
	var t interface{}
	t = 10
	switch t := t.(type){
	case int:
		out = t
	case bool:
		out = 2
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 10
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSwitchNone(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	out := 0
	switch {
	case false:
		out = 1
	}
	out
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 0
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestIf(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	if true {
		a = 1
	} else {
		a = 2
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 1
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestIfElse(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	if false {
		a = 1
	} else {
		a = 2
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestIfIfElse(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 0
	if false {
		a = 1
	} else if true {
		a = 2
	}
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestFunctionArgs(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	f := func(b, c int) {
		return b + c
	}
	f(10, 5)
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 15
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestDefer(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a :=  0
	f := func() {
		defer func() {
			a = 2
		}()
		defer func() {
			a = 3
		}()
		a = 1
	}
	f()
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 2
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestStringAppend(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := "foo"
	a += "bar"
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := "foobar"
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestIntMod(t *testing.T) {
	t.Parallel()

	scope := NewScope()

	out, err := scope.InterpretString(`
	a := 10
	a += 6
	a -= 1
	a /= 3
	a *= 4
	a
	`)
	if err != nil {
		t.Error(err)
	}
	expected := 20
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

// TODO Packages

// TODO References
