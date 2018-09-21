package pry

import (
	"reflect"
	"regexp"
	"sort"
	"strings"
)

var suggestionsRegexp = regexp.MustCompile("[.0-9a-zA-Z]+$")

func (s *Scope) SuggestionsPry(line string, index int) ([]string, error) {
	text := line[:index]
	wip := suggestionsRegexp.FindString(text)

	if len(wip) == 0 {
		return nil, nil
	}

	var ok bool
	v := interface{}(s)
	parts := strings.Split(wip, ".")
	for _, k := range parts[:len(parts)-1] {
		v, ok = get(v, k)
		if !ok {
			return nil, nil
		}
	}

	partial := parts[len(parts)-1]

	var matchingKeys []string
	for _, key := range keys(v) {
		if strings.HasPrefix(key, partial) {
			matchingKeys = append(matchingKeys, key)
		}
	}

	sort.Strings(matchingKeys)

	return matchingKeys, nil
}

type keyser interface {
	Keys() []string
}

type getter interface {
	Get(string) (interface{}, bool)
}

func get(v interface{}, key string) (interface{}, bool) {
	if v == nil {
		return nil, false
	}

	g, ok := v.(getter)
	if ok {
		return g.Get(key)
	}

	val := reflect.ValueOf(v)

	typ := val.Type()
	switch typ.Kind() {
	case reflect.Ptr:
		return get(val.Elem().Addr(), key)

	case reflect.Struct:
		if _, ok := typ.FieldByName(key); !ok {
			return nil, false
		}
		return val.FieldByName(key).Interface(), true
	}

	return nil, false
}

func keys(v interface{}) []string {
	if v == nil {
		return nil
	}

	g, ok := v.(keyser)
	if ok {
		return g.Keys()
	}

	val := reflect.ValueOf(v)

	typ := val.Type()
	switch typ.Kind() {
	case reflect.Ptr:
		return keys(val.Elem().Addr())

	case reflect.Struct:
		var keys []string
		for i := 0; i < typ.NumField(); i++ {
			keys = append(keys, typ.Field(i).Name)
		}
		for i := 0; i < typ.NumMethod(); i++ {
			keys = append(keys, typ.Method(i).Name)
		}
		return keys
	}

	return nil
}
