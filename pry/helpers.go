package pry

import (
	"errors"
	"fmt"
	"reflect"
)

// Append is a runtime replacement for the append function
func Append(arr interface{}, elems ...interface{}) (interface{}, error) {
	arrVal := reflect.ValueOf(arr)
	valArr := make([]reflect.Value, len(elems))
	for i, elem := range elems {
		if reflect.TypeOf(arr) != reflect.SliceOf(reflect.TypeOf(elem)) {
			return nil, fmt.Errorf("%T cannot append to %T", elem, arr)
		}
		valArr[i] = reflect.ValueOf(elem)
	}
	return reflect.Append(arrVal, valArr...).Interface(), nil
}

// Make is a runtime replacement for the make function
func Make(t interface{}, args ...interface{}) (interface{}, error) {
	typ, isType := t.(reflect.Type)
	if !isType {
		return nil, fmt.Errorf("invalid type %#v", t)
	}
	switch typ.Kind() {
	case reflect.Slice:
		if len(args) < 1 || len(args) > 2 {
			return nil, errors.New("invalid number of arguments. Missing len or extra?")
		}
		length, isInt := args[0].(int)
		if !isInt {
			return nil, errors.New("len is not int")
		}
		capacity := length
		if len(args) == 2 {
			capacity, isInt = args[0].(int)
			if !isInt {
				return nil, errors.New("len is not int")
			}
		}
		slice := reflect.MakeSlice(typ, length, capacity)
		return slice.Interface(), nil

	case reflect.Chan:
		if len(args) > 1 {
			fmt.Printf("CHAN ARGS %#v", args)
			return nil, errors.New("too many arguments")
		}
		size := 0
		if len(args) == 1 {
			var isInt bool
			size, isInt = args[0].(int)
			if !isInt {
				return nil, errors.New("size is not int")
			}
		}
		buffer := reflect.MakeChan(typ, size)
		return buffer.Interface(), nil

	default:
		return nil, fmt.Errorf("unknown kind type %T", t)
	}
}

// Len is a runtime replacement for the len function
func Len(t interface{}) (interface{}, error) {
	return reflect.ValueOf(t).Len(), nil
}
