package pry

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStringLiteral(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "\"Hello!\"")
	if err != nil {
		t.Error(err)
	}
	if out != "Hello!" {
		t.Error("Expected Hello!")
	}
}
func TestIntLiteral(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "-1234")
	if err != nil {
		t.Error(err)
	}
	if out != -1234 {
		t.Error("Expected -1234")
	}
}
func TestCharLiteral(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "'a'")
	if err != nil {
		t.Error(err)
	}
	if out != 'a' {
		t.Error(fmt.Sprintf("Expected 'a' got %#v.", out))
	}
}
func TestArrayLiteral(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "[]int{1,2,3,4}")
	if err != nil {
		t.Error(err)
	}
	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(expected, out) {
		t.Error(fmt.Sprintf("Expected %#v got %#v.", expected, out))
	}
}
func TestMapLiteral(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "map[string]int{\"duck\": 5,\n \"banana\": -123,\n}")
	if err != nil {
		t.Error(err)
	}
	expected := map[string]int{
		"duck":   5,
		"banana": -123,
	}
	if !reflect.DeepEqual(expected, out) {
		t.Error(fmt.Sprintf("Expected %#v got %#v.", expected, out))
	}
}
func TestBasicIdent(t *testing.T) {
	emptyScope := Scope{
		"a": 5,
	}
	out, err := InterpretString(emptyScope, "a")
	if err != nil {
		t.Error(err)
	}
	expected := 5
	if !reflect.DeepEqual(expected, out) {
		t.Error(fmt.Sprintf("Expected %#v got %#v.", expected, out))
	}
}
func TestMissingBasicIdent(t *testing.T) {
	emptyScope := Scope{}
	out, err := InterpretString(emptyScope, "a")
	if err == nil || out != nil {
		t.Error("Found non-existant ident.")
	}
}
func TestMapIdent(t *testing.T) {
	emptyScope := Scope{
		"a": Scope{
			"B": 10,
		},
	}
	out, err := InterpretString(emptyScope, "a[\"B\"]")
	if err != nil {
		t.Error(err)
	}
	expected := 10
	if !reflect.DeepEqual(expected, out) {
		t.Error(fmt.Sprintf("Expected %#v got %#v.", expected, out))
	}
}
func TestMissingMapIdent(t *testing.T) {
	emptyScope := Scope{
		"a": Scope{},
	}
	out, err := InterpretString(emptyScope, "a[\"b\"]")
	if err == nil || out != nil {
		t.Error("Found non-existant ident.")
	}
}
