package generate

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/d4l3k/go-pry/pry"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

type Generator struct {
	contexts []pryContext
	debug    bool
	Config   packages.Config
}

func NewGenerator(debug bool) *Generator {
	return &Generator{
		debug: debug,
		Config: packages.Config{
			Mode: packages.NeedName | packages.NeedSyntax,
		},
	}
}

// Debug prints debug statements if debug is true.
func (g Generator) Debug(templ string, k ...interface{}) {
	if g.debug {
		log.Printf(templ, k...)
	}
}

// ExecuteGoCmd runs the 'go' command with certain parameters.
func (g *Generator) ExecuteGoCmd(ctx context.Context, args []string, env []string) error {
	binary, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// InjectPry walks the scope and replaces pry.Pry with pry.Apply(pry.Scope{...}).
func (g *Generator) InjectPry(filePath string) (string, error) {
	g.Debug("Prying into %s\n", filePath)
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", nil
	}

	g.contexts = make([]pryContext, 0)

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return "", err
	}

	g.Config.Dir = filepath.Dir(filePath)

	packagePairs := []string{}
	for _, imp := range f.Imports {
		importStr := imp.Path.Value[1 : len(imp.Path.Value)-1]
		if importStr != "../pry" {
			pkgs, err := packages.Load(&g.Config, importStr)
			if err != nil {
				return "", err
			}
			pkg := pkgs[0]
			importName := pkg.Name
			if imp.Name != nil {
				importName = imp.Name.Name
			}
			pair := "\"" + importName + "\": pry.Package{Name: \"" + pkg.Name + "\", Functions: map[string]interface{}{"
			added := make(map[string]bool)
			exports, err := g.GetExports(importName, pkg.Syntax, added)
			if err != nil {
				return "", err
			}
			pair += exports
			pair += "}}, "
			packagePairs = append(packagePairs, pair)
		}
	}

	var funcs []*ast.FuncDecl
	var vars []string

	// Print the imports from the file's AST.
	for k, v := range f.Scope.Objects {
		switch decl := v.Decl.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, decl)
		case *ast.ValueSpec:
			vars = append(vars, k)
		}
	}

	for _, f := range funcs {
		vars = append(vars, f.Name.Name)
		if f.Recv != nil {
			vars = g.extractFields(vars, f.Recv.List)
		}
		if f.Type != nil {
			if f.Type.Params != nil {
				vars = g.extractFields(vars, f.Type.Params.List)
			}
			if f.Type.Results != nil {
				vars = g.extractFields(vars, f.Type.Results.List)
			}
		}
		vars = g.extractVariables(vars, f.Body.List)
	}

	fileTextBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", nil
	}

	fileText := (string)(fileTextBytes)

	offset := 0

	if len(g.contexts) == 0 {
		return "", nil
	}

	g.Debug(" :: Found %d pry statements.\n", len(g.contexts))

	for _, context := range g.contexts {
		filteredVars := filterVars(context.Vars)
		obj := "&pry.Scope{Vals:map[string]interface{}{ "
		for _, v := range filteredVars {
			obj += "\"" + v + "\": " + v + ", "
		}
		obj += strings.Join(packagePairs, "")
		obj += "}}"
		text := "pry.Apply(" + obj + ")"
		fileText = fileText[0:context.Start+offset] + text + fileText[context.End+offset:]
		offset = len(text) - (context.End - context.Start)
	}

	newPath := filepath.Dir(filePath) + "/." + filepath.Base(filePath) + "pry"

	err = os.Rename(filePath, newPath)
	if err != nil {
		return "", err
	}
	ioutil.WriteFile(filePath, ([]byte)(fileText), 0644)
	return filePath, nil
}

// GetExports returns a string of gocode that represents the exports (constants/functions) of an ast.Package.
func (g *Generator) GetExports(importName string, files []*ast.File, added map[string]bool) (string, error) {
	vars := ""
	for _, file := range files {
		// Print the imports from the file's AST.
		scope := pry.NewScope()
		for k, obj := range file.Scope.Objects {
			if added[k] {
				continue
			}
			added[k] = true
			firstLetter := k[0:1]
			if firstLetter == strings.ToUpper(firstLetter) && firstLetter != "_" {

				isType := false

				switch stmt := obj.Decl.(type) {
				/*
					case *ast.ValueSpec:
						if len(stmt.Values) > 0 {
							out, err := pry.InterpretExpr(scope, stmt.Values[0])
							if err != nil {
								fmt.Println("ERR", err)
								//continue
							} else {
								scope[obj.Name] = out
							}
						}
				*/
				case *ast.TypeSpec:
					switch typ := stmt.Type.(type) {
					case *ast.StructType:
						isType = true
						scope.Set(obj.Name, typ)

					default:
						out, err := scope.Interpret(stmt.Type)
						if err != nil {
							g.Debug("TypeSpec ERR %s\n", err.Error())
							//continue
						} else {
							scope.Set(obj.Name, out)
							isType = true
						}
					}
				}

				if obj.Kind != ast.Typ || isType {
					path := importName + "." + k
					vars += "\"" + k + "\": "
					if isType {
						out, _ := scope.Get(obj.Name)
						switch v := out.(type) {
						case reflect.Type:
							zero := reflect.Zero(v).Interface()
							val := fmt.Sprintf("%#v", zero)
							if zero == nil {
								val = "nil"
							}
							vars += fmt.Sprintf("pry.Type(%s(%s))", path, val)
						case *ast.StructType:
							vars += fmt.Sprintf("pry.Type(%s{})", path)
						default:
							log.Fatalf("got unknown type: %T %+v", out, out)
						}

						// TODO Fix hack for very large constants
					} else if path == "math.MaxUint64" || path == "crc64.ISO" || path == "crc64.ECMA" {
						vars += fmt.Sprintf("uint64(%s)", path)
					} else {
						vars += path
					}
					vars += ","
					if g.debug {
						vars += "\n"
					}
				}
			}
		}
	}
	return vars, nil
}

// GenerateFile generates a injected file.
func (g *Generator) GenerateFile(imports []string, extraStatements, path string) error {
	file := "package main\nimport (\n\t\"github.com/d4l3k/go-pry/pry\"\n\n"
	for _, imp := range imports {
		if len(imp) == 0 {
			continue
		}
		file += fmt.Sprintf("\t%#v\n", imp)
	}
	file += ")\nfunc main() {\n\t" + extraStatements + "\n\tpry.Pry()\n}\n"

	if err := ioutil.WriteFile(path, []byte(file), 0644); err != nil {
		return err
	}

	_, err := g.InjectPry(path)
	return err
}

// GenerateAndExecuteFile generates and executes a temp file with the given imports
func (g *Generator) GenerateAndExecuteFile(ctx context.Context, imports []string, extraStatements string) error {
	dir, err := ioutil.TempDir("", "pry")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			log.Fatal(err)
		}
	}()
	newPath := dir + "/main.go"

	if err := g.GenerateFile(imports, extraStatements, newPath); err != nil {
		return err
	}

	if err := g.ExecuteGoCmd(ctx, []string{"run", newPath}, nil); err != nil {
		return err
	}
	return nil
}

// RevertPry reverts the changes made by InjectPry.
func (g *Generator) RevertPry(modifiedFiles []string) error {
	fmt.Println("Reverting files")
	for _, file := range modifiedFiles {
		newPath := filepath.Dir(file) + "/." + filepath.Base(file) + "pry"
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return errors.Errorf("no such file or directory: %s", newPath)
		}

		err := os.Remove(file)
		if err != nil {
			return err
		}
		err = os.Rename(newPath, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func filterVars(vars []string) (fVars []string) {
	for _, v := range vars {
		if v != "_" {
			fVars = append(fVars, v)
		}
	}
	return
}

func (g *Generator) extractVariables(vars []string, l []ast.Stmt) []string {
	for _, s := range l {
		vars = g.handleStatement(vars, s)
	}
	return vars
}

func (g *Generator) extractFields(vars []string, l []*ast.Field) []string {
	for _, s := range l {
		vars = g.handleIdents(vars, s.Names)
	}
	return vars
}

func (g *Generator) handleStatement(vars []string, s ast.Stmt) []string {
	switch stmt := s.(type) {
	case *ast.ExprStmt:
		vars = g.handleExpr(vars, stmt.X)
	case *ast.AssignStmt:
		lhsStatements := (*stmt).Lhs
		for _, v := range lhsStatements {
			vars = g.handleExpr(vars, v)
		}
	case *ast.GoStmt:
		g.handleExpr(vars, stmt.Call)
	case *ast.IfStmt:
		g.handleIfStmt(vars, stmt)
	case *ast.DeclStmt:
		decl := stmt.Decl.(*ast.GenDecl)
		if decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				valSpec := spec.(*ast.ValueSpec)
				vars = g.handleIdents(vars, valSpec.Names)
			}
		}
	case *ast.BlockStmt:
		vars = g.handleBlockStmt(vars, stmt)
	case *ast.RangeStmt:
		g.handleRangeStmt(vars, stmt)
	case *ast.ForStmt:
		vars = g.handleForStmt(vars, stmt)
	default:
		g.Debug("Unknown %T\n", stmt)
	}
	return vars
}

func (g *Generator) handleIfStmt(vars []string, stmt *ast.IfStmt) []string {
	vars = g.handleStatement(vars, stmt.Init)
	vars = g.handleStatement(vars, stmt.Body)
	return vars
}

func (g *Generator) handleRangeStmt(vars []string, stmt *ast.RangeStmt) []string {
	vars = g.handleExpr(vars, stmt.Key)
	vars = g.handleExpr(vars, stmt.Value)
	vars = g.handleStatement(vars, stmt.Body)
	return vars
}

func (g *Generator) handleForStmt(vars []string, stmt *ast.ForStmt) []string {
	vars = g.handleStatement(vars, stmt.Init)
	vars = g.handleStatement(vars, stmt.Body)
	return vars
}

func (g *Generator) handleBlockStmt(vars []string, stmt *ast.BlockStmt) []string {
	vars = g.extractVariables(vars, stmt.List)
	return vars
}

func (g *Generator) handleIdents(vars []string, idents []*ast.Ident) []string {
	for _, i := range idents {
		vars = append(vars, i.Name)
	}
	return vars
}

func (g *Generator) handleExpr(vars []string, v ast.Expr) []string {
	switch expr := v.(type) {
	case *ast.Ident:
		varMap := make(map[string]bool)
		for _, v := range vars {
			varMap[v] = true
		}
		if !varMap[expr.Name] {
			vars = append(vars, expr.Name)
		}
	case *ast.CallExpr:
		switch fun := expr.Fun.(type) {
		case *ast.SelectorExpr:
			funcName := fun.Sel.Name
			if funcName == "Pry" || funcName == "Apply" {
				g.contexts = append(g.contexts, pryContext{(int)(expr.Pos() - 1), (int)(expr.End() - 1), vars})
			}
			//handleExpr(vars, fun.X)
		case *ast.FuncLit:
			g.handleExpr(vars, fun)
		default:
			g.Debug("Unknown function type %T\n", fun)
		}
		for _, arg := range expr.Args {
			g.handleExpr(vars, arg)
		}
	case *ast.FuncLit:
		if expr.Type.Params != nil {
			for _, param := range expr.Type.Params.List {
				for _, name := range param.Names {
					vars = g.handleExpr(vars, name)
				}
			}
		}
		if expr.Type.Results != nil {
			for _, param := range expr.Type.Results.List {
				for _, name := range param.Names {
					vars = g.handleExpr(vars, name)
				}
			}
		}
		g.handleStatement(vars, expr.Body)
	default:
		g.Debug("Unknown %T\n", expr)
	}
	return vars
}

type pryContext struct {
	Start, End int
	Vars       []string
}
