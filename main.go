package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

var contexts []PryContext

func main() {
	filePath := os.Args[len(os.Args)-1]

	fmt.Println("Prying into ", filePath)

	contexts = make([]PryContext, 0)

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
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

	for _, context := range contexts {
		vars := FilterVars(context.Vars)
		obj := "map[string]interface{}{ "
		for _, v := range vars {
			obj = obj + "\"" + v + "\": " + v + ", "
		}
		obj += "}"
		text := "pry.Apply(" + obj + ")\n"
		fileText = fileText[0:context.Start+offset] + text + fileText[context.End+offset:]
		offset = len(text) - (context.End - context.Start)
	}

	tmpPath := filePath + ".go"
	ioutil.WriteFile(tmpPath, ([]byte)(fileText), 0644)

	binary, lookErr := exec.LookPath("go")
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{"go", "run", tmpPath}
	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}

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
		if funcName == "Pry" {
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
