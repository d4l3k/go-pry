package pry

import (
	"reflect"
)

type Package struct {
	Name      string
	Functions map[string]interface{}
}

func (p *Package) Process() {
	for k, v := range p.Functions {
		if reflect.TypeOf(v).Kind() != reflect.Func {
			p.Functions[k] = reflect.TypeOf(v)
		}
	}
}
