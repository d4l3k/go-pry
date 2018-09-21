// +build js

package pry

import (
	"encoding/json"
	"go/ast"
	"go/types"
)

func (s *Scope) parseDir() (map[string]*ast.Package, error) {
	return defaultImporter.Dir, nil
}

func getImporter() types.Importer {
	return defaultImporter
}

var defaultImporter = &JSImporter{
	packages: map[string]*types.Package{},
	Dir:      map[string]*ast.Package{},
}

func InternalSetImports(raw string) {
	if err := json.Unmarshal([]byte(raw), defaultImporter); err != nil {
		panic(err)
	}
}
