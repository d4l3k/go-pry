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
	isFunction bool
	defers     []*Defer

	sync.Mutex
}

type Defer struct {
	fun       ast.Expr
	scope     *Scope
	arguments []interface{}
}

func (scope *Scope) Defer(d *Defer) error {
	for ; scope != nil; scope = scope.Parent {
		if scope.isFunction {
			scope.defers = append(scope.defers, d)
			return nil
		}
	}
	return errors.New("defer: can't find function scope")
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

// GetPointer walks the scope and finds the pointer to the value of interest
func (scope *Scope) GetPointer(name string) (val interface{}, exists bool) {
	currentScope := scope
	for !exists && currentScope != nil {
		currentScope.Lock()
		val, exists = currentScope.Vals[name]
		currentScope.Unlock()
		currentScope = currentScope.Parent
	}
	return
}

// Get walks the scope and finds the value of interest
func (scope *Scope) Get(name string) (interface{}, bool) {
	val, exists := scope.GetPointer(name)
	if !exists || val == nil {
		return val, exists
	}
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		return v.Elem().Interface(), exists
	}
	return v.Interface(), exists
}

// Set walks the scope and sets a value in a parent scope if it exists, else current.
func (scope *Scope) Set(name string, val interface{}) {
	if val != nil {
		value := reflect.ValueOf(val)
		if !value.CanAddr() {
			nv := reflect.New(value.Type())
			nv.Elem().Set(value)
			val = nv.Interface()
		} else {
			val = value.Addr().Interface()
		}
	}

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
		node, ok := expr.(ast.Node)
		if !ok {
			return nil, 0, errors.Errorf("expected ast.Node; got %#v", expr)
		}
		return node, shifted, nil
	} else if err != nil {
		return expr, shifted, err
	}
	if expr == nil {
		return nil, 0, errors.Errorf("expression is empty")
	}
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, 0, errors.Errorf("expected CallExpr; got %#v", callExpr)
	}
	return callExpr.Fun.(*ast.FuncLit).Body, shifted, nil
}

// InterpretString interprets a string of go code and returns the result.
func (scope *Scope) InterpretString(exprStr string) (v interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("interpreting %q: %s", exprStr, fmt.Sprint(r))
		}
	}()

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
		args := make([]interface{}, len(e.Args))
		for i, arg := range e.Args {
			interpretedArg, err := scope.Interpret(arg)
			if err != nil {
				return nil, err
			}
			args[i] = interpretedArg
		}

		return scope.ExecuteFunc(e.Fun, args)

	case *ast.GoStmt:
		go func() {
			_, err := scope.NewChild().Interpret(e.Call)
			if err != nil {
				fmt.Printf("goroutine failed: %s\n", err)
			}
		}()
		return nil, nil

	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			n, err := strconv.ParseInt(e.Value, 0, 64)
			if err != nil {
				return nil, err
			}
			return int(n), nil
		case token.FLOAT, token.IMAG:
			v, err := strconv.ParseFloat(e.Value, 64)
			if err != nil {
				return nil, err
			}
			return v, nil
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
			aType := typ.(reflect.Type)
			var slice reflect.Value
			switch aType.Kind() {
			case reflect.Slice:
				slice = reflect.MakeSlice(aType, l, l)
			case reflect.Array:
				slice = reflect.New(aType).Elem()
			default:
				return nil, errors.Errorf("unknown array type %#v", typ)
			}

			if len(e.Elts) > slice.Len() {
				return nil, errors.Errorf("array index %d out of bounds [0:%d]", slice.Len(), slice.Len())
			}

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

		case *ast.Ident, *ast.SelectorExpr:
			objPtr := reflect.New(typ.(reflect.Type))
			obj := objPtr.Elem()
			for i, elem := range e.Elts {
				switch eT := elem.(type) {
				case *ast.BasicLit:
					val, err := scope.Interpret(eT)
					if err != nil {
						return nil, err
					}
					obj.Field(i).Set(reflect.ValueOf(val))

				case *ast.KeyValueExpr:
					key := eT.Key.(*ast.Ident).Name
					val, err := scope.Interpret(eT.Value)
					if err != nil {
						return nil, err
					}
					obj.FieldByName(key).Set(reflect.ValueOf(val))

				default:
					return nil, fmt.Errorf("invalid element type %T %#v to struct literal", eT, eT)
				}
			}
			return obj.Interface(), nil

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
		// Handle indirection cases.
		if e.Op == token.AND {
			ident, isIdent := e.X.(*ast.Ident)
			if !isIdent {
				return nil, errors.Errorf("expected identifier; got %#v", e.X)
			}
			val, exists := scope.GetPointer(ident.Name)
			if !exists {
				return nil, errors.Errorf("unknown identifier %#v", ident)
			}
			return val, nil
		}

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
		rType, ok := typ.(reflect.Type)
		if !ok {
			return nil, errors.Errorf("invalid type %#v", typ)
		}
		if e.Len == nil {
			return reflect.SliceOf(rType), nil
		}

		len, err := scope.Interpret(e.Len)
		if err != nil {
			return nil, err
		}
		lenI, ok := len.(int)
		if !ok {
			return nil, errors.Errorf("expected int; got %#v", len)
		}
		if lenI < 0 {
			return nil, errors.Errorf("negative array size")
		}
		return reflect.ArrayOf(lenI, rType), nil

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
		for xVal.Type().Kind() == reflect.Ptr {
			xVal = xVal.Elem()
		}
		switch xVal.Type().Kind() {
		case reflect.Map:
			val := xVal.MapIndex(reflect.ValueOf(i))
			if !val.IsValid() {
				// If not valid key, return the "zero" type. Eg for int 0, string ""
				return reflect.Zero(xVal.Type().Elem()).Interface(), nil
			}
			return val.Interface(), nil

		case reflect.Slice, reflect.Array:
			iVal, isInt := i.(int)
			if !isInt {
				return nil, fmt.Errorf("index has to be an int not %T", i)
			}
			if iVal >= xVal.Len() || iVal < 0 {
				return nil, errors.New("slice index out of range")
			}

			return xVal.Index(iVal).Interface(), nil

		default:
			return nil, errors.Errorf("invalid X for IndexExpr: %#v", X)
		}

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
		kind := xVal.Kind()
		if kind != reflect.Array && kind != reflect.Slice {
			return nil, errors.Errorf("invalid X for SliceExpr: %#v", X)
		}
		if high == nil {
			high = xVal.Len()
		}
		lowVal, isLowInt := low.(int)
		highVal, isHighInt := high.(int)
		if !isLowInt || !isHighInt {
			return nil, fmt.Errorf("slice: indexes have to be an ints not %T and %T", low, high)
		}
		if lowVal < 0 || highVal >= xVal.Len() || highVal < lowVal {
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
		//define := e.Tok == token.DEFINE
		rhs := make([]interface{}, len(e.Rhs))
		for i, expr := range e.Rhs {
			val, err := scope.Interpret(expr)
			if err != nil {
				return nil, err
			}
			rhs[i] = val
		}

		if len(rhs) == 1 && len(e.Lhs) > 1 && reflect.TypeOf(rhs[0]).Kind() == reflect.Slice {
			rhsV := reflect.ValueOf(rhs[0])
			rhsLen := rhsV.Len()
			if rhsLen != len(e.Lhs) {
				return nil, fmt.Errorf("assignment count mismatch: %d = %d", len(e.Lhs), rhsLen)
			}

			rhs = rhs[:0]

			for i := 0; i < rhsLen; i++ {
				rhs = append(rhs, rhsV.Index(i).Interface())
			}
		}

		if len(rhs) != len(e.Lhs) {
			return nil, fmt.Errorf("assignment count mismatch: %d = %d (%+v)", len(e.Lhs), len(rhs), rhs)
		}

		for i, id := range e.Lhs {
			getR := func(val interface{}) (interface{}, error) {
				r := rhs[i]
				isModAssign := e.Tok != token.ASSIGN && e.Tok != token.DEFINE
				if isModAssign {
					var err error
					r, err = ComputeBinaryOp(val, r, DeAssign(e.Tok))
					if err != nil {
						return nil, err
					}
				}
				return r, nil
			}

			if ident, ok := id.(*ast.Ident); ok {
				val, exists := scope.Get(ident.Name)
				if !exists && (e.Tok != token.DEFINE) {
					return nil, errors.Errorf("undefined %s", ident.Name)
				}

				r, err := getR(val)
				if err != nil {
					return nil, err
				}
				scope.Set(ident.Name, r)
				continue
			} else if idx, ok := id.(*ast.IndexExpr); ok {
				left, err := scope.getValue(idx.X)
				if err != nil {
					return nil, err
				}
				if left.Type().Kind() == reflect.Map {
					index, err := scope.Interpret(idx.Index)
					if err != nil {
						return nil, err
					}
					var val interface{}
					leftV := left.MapIndex(reflect.ValueOf(index))
					if leftV.IsValid() {
						val = leftV.Interface()
					} else {
						val = reflect.Zero(left.Type().Elem()).Interface()
					}
					r, err := getR(val)
					if err != nil {
						return nil, err
					}
					left.SetMapIndex(reflect.ValueOf(index), reflect.ValueOf(r))
					continue
				}
			}

			val, err := scope.getValue(id)
			if err != nil {
				return nil, err
			}

			r, err := getR(val.Interface())
			if err != nil {
				return nil, err
			}
			val.Set(reflect.ValueOf(r))
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
		zero := reflect.Zero(typ.(reflect.Type)).Interface()
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
		chanV := reflect.ValueOf(channel)
		if chanV.Kind() != reflect.Chan {
			return nil, errors.Errorf("expected chan; got %#v", channel)
		}
		succeeded := chanV.TrySend(reflect.ValueOf(val))
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
		return reflect.TypeOf((*interface{})(nil)).Elem(), nil

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

	case *ast.IfStmt:
		currentScope := scope.NewChild()
		if e.Init != nil {
			if _, err := currentScope.Interpret(e.Init); err != nil {
				return nil, err
			}
		}
		cond, err := currentScope.Interpret(e.Cond)
		if err != nil {
			return nil, err
		}
		if cond == true {
			return currentScope.Interpret(e.Body)
		}
		return currentScope.Interpret(e.Else)

	case *ast.DeferStmt:
		var args []interface{}
		for _, arg := range e.Call.Args {
			v, err := scope.Interpret(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, v)
		}
		scope.Defer(&Defer{
			fun:       e.Call.Fun,
			scope:     scope,
			arguments: args,
		})
		return nil, nil

	case *ast.StructType:
		if len(e.Fields.List) > 0 {
			return nil, errors.New("don't support non-empty structs yet")
		}
		return reflect.TypeOf(struct{}{}), nil

	default:
		return nil, fmt.Errorf("unknown node %#v", e)
	}
}

func (scope *Scope) getValue(id ast.Expr) (reflect.Value, error) {
	switch id := id.(type) {
	case *ast.Ident:
		variable := id.Name
		current, exists := scope.GetPointer(variable)
		if !exists {
			return reflect.Value{}, fmt.Errorf("variable %#v is not defined", variable)
		}
		return reflect.ValueOf(current).Elem(), nil

	case *ast.IndexExpr:
		index, err := scope.Interpret(id.Index)
		if err != nil {
			return reflect.Value{}, err
		}
		elem, err := scope.getValue(id.X)
		if err != nil {
			return reflect.Value{}, err
		}

		switch elem.Kind() {
		case reflect.Slice, reflect.Array:
			indexInt, ok := index.(int)
			if !ok {
				return reflect.Value{}, errors.Errorf("expected index to be int, got %#v", index)
			}
			if indexInt >= elem.Len() {
				return reflect.Value{}, errors.Errorf("index out of range")
			}
			return elem.Index(indexInt), nil

		case reflect.Map:
			return elem.MapIndex(reflect.ValueOf(index)), nil

		default:
			return reflect.Value{}, errors.Errorf("unknown type of X %#v", id)
		}

	case *ast.SelectorExpr:
		elem, err := scope.getValue(id.X)
		if err != nil {
			return reflect.Value{}, err
		}
		return elem.FieldByName(id.Sel.Name), nil

	default:
		return reflect.Value{}, errors.Errorf("unknown assignment expr %#v", id)
	}
}

func (scope *Scope) ExecuteFunc(funExpr ast.Expr, args []interface{}) (interface{}, error) {
	fun, err := scope.Interpret(funExpr)
	if err != nil {
		return nil, err
	}

	switch funV := fun.(type) {
	case reflect.Type:
		if len(args) != 1 {
			return nil, errors.Errorf("expected args len = 1; args %#v", args)
		}
		return reflect.ValueOf(args[0]).Convert(funV).Interface(), nil

	case *Func:
		// TODO enforce func return values
		currentScope := scope.NewChild()
		i := 0
		for _, arg := range funV.Def.Type.Params.List {
			for _, name := range arg.Names {
				currentScope.Set(name.Name, args[i])
				i++
			}
		}
		currentScope.isFunction = true
		ret, err := currentScope.Interpret(funV.Def.Body)
		if err != nil {
			return nil, err
		}
		for i := len(currentScope.defers) - 1; i >= 0; i-- {
			d := currentScope.defers[i]
			if _, err := d.scope.ExecuteFunc(d.fun, d.arguments); err != nil {
				return nil, err
			}
		}
		return ret, nil
	}

	funVal := reflect.ValueOf(fun)

	if funVal.Kind() != reflect.Func {
		return nil, errors.Errorf("expected func; got %#v", fun)
	}

	var valueArgs []reflect.Value
	for _, v := range args {
		valueArgs = append(valueArgs, reflect.ValueOf(v))
	}
	funType := funVal.Type()
	if (funType.NumIn() != len(valueArgs) && !funType.IsVariadic()) || (funType.IsVariadic() && len(valueArgs) < funType.NumIn()-1) {
		return nil, errors.Errorf("number of arguments doesn't match function; expected %d; got %+v", funVal.Type().NumIn(), args)
	}
	values := ValuesToInterfaces(funVal.Call(valueArgs))
	if len(values) > 0 {
		if last, ok := values[len(values)-1].(*InterpretError); ok {
			values = values[:len(values)-1]
			if err := last.Error(); err != nil {
				return nil, err
			}
		}
	}

	if len(values) == 0 {
		return nil, nil
	} else if len(values) == 1 {
		return values[0], nil
	}
	return values, nil
}

// ConfigureTypes configures the scope type checker
func (scope *Scope) ConfigureTypes(path string, line int) error {
	scope.path = path
	scope.line = line
	scope.fset = token.NewFileSet() // positions are relative to fset
	scope.config = &types.Config{
		FakeImportC: true,
		Importer:    getImporter(),
	}

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := scope.parseDir()
	if err != nil {
		return errors.Wrapf(err, "parser.ParseDir %q", scope.path)
	}

	for name, file := range f {
		scope.Files[name] = file
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
			err := errors.New(strings.TrimPrefix(err.Error(), scope.path))
			errs = append(errs, errors.Wrapf(err, "path %q", scope.path))
		}
	}
	info := &types.Info{}
	var files []*ast.File
	for _, f := range scope.Files {
		files = append(files, f)
	}
	// these errors should be reported via the error reporter above
	if _, err := scope.config.Check(filepath.Dir(scope.path), scope.fset, files, info); errs == nil && err != nil {
		return nil, []error{err}
	}
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
