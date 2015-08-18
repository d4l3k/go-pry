package pry

import "reflect"

// Type returns the reflect type of the passed object.
func Type(t interface{}) reflect.Type {
	return reflect.TypeOf(t)
}
