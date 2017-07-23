package pry

import (
	"bufio"
	"go/ast"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const placeholder = "pryPlaceholderAutoComplete"

// SuggestionsGoCode is a suggestion engine that uses gocode for autocomplete.
func (scope *Scope) SuggestionsGoCode(line string, index int) ([]string, error) {
	var suggestions []string
	var code string
	for name, file := range scope.Files {
		moddedName := filepath.Dir(name) + "/." + filepath.Base(name) + "pry"
		if scope.path == moddedName {
			name = moddedName
		}
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

			subProcess := exec.Command("gocode", "autocomplete", filepath.Dir(name), strconv.Itoa(i))

			stdin, err := subProcess.StdinPipe()
			if err != nil {
				return nil, err
			}

			stdout, err := subProcess.StdoutPipe()
			if err != nil {
				return nil, err
			}
			defer stdout.Close()

			subProcess.Stderr = os.Stderr

			if err = subProcess.Start(); err != nil {
				return nil, err
			}

			io.WriteString(stdin, code)
			stdin.Close()

			output, err := ioutil.ReadAll(bufio.NewReader(stdout))
			if err != nil {
				return nil, err
			}
			rawSuggestions := strings.Split(string(output), "\n")[1:]
			for _, suggestion := range rawSuggestions {
				trimmed := strings.TrimSpace(suggestion)
				if len(trimmed) > 0 {
					suggestions = append(suggestions, trimmed)
				}
			}
			subProcess.Wait()

			break
		}
	}
	return suggestions, nil
}
