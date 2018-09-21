// +build js

package pry

import (
	"encoding/json"
	"go/ast"
	"go/types"

	"github.com/pkg/errors"
)

func (s *Scope) parseDir() (map[string]*ast.Package, error) {
	return defaultImporter.Dir, nil
}

var defaultImporter = &JSImporter{
	Packages: map[string]*types.Package{},
	Dir:      map[string]*ast.Package{},
}

type JSImporter struct {
	Packages map[string]*types.Package
	Dir      map[string]*ast.Package
}

func (i *JSImporter) Import(path string) (*types.Package, error) {
	p, ok := i.Packages[path]
	if !ok {
		return nil, errors.Errorf("package %q not found", path)
	}
	return p, nil
}

func getImporter() types.Importer {
	return defaultImporter
}

func InternalSetImports(raw string) {
	if err := json.Unmarshal([]byte(raw), defaultImporter); err != nil {
		panic(err)
	}
}
