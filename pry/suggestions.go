package pry

import (
	"log"
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
	g, ok := v.(getter)
	if ok {
		return g.Get(key)
	}

	switch v := v.(type) {
	default:
		log.Fatalf("can't handle get for %T: %+v", v, v)
		return nil, false
	}
}

func keys(v interface{}) []string {
	g, ok := v.(keyser)
	if ok {
		return g.Keys()
	}

	switch v := v.(type) {
	default:
		log.Fatalf("can't handle keys for %T: %+v", v, v)
		return nil
	}
}
