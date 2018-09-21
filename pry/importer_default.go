// +build !js

package pry

import (
	"go/ast"
	gcimporter "go/importer"
	"go/parser"
	"go/types"
	"path/filepath"
)

func getImporter() types.Importer {
	return gcimporter.Default()
}

func (s *Scope) parseDir() (map[string]*ast.Package, error) {
	return parser.ParseDir(s.fset, filepath.Dir(s.path), nil, 0)
}
