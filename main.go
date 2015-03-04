package main

import (
	"flag"
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
	filePath := flag.String("f", "", "The file to run.")
	flag.Parse()

	fmt.Println("Parsing:", *filePath)

	contexts = make([]PryContext, 0)

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, *filePath, nil, 0)
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
	fmt.Println("FINAL VARS", vars, "FUNCS", funcs)

	template, err := ioutil.ReadFile("pry.go.tmpl")
	if err != nil {
		panic(err)
	}

	fileTextBytes, err := ioutil.ReadFile(*filePath)
	if err != nil {
		panic(err)
	}

	fileText := (string)(fileTextBytes)

	offset := 0

	fmt.Println("Applying PRY")
	for _, context := range contexts {
		vars := FilterVars(context.Vars)
		obj := "map[string]interface{}{ "
		for _, v := range vars {
			obj = obj + "\"" + v + "\": " + v + ", "
		}
		obj += "}"
		text := "pry.Apply(" + obj + ")\n" + (string)(template)
		fileText = fileText[0:context.Start+offset] + text + fileText[context.End+offset:]
		offset = len(text) - (context.End - context.Start)
	}

	tmpPath := *filePath + ".go"
	ioutil.WriteFile(tmpPath, ([]byte)(fileText), 0644)
	fmt.Println("DONE!")

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
	fmt.Println("VARS", vars)
	for _, s := range l {
		vars = HandleStatement(vars, s)
	}
	return vars
}

func HandleStatement(vars []string, s ast.Stmt) []string {
	switch stmt := s.(type) {
	case *ast.ExprStmt:
		fmt.Println("EXPR", stmt)
		vars = HandleExpr(vars, stmt.X)
	case *ast.AssignStmt:
		lhsStatements := (*stmt).Lhs
		fmt.Println("ASSIGN", lhsStatements)
		for _, v := range lhsStatements {
			vars = HandleExpr(vars, v)
		}
	case *ast.IfStmt:
		HandleIfStmt(vars, stmt)
	case *ast.DeclStmt:
		decl := stmt.Decl.(*ast.GenDecl)
		fmt.Println("DECL", decl)
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
		fmt.Printf("Unknown %T", stmt)
	}
	fmt.Println("VARS", vars)
	return vars
}

func HandleIfStmt(vars []string, stmt *ast.IfStmt) []string {
	fmt.Println("IF", stmt)
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
	fmt.Println("BLOCK", stmt)
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
		fmt.Println("IDENT", expr)
		vars = append(vars, expr.Name)
	case *ast.CallExpr:
		funcName := expr.Fun.(*ast.SelectorExpr).Sel.Name
		fmt.Println("CALL EXPR", funcName, expr.Pos(), expr.End())
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
