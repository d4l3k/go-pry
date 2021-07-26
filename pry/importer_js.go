// +build js

package pry

import (
	"encoding/json"
	"go/ast"
	"go/types"
)

func (s *Scope) parseDir() (map[string]*ast.File, error) {
	files := map[string]*ast.File{}
	for _, p := range defaultImporter.Dir {
		for name, file := range p.Files {
			files[name] = file
		}
	}
	return files, nil
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
