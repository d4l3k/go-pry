package pry

import (
	"reflect"
	"sort"
)

// Suggestions returns auto complete suggestions for the query.
func (scope *Scope) Suggestions(query string) []string {
	terms := []string{}
	if len(query) == 0 {
		terms = append(terms, scope.Keys()...)
	} else if query[len(query)-1] == '.' {
		val, present := scope.Get(query[:len(query)-1])
		if present {
			pkg, isPackage := val.(Package)
			if isPackage {
				val = pkg.Functions
				for k, v := range pkg.Functions {
					if reflect.TypeOf(v).Kind() == reflect.Func {
						k += "("
					}
					terms = append(terms, k)
				}
			}
			typeOf := reflect.TypeOf(val)
			methods := make([]string, typeOf.NumMethod())
			for i := range methods {
				methods[i] = typeOf.Method(i).Name + "("
			}
			terms = append(terms, methods...)

			if typeOf.Kind() == reflect.Struct {
				fields := make([]string, typeOf.NumField())
				for i := range fields {
					fields[i] = typeOf.Field(i).Name
				}
				terms = append(terms, fields...)
			}
		}
	}
	sort.Sort(sort.StringSlice(terms))
	return terms
}
