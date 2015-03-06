package pry

import (
	"errors"
	"fmt"
	"reflect"
)

func Append(arr, elem interface{}) (interface{}, error) {
	if reflect.TypeOf(arr) != reflect.SliceOf(reflect.TypeOf(elem)) {
		return nil, errors.New(fmt.Sprintf("%T cannot append to %T.", elem, arr))
	}
	arrVal := reflect.ValueOf(arr)
	elemVal := reflect.ValueOf(elem)
	return reflect.Append(arrVal, elemVal).Interface(), nil
}
