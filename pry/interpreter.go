package pry

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"go/types"
	// Used by types for import determination
)

var (
	// ErrChanSendFailed occurs when a channel is full or there are no receivers
	// available.
	ErrChanSendFailed = errors.New("failed to send, channel full or no receivers")

	// ErrBranchBreak is an internal error thrown when a for loop breaks.
	ErrBranchBreak = errors.New("branch break")
	// ErrBranchContinue is an internal error thrown when a for loop continues.
	ErrBranchContinue = errors.New("branch continue")
)

// Scope is a string-interface key-value pair that represents variables/functions in scope.
type Scope struct {
	Vals   map[string]interface{}
	Parent *Scope
	Files  map[string]*ast.File
	config *types.Config
	path   string
	line   int
	fset   *token.FileSet

	isSelect   bool
	typeAssert reflect.Type

	sync.Mutex
}

// NewScope creates a new initialized scope
func NewScope() *Scope {
	s := &Scope{
		Vals:  map[string]interface{}{},
		Files: map[string]*ast.File{},
	}
	s.Set("_pryScope", s)
	return s
}

// Get walks the scope and finds the value of interest
func (scope *Scope) Get(name string) (val interface{}, exists bool) {
	currentScope := scope
	for !exists && currentScope != nil {
		currentScope.Lock()
		val, exists = currentScope.Vals[name]
		currentScope.Unlock()
		currentScope = currentScope.Parent
	}
	return
}

// Set walks the scope and sets a value in a parent scope if it exists, else current.
func (scope *Scope) Set(name string, val interface{}) {
	exists := false
	currentScope := scope
	for !exists && currentScope != nil {
		currentScope.Lock()
		_, exists = currentScope.Vals[name]
		if exists {
			currentScope.Vals[name] = val
		}
		currentScope.Unlock()
		currentScope = currentScope.Parent
	}
	if !exists {
		scope.Lock()
		scope.Vals[name] = val
		scope.Unlock()
	}
}

// Keys returns all keys in scope
func (scope *Scope) Keys() (keys []string) {
	currentScope := scope
	for currentScope != nil {
		currentScope.Lock()
		for k := range currentScope.Vals {
			keys = append(keys, k)
		}
		currentScope.Unlock()
		currentScope = scope.Parent
	}
	return
}

// NewChild creates a scope under the existing scope.
func (scope *Scope) NewChild() *Scope {
	s := NewScope()
	s.Parent = scope
	return s
}

// Func represents an interpreted function definition.
type Func struct {
	Def *ast.FuncLit
}

// ParseString parses go code into the ast nodes.
func (scope *Scope) ParseString(exprStr string) (ast.Node, int, error) {
	exprStr = strings.Trim(exprStr, " \n\t")
	wrappedExpr := "func(){" + exprStr + "}()"
	shifted := 7
	expr, err := parser.ParseExpr(wrappedExpr)
	if err != nil && strings.HasPrefix(err.Error(), "1:8: expected statement, found '") {
		expr, err = parser.ParseExpr(exprStr)
		shifted = 0
		if err != nil {
			return expr, shifted, err
		}
		return expr.(ast.Node), shifted, nil
	} else if err != nil {
		return expr, shifted, err
	} else {
		return expr.(*ast.CallExpr).Fun.(*ast.FuncLit).Body, shifted, nil
	}
}

// InterpretString interprets a string of go code and returns the result.
func (scope *Scope) InterpretString(exprStr string) (interface{}, error) {
	node, _, err := scope.ParseString(exprStr)
	if err != nil {
		return node, err
	}
	errs := scope.CheckStatement(node)
	if len(errs) > 0 {
		return node, errs[0]
	}
	return scope.Interpret(node)
}

// Interpret interprets an ast.Node and returns the value.
func (scope *Scope) Interpret(expr ast.Node) (interface{}, error) {
	builtinScope := map[string]interface{}{
		"nil":    nil,
		"true":   true,
		"false":  false,
		"append": Append,
		"make":   Make,
		"len":    Len,
		"close":  Close,
	}

	switch e := expr.(type) {
	case *ast.Ident:

		typ, err := StringToType(e.Name)
		if err == nil {
			return typ, err
		}

		obj, exists := scope.Get(e.Name)
		if !exists {
			// TODO make builtinScope root of other scopes
			obj, exists = builtinScope[e.Name]
			if !exists {
				return nil, fmt.Errorf("can't find EXPR %s", e.Name)
			}
		}
		return obj, nil

	case *ast.SelectorExpr:
		X, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		sel := e.Sel

		rVal := reflect.ValueOf(X)
		if rVal.Kind() != reflect.Struct && rVal.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("%#v is not a struct and thus has no field %#v", X, sel.Name)
		}

		pkg, isPackage := X.(Package)
		if isPackage {
			obj, isPresent := pkg.Functions[sel.Name]
			if isPresent {
				return obj, nil
			}
			return nil, fmt.Errorf("unknown field %#v", sel.Name)
		}

		if method := rVal.MethodByName(sel.Name); method.IsValid() {
			return method.Interface(), nil
		}
		if rVal.Kind() == reflect.Ptr {
			rVal = rVal.Elem()
		}
		if field := rVal.FieldByName(sel.Name); field.IsValid() {
			return field.Interface(), nil
		}
		return nil, fmt.Errorf("unknown field %#v", sel.Name)

	case *ast.CallExpr:
		fun, err := scope.Interpret(e.Fun)
		if err != nil {
			return nil, err
		}

		args := make([]reflect.Value, len(e.Args))
		for i, arg := range e.Args {
			interpretedArg, err := scope.Interpret(arg)
			if err != nil {
				return nil, err
			}
			args[i] = reflect.ValueOf(interpretedArg)
		}

		switch funV := fun.(type) {
		case reflect.Type:
			return args[0].Convert(funV).Interface(), nil
		case *Func:
			// TODO enforce func return values
			return scope.Interpret(funV.Def.Body)
		}

		funVal := reflect.ValueOf(fun)

		values := ValuesToInterfaces(funVal.Call(args))
		if len(values) == 0 {
			return nil, nil
		} else if len(values) == 1 {
			return values[0], nil
		}
		err, _ = values[1].(error)
		return values[0], err

	case *ast.GoStmt:
		go func() {
			_, err := scope.NewChild().Interpret(e.Call)
			if err != nil {
				fmt.Fprintf(out, "goroutine failed: %s\n", err)
			}
		}()
		return nil, nil

	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			n, err := strconv.ParseInt(e.Value, 0, 64)
			return int(n), err
		case token.FLOAT, token.IMAG:
			return strconv.ParseFloat(e.Value, 64)
		case token.CHAR:
			return (rune)(e.Value[1]), nil
		case token.STRING:
			return e.Value[1 : len(e.Value)-1], nil
		default:
			return nil, fmt.Errorf("unknown basic literal %d", e.Kind)
		}

	case *ast.CompositeLit:
		typ, err := scope.Interpret(e.Type)
		if err != nil {
			return nil, err
		}

		switch t := e.Type.(type) {
		case *ast.ArrayType:
			l := len(e.Elts)
			slice := reflect.MakeSlice(typ.(reflect.Type), l, l)
			for i, elem := range e.Elts {
				elemValue, err := scope.Interpret(elem)
				if err != nil {
					return nil, err
				}
				slice.Index(i).Set(reflect.ValueOf(elemValue))
			}
			return slice.Interface(), nil

		case *ast.MapType:
			nMap := reflect.MakeMap(typ.(reflect.Type))
			for _, elem := range e.Elts {
				switch eT := elem.(type) {
				case *ast.KeyValueExpr:
					key, err := scope.Interpret(eT.Key)
					if err != nil {
						return nil, err
					}
					val, err := scope.Interpret(eT.Value)
					if err != nil {
						return nil, err
					}
					nMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))

				default:
					return nil, fmt.Errorf("invalid element type %#v to map. Expecting key value pair", eT)
				}
			}
			return nMap.Interface(), nil

		default:
			return nil, fmt.Errorf("unknown composite literal %#v", t)
		}

	case *ast.BinaryExpr:
		x, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		y, err := scope.Interpret(e.Y)
		if err != nil {
			return nil, err
		}
		return ComputeBinaryOp(x, y, e.Op)

	case *ast.UnaryExpr:
		x, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		return scope.ComputeUnaryOp(x, e.Op)

	case *ast.ArrayType:
		typ, err := scope.Interpret(e.Elt)
		if err != nil {
			return nil, err
		}
		arrType := reflect.SliceOf(typ.(reflect.Type))
		return arrType, nil

	case *ast.MapType:
		keyType, err := scope.Interpret(e.Key)
		if err != nil {
			return nil, err
		}
		valType, err := scope.Interpret(e.Value)
		if err != nil {
			return nil, err
		}
		mapType := reflect.MapOf(keyType.(reflect.Type), valType.(reflect.Type))
		return mapType, nil

	case *ast.ChanType:
		typeI, err := scope.Interpret(e.Value)
		if err != nil {
			return nil, err
		}
		typ, isType := typeI.(reflect.Type)
		if !isType {
			return nil, fmt.Errorf("chan needs to be passed a type not %T", typ)
		}
		return reflect.ChanOf(reflect.BothDir, typ), nil

	case *ast.IndexExpr:
		X, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		i, err := scope.Interpret(e.Index)
		if err != nil {
			return nil, err
		}
		xVal := reflect.ValueOf(X)
		if reflect.TypeOf(X).Kind() == reflect.Map {
			val := xVal.MapIndex(reflect.ValueOf(i))
			if !val.IsValid() {
				// If not valid key, return the "zero" type. Eg for int 0, string ""
				return reflect.Zero(xVal.Type().Elem()).Interface(), nil
			}
			return val.Interface(), nil
		}

		iVal, isInt := i.(int)
		if !isInt {
			return nil, fmt.Errorf("index has to be an int not %T", i)
		}
		if iVal >= xVal.Len() || iVal < 0 {
			return nil, errors.New("slice index out of range")
		}
		return xVal.Index(iVal).Interface(), nil
	case *ast.SliceExpr:
		low, err := scope.Interpret(e.Low)
		if err != nil {
			return nil, err
		}
		high, err := scope.Interpret(e.High)
		if err != nil {
			return nil, err
		}
		X, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		xVal := reflect.ValueOf(X)
		if low == nil {
			low = 0
		}
		if high == nil {
			high = xVal.Len()
		}
		lowVal, isLowInt := low.(int)
		highVal, isHighInt := high.(int)
		if !isLowInt || !isHighInt {
			return nil, fmt.Errorf("slice: indexes have to be an ints not %T and %T", low, high)
		}
		if lowVal < 0 || highVal >= xVal.Len() {
			return nil, errors.New("slice: index out of bounds")
		}
		return xVal.Slice(lowVal, highVal).Interface(), nil

	case *ast.ParenExpr:
		return scope.Interpret(e.X)

	case *ast.FuncLit:
		return &Func{e}, nil
	case *ast.BlockStmt:
		var outFinal interface{}
		for _, stmts := range e.List {
			out, err := scope.Interpret(stmts)
			if err != nil {
				return out, err
			}
			outFinal = out
		}
		return outFinal, nil

	case *ast.ReturnStmt:
		results := make([]interface{}, len(e.Results))
		for i, result := range e.Results {
			out, err := scope.Interpret(result)
			if err != nil {
				return out, err
			}
			results[i] = out
		}

		if len(results) == 0 {
			return nil, nil
		} else if len(results) == 1 {
			return results[0], nil
		}
		return results, nil

	case *ast.AssignStmt:
		// TODO implement type checking
		define := e.Tok == token.DEFINE
		lhs := make([]string, len(e.Lhs))
		for i, id := range e.Lhs {
			lhsIdent, isIdent := id.(*ast.Ident)
			if !isIdent {
				return nil, fmt.Errorf("%#v assignment is not ident", id)
			}
			lhs[i] = lhsIdent.Name
		}
		rhs := make([]interface{}, len(e.Rhs))
		for i, expr := range e.Rhs {
			val, err := scope.Interpret(expr)
			if err != nil {
				return nil, err
			}
			rhs[i] = val
		}
		if len(rhs) != 1 && len(rhs) != len(lhs) {
			return nil, fmt.Errorf("assignment count mismatch: %d = %d", len(lhs), len(rhs))
		}
		if len(rhs) == 1 && len(lhs) > 1 && reflect.TypeOf(rhs[0]).Kind() == reflect.Slice {
			rhsV := reflect.ValueOf(rhs[0])
			rhsLen := rhsV.Len()
			if rhsLen != len(lhs) {
				return nil, fmt.Errorf("assignment count mismatch: %d = %d", len(lhs), rhsLen)
			}
			for i := 0; i < rhsLen; i++ {
				variable := lhs[i]
				_, exists := scope.Get(variable)
				if !exists && !define {
					return nil, fmt.Errorf("variable %#v is not defined", variable)
				}
				scope.Set(variable, rhsV.Index(i).Interface())
			}
		} else {
			for i, r := range rhs {
				variable := lhs[i]
				_, exists := scope.Get(variable)
				if !exists && !define {
					return nil, fmt.Errorf("variable %#v is not defined", variable)
				}
				scope.Set(variable, r)
			}
		}
		if len(rhs) > 1 {
			return rhs, nil
		}
		return rhs[0], nil
	case *ast.IncDecStmt:
		var dir string
		switch e.Tok {
		case token.INC:
			dir = "1"
		case token.DEC:
			dir = "-1"
		}
		ass := &ast.AssignStmt{
			Tok: token.ASSIGN,
			Lhs: []ast.Expr{e.X},
			Rhs: []ast.Expr{&ast.BinaryExpr{
				X:  e.X,
				Op: token.ADD,
				Y: &ast.BasicLit{
					Kind:  token.INT,
					Value: dir,
				},
			}},
		}
		return scope.Interpret(ass)
	case *ast.RangeStmt:
		s := scope.NewChild()
		ranger, err := s.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		var key, value string
		if e.Key != nil {
			key = e.Key.(*ast.Ident).Name
		}
		if e.Value != nil {
			value = e.Value.(*ast.Ident).Name
		}
		rv := reflect.ValueOf(ranger)
		switch rv.Type().Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < rv.Len(); i++ {
				if len(key) > 0 {
					s.Set(key, i)
				}
				if len(value) > 0 {
					s.Set(value, rv.Index(i).Interface())
				}
				s.Interpret(e.Body)
			}
		case reflect.Map:
			keys := rv.MapKeys()
			for _, keyV := range keys {
				if len(key) > 0 {
					s.Set(key, keyV.Interface())
				}
				if len(value) > 0 {
					s.Set(value, rv.MapIndex(keyV).Interface())
				}
				s.Interpret(e.Body)
			}
		default:
			return nil, fmt.Errorf("ranging on %s is unsupported", rv.Type().Kind().String())
		}
		return nil, nil
	case *ast.ExprStmt:
		return scope.Interpret(e.X)
	case *ast.DeclStmt:
		return scope.Interpret(e.Decl)
	case *ast.GenDecl:
		for _, spec := range e.Specs {
			if _, err := scope.Interpret(spec); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *ast.ValueSpec:
		typ, err := scope.Interpret(e.Type)
		if err != nil {
			return nil, err
		}
		var zero interface{}
		if typ != nil {
			zero = reflect.Zero(typ.(reflect.Type)).Interface()
		}
		for i, name := range e.Names {
			if len(e.Values) > i {
				v, err := scope.Interpret(e.Values[i])
				if err != nil {
					return nil, err
				}
				scope.Set(name.Name, v)
			} else {
				scope.Set(name.Name, zero)
			}
		}
		return nil, nil
	case *ast.ForStmt:
		s := scope.NewChild()
		if e.Init != nil {
			if _, err := s.Interpret(e.Init); err != nil {
				return nil, err
			}
		}
		var err error
		var last interface{}
		for {
			if e.Cond != nil {
				cond, err := s.Interpret(e.Cond)
				if err != nil {
					return nil, err
				}
				if cont, ok := cond.(bool); !ok {
					return nil, fmt.Errorf("for loop requires a boolean condition not %#v", cond)
				} else if !cont {
					return last, nil
				}
			}

			last, err = s.Interpret(e.Body)
			if err == ErrBranchBreak {
				break
			} else if err != nil && err != ErrBranchContinue {
				return nil, err
			}

			if e.Post != nil {
				if _, err := s.Interpret(e.Post); err != nil {
					return nil, err
				}
			}
		}
		return last, nil

	case *ast.BranchStmt:
		switch e.Tok {
		case token.BREAK:
			return nil, ErrBranchBreak
		case token.CONTINUE:
			return nil, ErrBranchContinue
		default:
			return nil, fmt.Errorf("unsupported BranchStmt %#v", e)
		}

	case *ast.SendStmt:
		val, err := scope.Interpret(e.Value)
		if err != nil {
			return nil, err
		}
		channel, err := scope.Interpret(e.Chan)
		if err != nil {
			return nil, err
		}
		succeeded := reflect.ValueOf(channel).TrySend(reflect.ValueOf(val))
		if !succeeded {
			return nil, ErrChanSendFailed
		}
		return nil, nil

	case *ast.SelectStmt:
		list := e.Body.List
		var defaultCase *ast.CommClause

		// We're using a map here since we want iteration on clauses to be
		// pseudo-random.
		clauses := map[int]*ast.CommClause{}
		for i, stmt := range list {
			cc := stmt.(*ast.CommClause)
			if cc.Comm == nil {
				defaultCase = cc
			} else {
				clauses[i] = cc
			}
		}

		for {
			for _, cc := range clauses {
				child := scope.NewChild()
				child.isSelect = true
				_, err := child.Interpret(cc.Comm)
				child.isSelect = false
				if err == ErrChanSendFailed || err == ErrBranchContinue || err == ErrChanRecvInSelect {
					continue
				} else if err != nil {
					return nil, err
				}
				return child.Interpret(cc)
			}
			if defaultCase != nil {
				child := scope.NewChild()
				return child.Interpret(defaultCase)
			}
			time.Sleep(10 * time.Millisecond)
		}

	case *ast.SwitchStmt:
		list := e.Body.List
		var defaultCase *ast.CaseClause
		var clauses []*ast.CaseClause
		for _, stmt := range list {
			cc := stmt.(*ast.CaseClause)
			if cc.List == nil {
				defaultCase = cc
			} else {
				clauses = append(clauses, cc)
			}
		}

		currentScope := scope.NewChild()
		if e.Init != nil {
			if _, err := currentScope.Interpret(e.Init); err != nil {
				return nil, err
			}
		}

		var err error
		var want interface{}
		if e.Tag != nil {
			want, err = currentScope.Interpret(e.Tag)
		} else {
			want = true
		}
		if err != nil {
			return nil, err
		}

		for _, cc := range clauses {
			for _, c := range cc.List {
				child := currentScope.NewChild()
				out, err := child.Interpret(c)
				if err != nil {
					return nil, err
				}
				if reflect.DeepEqual(out, want) {
					return child.Interpret(cc)
				}
			}
		}
		if defaultCase != nil {
			child := scope.NewChild()
			return child.Interpret(defaultCase)
		}
		return nil, nil

	case *ast.TypeSwitchStmt:
		list := e.Body.List
		var defaultCase *ast.CaseClause
		var clauses []*ast.CaseClause
		for _, stmt := range list {
			cc := stmt.(*ast.CaseClause)
			if cc.List == nil {
				defaultCase = cc
			} else {
				clauses = append(clauses, cc)
			}
		}

		currentScope := scope.NewChild()
		if e.Init != nil {
			if _, err := currentScope.Interpret(e.Init); err != nil {
				return nil, err
			}
		}

		var want reflect.Type
		if e.Assign != nil {
			_, err := currentScope.Interpret(e.Assign)
			if err != nil {
				return nil, err
			}
			want = currentScope.typeAssert
		}

		for _, cc := range clauses {
			for _, c := range cc.List {
				child := currentScope.NewChild()
				out, err := child.Interpret(c)
				if err != nil {
					return nil, err
				}
				if out == want {
					return child.Interpret(cc)
				}
			}
		}
		if defaultCase != nil {
			child := scope.NewChild()
			return child.Interpret(defaultCase)
		}
		return nil, nil

	case *ast.CommClause:
		return scope.Interpret(&ast.BlockStmt{List: e.Body})

	case *ast.CaseClause:
		return scope.Interpret(&ast.BlockStmt{List: e.Body})

	case *ast.InterfaceType:
		if len(e.Methods.List) > 0 {
			return nil, fmt.Errorf("don't support non-anonymous interfaces yet")
		}
		return reflect.TypeOf(nil), nil

	case *ast.TypeAssertExpr:
		out, err := scope.Interpret(e.X)
		if err != nil {
			return nil, err
		}
		outType := reflect.TypeOf(out)
		if e.Type == nil {
			scope.typeAssert = outType
			return out, nil
		}
		typ, err := scope.Interpret(e.Type)
		if err != nil {
			return nil, err
		}
		if typ != outType {
			return nil, fmt.Errorf("%#v is not of type %#v, is %T", out, typ, out)
		}
		return out, nil

	default:
		return nil, fmt.Errorf("unknown node %#v", e)
	}
}

// ConfigureTypes configures the scope type checker
func (scope *Scope) ConfigureTypes(path string, line int) error {
	scope.path = path
	scope.line = line
	scope.fset = token.NewFileSet() // positions are relative to fset
	scope.config = &types.Config{
		FakeImportC: true,
		Importer:    gcImporter,
	}

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseDir(scope.fset, filepath.Dir(scope.path), nil, 0)
	if err != nil {
		return errors.Wrapf(err, "parser.ParseDir %q", scope.path)
	}

	for _, pkg := range f {
		for name, file := range pkg.Files {
			scope.Files[name] = file
		}
	}

	_, errs := scope.TypeCheck()
	if len(errs) > 0 {
		return errors.Wrap(errs[0], "failed to TypeCheck")
	}

	return nil
}

// walker adapts a function to satisfy the ast.Visitor interface.
// The function return whether the walk should proceed into the node's children.
type walker func(ast.Node) bool

func (w walker) Visit(node ast.Node) ast.Visitor {
	if w(node) {
		return w
	}
	return nil
}

// CheckStatement checks if a statement is type safe
func (scope *Scope) CheckStatement(node ast.Node) (errs []error) {
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
								switch s2 := node.(type) {
								case *ast.BlockStmt:
									iStmt = append(iStmt, s2.List...)
								case ast.Stmt:
									iStmt = append(iStmt, s2)
								case ast.Expr:
									iStmt = append(iStmt, ast.Stmt(&ast.ExprStmt{X: s2}))
								default:
									errs = append(errs, errors.New("not a statement"))
									return false
								}
								oldList := make([]ast.Stmt, len(s.List))
								copy(oldList, s.List)

								s.List = append(s.List, make([]ast.Stmt, len(iStmt))...)

								copy(s.List[i+len(iStmt):], s.List[i:])
								copy(s.List[i:], iStmt)

								_, errs = scope.TypeCheck()
								if len(errs) > 0 {
									s.List = oldList
									return false
								}
								return false
							}
						}
					}
				}
				return true
			}), file)
		}
	}
	return
}

// Render renders an ast node
func (scope *Scope) Render(x ast.Node) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, scope.fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}

// TypeCheck does type checking and returns the info object
func (scope *Scope) TypeCheck() (*types.Info, []error) {
	var errs []error
	scope.config.Error = func(err error) {
		if !strings.HasSuffix(err.Error(), "not used") {
			errs = append(errs, errors.New(strings.TrimPrefix(err.Error(), scope.path)))
		}
	}
	info := &types.Info{}
	var files []*ast.File
	for _, f := range scope.Files {
		files = append(files, f)
	}
	scope.config.Check(filepath.Dir(scope.path), scope.fset, files, info)
	return info, errs
}

// StringToType returns the reflect.Type corresponding to the type string provided. Ex: StringToType("int")
func StringToType(str string) (reflect.Type, error) {
	builtinTypes := map[string]reflect.Type{
		"bool":       reflect.TypeOf(true),
		"byte":       reflect.TypeOf(byte(0)),
		"rune":       reflect.TypeOf(rune(0)),
		"string":     reflect.TypeOf(""),
		"int":        reflect.TypeOf(int(0)),
		"int8":       reflect.TypeOf(int8(0)),
		"int16":      reflect.TypeOf(int16(0)),
		"int32":      reflect.TypeOf(int32(0)),
		"int64":      reflect.TypeOf(int64(0)),
		"uint":       reflect.TypeOf(uint(0)),
		"uint8":      reflect.TypeOf(uint8(0)),
		"uint16":     reflect.TypeOf(uint16(0)),
		"uint32":     reflect.TypeOf(uint32(0)),
		"uint64":     reflect.TypeOf(uint64(0)),
		"uintptr":    reflect.TypeOf(uintptr(0)),
		"float32":    reflect.TypeOf(float32(0)),
		"float64":    reflect.TypeOf(float64(0)),
		"complex64":  reflect.TypeOf(complex64(0)),
		"complex128": reflect.TypeOf(complex128(0)),
		"error":      reflect.TypeOf(errors.New("")),
	}
	val, present := builtinTypes[str]
	if !present {
		return nil, fmt.Errorf("type %#v is not in table", str)
	}
	return val, nil
}

// ValuesToInterfaces converts a slice of []reflect.Value to []interface{}
func ValuesToInterfaces(vals []reflect.Value) []interface{} {
	inters := make([]interface{}, len(vals))
	for i, val := range vals {
		inters[i] = val.Interface()
	}
	return inters
}
