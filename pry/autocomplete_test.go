package pry

import (
	"reflect"
	"testing"
)

func TestSuggestionsNone(t *testing.T) {
	scope := NewScope()
	out := scope.Suggestions("")
	expected := []string{}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
func TestSuggestionsSome(t *testing.T) {
	scope := NewScope()
	scope.Set("b", 6)
	scope.Set("c", 6)
	scope.Set("a", 5)
	out := scope.Suggestions("")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSuggestionsStruct(t *testing.T) {
	scope := NewScope()
	scope.Set("a", testStruct{})
	out := scope.Suggestions("a.")
	expected := []string{"A", "B(", "C", "D"}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSuggestionsLit(t *testing.T) {
	scope := NewScope()
	out := scope.Suggestions("\"test\"")
	expected := []string{}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}

func TestSuggestionsPackage(t *testing.T) {
	scope := NewScope()
	scope.Set("test", Package{
		Name: "test",
		Functions: map[string]interface{}{
			"A": 5,
			"B": func() {},
		},
	})
	out := scope.Suggestions("test.")
	expected := []string{"A", "B("}
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
