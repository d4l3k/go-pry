package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/d4l3k/go-pry/pry"
)

var contexts []pryContext

// ExecuteGoCmd runs the 'go' command with certain parameters.
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

// InjectPry walks the scope and replaces pry.Pry with pry.Apply(pry.Scope{...}).
func InjectPry(filePath string) (string, error) {
	Debug("Prying into %s\n", filePath)
	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", nil
	}

	contexts = make([]pryContext, 0)

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return "", err
	}

	packagePairs := []string{}
	for _, imp := range f.Imports {
		importStr := imp.Path.Value[1 : len(imp.Path.Value)-1]
		if importStr != "../pry" {
			dir := filepath.Dir(filePath)
			pkg, err := build.Import(importStr, dir, build.AllowBinary)
			if err != nil {
				panic(err)
			}
			importName := pkg.Name
			if imp.Name != nil {
				importName = imp.Name.Name
			}
			pkgAst, err := parser.ParseDir(fset, pkg.Dir, nil, parser.ParseComments)
			if err != nil {
				panic(err)
			}
			pair := "\"" + importName + "\": pry.Package{Name: \"" + pkg.Name + "\", Functions: map[string]interface{}{"
			added := make(map[string]bool)
			for _, nPkg := range pkgAst {
				pair += GetExports(importName, nPkg, added)
			}
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
			vars = extractFields(vars, f.Recv.List)
		}
		if f.Type != nil {
			if f.Type.Params != nil {
				vars = extractFields(vars, f.Type.Params.List)
			}
			if f.Type.Results != nil {
				vars = extractFields(vars, f.Type.Results.List)
			}
		}
		vars = extractVariables(vars, f.Body.List)
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

	Debug(" :: Found %d pry statements.\n", len(contexts))

	for _, context := range contexts {
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

// GenerateFile generates and executes a temp file with the given imports
func GenerateFile(imports []string, extraStatements, generatePath string) error {
	newPath := generatePath
	if len(generatePath) == 0 {
		dir, err := ioutil.TempDir("", "pry")
		if err != nil {
			return err
		}
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				log.Fatal(err)
			}
		}()
		newPath = dir + "/main.go"
	}

	file := "package main\nimport (\n\t\"github.com/d4l3k/go-pry/pry\"\n\n"
	for _, imp := range imports {
		if len(imp) == 0 {
			continue
		}
		file += fmt.Sprintf("\t%#v\n", imp)
	}
	file += ")\nfunc main() {\n\t" + extraStatements + "\n\tpry.Pry()\n}\n"

	ioutil.WriteFile(newPath, []byte(file), 0644)
	InjectPry(newPath)

	if len(generatePath) == 0 {
		ExecuteGoCmd([]string{"run", newPath})
	}
	return nil
}

var debug bool

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	// Catch Ctrl-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
		}
	}()

	// FLAGS
	imports := flag.String("i", "fmt,math", "packages to import, comma seperated")
	revert := flag.Bool("r", true, "whether to revert changes on exit")
	execute := flag.String("e", "", "statements to execute")
	generatePath := flag.String("generate", "", "the path to generate a go-pry injected file - EXPERIMENTAL")
	flag.BoolVar(&debug, "d", false, "display debug statements")

	flag.CommandLine.Usage = func() {
		ExecuteGoCmd([]string{})
		fmt.Println("----")
		fmt.Println("go-pry is an interactive REPL and wrapper around the go command.")
		fmt.Println("You can execute go commands as normal and go-pry will take care of generating the pry code.")
		fmt.Println("Running go-pry with no arguments will drop you into an interactive REPL.")
		flag.PrintDefaults()
		fmt.Println("  revert: cleans up go-pry generated files if not automatically done")
		return
	}
	flag.Parse()
	cmdArgs := flag.Args()
	if len(cmdArgs) == 0 {
		err := GenerateFile(strings.Split(*imports, ","), *execute, *generatePath)
		if err != nil {
			panic(err)
		}
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

	testsRequired := cmdArgs[0] == "test"
	for _, dir := range goDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !testsRequired && strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") {
				return nil
			}
			for _, file := range processedFiles {
				if file == path {
					return nil
				}
			}
			file, err := InjectPry(path)
			if err != nil {
				panic(err)
			}
			if file != "" {
				modifiedFiles = append(modifiedFiles, path)
			}
			return nil
		})
	}

	if cmdArgs[0] == "apply" {
		return
	}

	ExecuteGoCmd(cmdArgs)

	if *revert {
		RevertPry(modifiedFiles)
	}
}

// RevertPry reverts the changes made by InjectPry.
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

// Debug prints debug statements if debug is true.
func Debug(templ string, k ...interface{}) {
	if debug {
		log.Printf(templ, k...)
	}
}

// GetExports returns a string of gocode that represents the exports (constants/functions) of an ast.Package.
func GetExports(importName string, pkg *ast.Package, added map[string]bool) string {
	if pkg.Name == "main" {
		return ""
	}
	vars := ""
	for name, file := range pkg.Files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		match, err := build.Default.MatchFile(path.Dir(name), path.Base(name))
		if err != nil {
			panic(err)
		}
		if !match {
			continue
		}

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
							Debug("TypeSpec ERR %s\n", err.Error())
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
					if debug {
						vars += "\n"
					}
				}
			}
		}
	}
	return vars
}

func filterVars(vars []string) (fVars []string) {
	for _, v := range vars {
		if v != "_" {
			fVars = append(fVars, v)
		}
	}
	return
}

func extractVariables(vars []string, l []ast.Stmt) []string {
	for _, s := range l {
		vars = handleStatement(vars, s)
	}
	return vars
}

func extractFields(vars []string, l []*ast.Field) []string {
	for _, s := range l {
		vars = handleIdents(vars, s.Names)
	}
	return vars
}

func handleStatement(vars []string, s ast.Stmt) []string {
	switch stmt := s.(type) {
	case *ast.ExprStmt:
		vars = handleExpr(vars, stmt.X)
	case *ast.AssignStmt:
		lhsStatements := (*stmt).Lhs
		for _, v := range lhsStatements {
			vars = handleExpr(vars, v)
		}
	case *ast.GoStmt:
		handleExpr(vars, stmt.Call)
	case *ast.IfStmt:
		handleIfStmt(vars, stmt)
	case *ast.DeclStmt:
		decl := stmt.Decl.(*ast.GenDecl)
		if decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				valSpec := spec.(*ast.ValueSpec)
				vars = handleIdents(vars, valSpec.Names)
			}
		}
	case *ast.BlockStmt:
		vars = handleBlockStmt(vars, stmt)
	case *ast.RangeStmt:
		handleRangeStmt(vars, stmt)
	case *ast.ForStmt:
		vars = handleForStmt(vars, stmt)
	default:
		Debug("Unknown %T\n", stmt)
	}
	return vars
}

func handleIfStmt(vars []string, stmt *ast.IfStmt) []string {
	vars = handleStatement(vars, stmt.Init)
	vars = handleStatement(vars, stmt.Body)
	return vars
}

func handleRangeStmt(vars []string, stmt *ast.RangeStmt) []string {
	vars = handleExpr(vars, stmt.Key)
	vars = handleExpr(vars, stmt.Value)
	vars = handleStatement(vars, stmt.Body)
	return vars
}

func handleForStmt(vars []string, stmt *ast.ForStmt) []string {
	vars = handleStatement(vars, stmt.Init)
	vars = handleStatement(vars, stmt.Body)
	return vars
}

func handleBlockStmt(vars []string, stmt *ast.BlockStmt) []string {
	vars = extractVariables(vars, stmt.List)
	return vars
}

func handleIdents(vars []string, idents []*ast.Ident) []string {
	for _, i := range idents {
		vars = append(vars, i.Name)
	}
	return vars
}

func handleExpr(vars []string, v ast.Expr) []string {
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
				contexts = append(contexts, pryContext{(int)(expr.Pos() - 1), (int)(expr.End() - 1), vars})
			}
			//handleExpr(vars, fun.X)
		case *ast.FuncLit:
			handleExpr(vars, fun)
		default:
			Debug("Unknown function type %T\n", fun)
		}
		for _, arg := range expr.Args {
			handleExpr(vars, arg)
		}
	case *ast.FuncLit:
		if expr.Type.Params != nil {
			for _, param := range expr.Type.Params.List {
				for _, name := range param.Names {
					vars = handleExpr(vars, name)
				}
			}
		}
		if expr.Type.Results != nil {
			for _, param := range expr.Type.Results.List {
				for _, name := range param.Names {
					vars = handleExpr(vars, name)
				}
			}
		}
		handleStatement(vars, expr.Body)
	default:
		Debug("Unknown %T\n", expr)
	}
	return vars
}

type pryContext struct {
	Start, End int
	Vars       []string
}
