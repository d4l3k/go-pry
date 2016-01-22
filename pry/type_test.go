package pry

import (
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	t.Parallel()

	a := 0
	out := Type(a)
	want := reflect.TypeOf(a)
	if !reflect.DeepEqual(want, out) {
		t.Errorf("Expected %#v got %#v.", want, out)
	}
}
