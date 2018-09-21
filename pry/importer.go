package pry

import (
	"go/ast"
	"go/types"

	"github.com/pkg/errors"
)

// JSImporter contains all the information needed to implement a types.Importer
// in a javascript environment.
type JSImporter struct {
	packages map[string]*types.Package
	Dir      map[string]*ast.Package
}

func (i *JSImporter) Import(path string) (*types.Package, error) {
	p, ok := i.packages[path]
	if !ok {
		return nil, errors.Errorf("package %q not found", path)
	}
	return p, nil
}
