package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/d4l3k/go-pry/pry"
)

var contexts []PryContext

func ExecuteGoCmd(args []string) {
	binary, lookErr := exec.LookPath("go")
	if lookErr != nil {
		panic(lookErr)
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func InjectPry(filePath string) (string, error) {
	fmt.Println("Prying into ", filePath)

	contexts = make([]PryContext, 0)

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return "", err
	}

	fmt.Println(f.Imports)

	packagePairs := []string{}
	for _, imp := range f.Imports {
		importStr := imp.Path.Value[1 : len(imp.Path.Value)-1]
		if importStr != "../pry" {
			dir := filepath.Dir(filePath)
			pkg, err := build.Import(importStr, dir, build.AllowBinary)
			if err != nil {
				panic(err)
			}
			pkgAst, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
			if err != nil {
				panic(err)
			}
			pair := "\"" + pkg.Name + "\": pry.Package{Name: \"" + pkg.Name + "\", Functions: map[string]interface{}{"
			for _, nPkg := range pkgAst {
				pair += GetExports(nPkg)
			}
			pair += "}}, "
			packagePairs = append(packagePairs, pair)
		}
	}

	funcs := make([]*ast.FuncDecl, 0)
	vars := make([]string, 0)

	// Print the imports from the file's AST.
	for k, v := range f.Scope.Objects {
		switch decl := v.Decl.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, decl)
		case *ast.ValueSpec:
			fmt.Println(k, decl)
			vars = append(vars, k)
		}
	}

	for _, f := range funcs {
		vars = append(vars, f.Name.Name)
		ExtractVariables(vars, f.Body.List)
	}

	fileTextBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	fileText := (string)(fileTextBytes)

	offset := 0

	if len(contexts) == 0 {
		return "", nil
	}

	fmt.Println(" :: Found", len(contexts), "pry statements.")

	for _, context := range contexts {
		vars := FilterVars(context.Vars)
		obj := "map[string]interface{}{ "
		for _, v := range vars {
			obj += "\"" + v + "\": " + v + ", "
		}
		obj += strings.Join(packagePairs, "")
		obj += "}"
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

func main() {
	cmdArgs := os.Args[1:]
	if len(cmdArgs) == 0 {
		ExecuteGoCmd([]string{})
		fmt.Println("----")
		fmt.Println("go-pry is a wrapper around the go command.")
		fmt.Println("You can execute go commands as normal and go-pry will take care of generating the pry code.")
		fmt.Println("You can also use 'go-pry revert' to cleanup go-pry generated files. They should automatically be removed.")
		return
	}

	goDirs := []string{}
	for _, arg := range cmdArgs {
		if strings.HasSuffix(arg, ".go") {
			goDirs = append(goDirs, filepath.Dir(arg))
		}
	}
	if len(goDirs) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		goDirs = []string{dir}
	}

	processedFiles := []string{}
	modifiedFiles := []string{}

	if cmdArgs[0] == "revert" {
		fmt.Println("REVERTING PRY")
		for _, dir := range goDirs {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".gopry") {
					processed := false
					for _, file := range processedFiles {
						if file == path {
							processed = true
						}
					}
					if !processed {
						base := filepath.Base(path)
						newPath := filepath.Dir(path) + "/" + base[1:len(base)-3]
						modifiedFiles = append(modifiedFiles, newPath)
					}
				}
				return nil
			})
		}
		RevertPry(modifiedFiles)
		return
	}

	for _, dir := range goDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".go") {
				processed := false
				for _, file := range processedFiles {
					if file == path {
						processed = true
					}
				}
				if !processed {
					file, err := InjectPry(path)
					if err != nil {
						panic(err)
					}
					if file != "" {
						modifiedFiles = append(modifiedFiles, path)
					}
				}
			}
			return nil
		})
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
		}
	}()

	ExecuteGoCmd(cmdArgs)

	RevertPry(modifiedFiles)
}

func RevertPry(modifiedFiles []string) {
	fmt.Println("Reverting files")
	for _, file := range modifiedFiles {
		newPath := filepath.Dir(file) + "/." + filepath.Base(file) + "pry"
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			fmt.Println(" :: no such file or directory:", newPath)
			continue
		}

		err := os.Remove(file)
		if err != nil {
			fmt.Println(err)
		}
		err = os.Rename(newPath, file)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func GetExports(pkg *ast.Package) string {
	vars := ""
	for name, file := range pkg.Files {
		if !strings.HasSuffix(name, "_test.go") {
			// Print the imports from the file's AST.
			for k, obj := range file.Scope.Objects {
				firstLetter := k[0:1]
				if firstLetter == strings.ToUpper(firstLetter) && firstLetter != "_" {
					vars += "\"" + k + "\": " + pkg.Name + "." + k
					switch stmt := obj.Decl.(type) {
					case *ast.ValueSpec:
						fmt.Printf("FILE %#v %#v\n", obj.Name, stmt.Values)
						if len(stmt.Values) > 0 {
							out, err := pry.InterpretExpr(pry.Scope{}, stmt.Values[0])
							if err != nil {
								fmt.Println("ERR", err)
								//continue
							}
							fmt.Println(out)
						}
					}
					if obj.Kind == ast.Typ {
						vars += "{}"
					}
					vars += ","
				}
			}
		}
	}
	return vars
}

func FilterVars(vars []string) []string {
	fVars := make([]string, 0)
	for _, v := range vars {
		if v != "_" {
			fVars = append(fVars, v)
		}
	}
	return fVars
}

func ExtractVariables(vars []string, l []ast.Stmt) []string {
	for _, s := range l {
		vars = HandleStatement(vars, s)
	}
	return vars
}

func HandleStatement(vars []string, s ast.Stmt) []string {
	switch stmt := s.(type) {
	case *ast.ExprStmt:
		vars = HandleExpr(vars, stmt.X)
	case *ast.AssignStmt:
		lhsStatements := (*stmt).Lhs
		for _, v := range lhsStatements {
			vars = HandleExpr(vars, v)
		}
	case *ast.IfStmt:
		HandleIfStmt(vars, stmt)
	case *ast.DeclStmt:
		decl := stmt.Decl.(*ast.GenDecl)
		if decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				valSpec := spec.(*ast.ValueSpec)
				vars = HandleIdents(vars, valSpec.Names)
			}
		}
	case *ast.BlockStmt:
		vars = HandleBlockStmt(vars, stmt)
	case *ast.RangeStmt:
		HandleRangeStmt(vars, stmt)
	case *ast.ForStmt:
		vars = HandleForStmt(vars, stmt)
	default:
		fmt.Printf("Unknown %T\n", stmt)
	}
	return vars
}

func HandleIfStmt(vars []string, stmt *ast.IfStmt) []string {
	vars = HandleStatement(vars, stmt.Init)
	vars = HandleStatement(vars, stmt.Body)
	return vars
}

func HandleRangeStmt(vars []string, stmt *ast.RangeStmt) []string {
	vars = HandleExpr(vars, stmt.Key)
	vars = HandleExpr(vars, stmt.Value)
	vars = HandleStatement(vars, stmt.Body)
	return vars
}

func HandleForStmt(vars []string, stmt *ast.ForStmt) []string {
	vars = HandleStatement(vars, stmt.Init)
	vars = HandleStatement(vars, stmt.Body)
	return vars
}

func HandleBlockStmt(vars []string, stmt *ast.BlockStmt) []string {
	vars = ExtractVariables(vars, stmt.List)
	return vars
}

func HandleIdents(vars []string, idents []*ast.Ident) []string {
	for _, i := range idents {
		vars = append(vars, i.Name)
	}
	return vars
}

func HandleExpr(vars []string, v ast.Expr) []string {
	switch expr := v.(type) {
	case *ast.Ident:
		vars = append(vars, expr.Name)
	case *ast.CallExpr:
		funcName := expr.Fun.(*ast.SelectorExpr).Sel.Name
		if funcName == "Pry" || funcName == "Apply" {
			contexts = append(contexts, PryContext{(int)(expr.Pos() - 1), (int)(expr.End() - 1), vars})
		}
	default:
		fmt.Printf("Unknown %T", expr)
	}
	return vars
}

type PryContext struct {
	Start, End int
	Vars       []string
}
