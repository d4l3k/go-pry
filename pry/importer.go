// +build !go1.5

package pry

import (
	"go/types"

	gcimporter "golang.org/x/tools/go/gcimporter15"
)

func getImporter() types.ImporterFrom {
	return importer{
		impFn:    gcimporter.Import,
		packages: make(map[string]*types.Package),
	}
}

// importer implements go/types.Importer.
// It also implements go/types.ImporterFrom, which was new in Go 1.6,
// so vendoring will work.
type importer struct {
	impFn    func(packages map[string]*types.Package, path, srcDir string) (*types.Package, error)
	packages map[string]*types.Package
}

func (i importer) Import(path string) (*types.Package, error) {
	return i.impFn(i.packages, path, "")
}

func (i importer) ImportFrom(path, srcDir string, mode types.ImportMode) (*types.Package, error) {
	return i.impFn(i.packages, path, srcDir)
}
