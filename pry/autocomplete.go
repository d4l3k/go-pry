package pry

import (
	"bufio"
	"fmt"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Suggestions returns auto complete suggestions for the query.
func (scope *Scope) Suggestions(query string) []string {
	terms := []string{}
	if len(query) == 0 {
		terms = append(terms, scope.Keys()...)
	} else if query[len(query)-1] == '.' {
		val, present := scope.Get(query[:len(query)-1])
		if present {
			pkg, isPackage := val.(Package)
			if isPackage {
				val = pkg.Functions
				for k, v := range pkg.Functions {
					if reflect.TypeOf(v).Kind() == reflect.Func {
						k += "("
					}
					terms = append(terms, k)
				}
			}
			typeOf := reflect.TypeOf(val)
			methods := make([]string, typeOf.NumMethod())
			for i := range methods {
				methods[i] = typeOf.Method(i).Name + "("
			}
			terms = append(terms, methods...)

			if typeOf.Kind() == reflect.Struct {
				fields := make([]string, typeOf.NumField())
				for i := range fields {
					fields[i] = typeOf.Field(i).Name
				}
				terms = append(terms, fields...)
			}
		}
	}
	sort.Sort(sort.StringSlice(terms))
	return terms
}

// SuggestionsWIP is a WIP intelligent suggestion provider.
func (scope *Scope) SuggestionsWIP(query string, index int) []string {
	query = strings.Trim(query, " \n\t")

	terms := []string{}
	if len(query) == 0 {
		terms = append(terms, scope.Keys()...)
	}
	node, shifted, err := scope.ParseString(query)
	if err != nil {
		return terms
	}

	index += shifted + 1

	fmt.Println()

	var final ast.Node
	var idents []*ast.Ident

	ast.Walk(walker(func(n ast.Node) bool {
		if n == nil {
			return true
		}
		start := int(n.Pos())
		end := int(n.End())
		if start < index && end >= index {
			expr, ok := n.(*ast.ExprStmt)
			if ok {
				ast.Walk(walker(func(n2 ast.Node) bool {
					ident, ok2 := n2.(*ast.Ident)
					if ok2 {
						idents = append(idents, ident)
					}
					return true
				}), expr)
			}
			final = n
		}
		return true
	}), node)

	fmt.Printf("F: %#v I: %#v\n", final, idents)

	_, isIdent := final.(*ast.Ident)
	if isIdent {
		typ := scope.getType(idents)
		_ = typ
	}

	_ = node

	return terms
}

const placeholder = "pryPlaceholderAutoComplete"

// SuggestionsGoCode is a suggestion engine that uses gocode for autocomplete.
func (scope *Scope) SuggestionsGoCode(line string, index int) []string {
	var suggestions []string
	var code string
	for name, file := range scope.Files {
		name = filepath.Dir(name) + "/." + filepath.Base(name) + "pry"
		if name == scope.path {

			ast.Walk(walker(func(n ast.Node) bool {
				switch s := n.(type) {
				case *ast.BlockStmt:
					for i, stmt := range s.List {
						pos := scope.fset.Position(stmt.Pos())
						if pos.Line == scope.line {
							r := scope.Render(stmt)
							if strings.HasPrefix(r, "pry.Apply") {
								var iStmt []ast.Stmt
								iStmt = append(iStmt, ast.Stmt(&ast.ExprStmt{X: ast.NewIdent(placeholder)}))
								oldList := make([]ast.Stmt, len(s.List))
								copy(oldList, s.List)

								s.List = append(s.List, make([]ast.Stmt, len(iStmt))...)

								copy(s.List[i+len(iStmt):], s.List[i:])
								copy(s.List[i:], iStmt)

								code = scope.Render(file)
								s.List = oldList
								return false
							}
						}
					}
				}
				return true
			}), file)
			i := strings.Index(code, placeholder) + index
			code = strings.Replace(code, placeholder, line, 1)
			//fmt.Println("COMPLETION", i, code)

			subProcess := exec.Command("gocode", "autocomplete", filepath.Dir(name), strconv.Itoa(i))

			stdin, err := subProcess.StdinPipe()
			if err != nil {
				fmt.Println(err)
				break
			}

			stdout, err := subProcess.StdoutPipe()
			if err != nil {
				fmt.Println(err)
				break
			}
			defer stdout.Close()

			subProcess.Stderr = os.Stderr

			if err = subProcess.Start(); err != nil {
				fmt.Println("An error occured: ", err) //replace with logger, or anything you want
			}

			io.WriteString(stdin, code)
			stdin.Close()

			output, err := ioutil.ReadAll(bufio.NewReader(stdout))
			if err != nil {
				fmt.Println(err)
				break
			}
			suggestions = strings.Split(string(output), "\n")[1:]
			subProcess.Wait()

			break
		}
	}
	return suggestions
}

func (scope *Scope) getType(idents []*ast.Ident) interface{} {
	var item interface{}
	for _, ident := range idents {
		if item == nil {
			val, exists := scope.Get(ident.Name)
			if exists {
				item = val
			} else {
				return nil
			}
		}
	}
	return item
}
