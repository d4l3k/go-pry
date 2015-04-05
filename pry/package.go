package pry

// Package represents a Go package for use with pry
type Package struct {
	Name      string
	Functions map[string]interface{}
}
