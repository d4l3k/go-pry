package pry

// Package represents a Go package for use with pry
type Package struct {
	Name      string
	Functions map[string]interface{}
}

func (p Package) Keys() []string {
	var keys []string
	for k := range p.Functions {
		keys = append(keys, k)
	}
	return keys
}

func (p Package) Get(key string) (interface{}, bool) {
	v, ok := p.Functions[key]
	return v, ok
}
