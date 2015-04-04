package pry

import (
	"reflect"
)

// Package represents a Go package for use with pry
type Package struct {
	Name      string
	Functions map[string]interface{}
}

// Process converts items in p.Functions into the corresponding reflect.Type if not a function.
func (p *Package) Process() {
	for k, v := range p.Functions {
		if reflect.TypeOf(v).Kind() != reflect.Func {
			p.Functions[k] = reflect.TypeOf(v)
		}
	}
}
