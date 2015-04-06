package pry

import (
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	a := 5
	out := Type(a)
	expected := reflect.TypeOf(a)
	if !reflect.DeepEqual(expected, out) {
		t.Errorf("Expected %#v got %#v.", expected, out)
	}
}
