// +build !js

package pry

import (
	"go/ast"
	"go/types"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

type packagesImporter struct {
}

func (i packagesImporter) Import(path string) (*types.Package, error) {
	return i.ImportFrom(path, "", 0)
}
func (packagesImporter) ImportFrom(path, dir string, mode types.ImportMode) (*types.Package, error) {
	conf := packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes,
		Dir:  dir,
	}
	pkgs, err := packages.Load(&conf, path)
	if err != nil {
		return nil, errors.Wrapf(err, "importing %q", path)
	}
	pkg := pkgs[0]
	return pkg.Types, nil
}

func getImporter() types.ImporterFrom {
	return packagesImporter{}
}

func (s *Scope) parseDir() (map[string]*ast.File, error) {
	conf := packages.Config{
		Fset: s.fset,
		Mode: packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Dir:  filepath.Dir(s.path),
	}
	pkgs, err := packages.Load(&conf, ".")
	if err != nil {
		return nil, errors.Wrapf(err, "parsing dir")
	}
	pkg := pkgs[0]
	files := map[string]*ast.File{}
	for i, name := range pkg.CompiledGoFiles {
		files[name] = pkg.Syntax[i]
	}
	return files, nil
}
