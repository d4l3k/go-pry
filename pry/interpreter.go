package pry

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

// Scope is a string-interface key-value pair that represents variables/functions in scope.
type Scope struct {
	Vals   map[string]interface{}
	Parent *Scope
}

// NewScope creates a new initialized scope
func NewScope() *Scope {
	return &Scope{
		map[string]interface{}{},
		nil,
	}
}

// Get walks the scope and finds the value of interest
func (scope *Scope) Get(name string) (val interface{}, exists bool) {
	currentScope := scope
	for !exists && currentScope != nil {
		val, exists = currentScope.Vals[name]
		currentScope = scope.Parent
	}
	return
}

// Set walks the scope and sets a value in a parent scope if it exists, else current.
func (scope *Scope) Set(name string, val interface{}) {
	exists := false
	currentScope := scope
	for !exists && currentScope != nil {
		_, exists = currentScope.Vals[name]
		if exists {
			currentScope.Vals[name] = val
		}
		currentScope = scope.Parent
	}
	if !exists {
		scope.Vals[name] = val
	}
}

// Keys returns all keys in scope
func (scope *Scope) Keys() (keys []string) {
	currentScope := scope
	for currentScope != nil {
		for k := range currentScope.Vals {
			keys = append(keys, k)
		}
		currentScope = scope.Parent
	}
	return
}

// Func represents an interpreted function definition.
type Func struct {
	Def *ast.FuncLit
}

// InterpretString interprets a string of go code and returns the result.
func (scope *Scope) InterpretString(exprStr string) (interface{}, error) {
	exprStr = strings.Trim(exprStr, " \n\t")
	wrappedExpr := "func(){" + exprStr + "}()"
	expr, err := parser.ParseExpr(wrappedExpr)
	if err != nil && strings.HasPrefix(err.Error(), "1:8: expected statement, found '") {
		expr, err = parser.ParseExpr(exprStr)
	}
	if err != nil {
		return nil, err
	}
	return scope.InterpretExpr(expr)
}

// InterpretExpr interprets an ast.Expr and returns the value.
func (scope *Scope) InterpretExpr(expr ast.Expr) (interface{}, error) {
	builtinScope := map[string]interface{}{
		"nil":    nil,
		"true":   true,
		"false":  false,
		"append": Append,
		"make":   Make,
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
		X, err := scope.InterpretExpr(e.X)
		if err != nil {
			return nil, err
		}
		sel := e.Sel

		rVal := reflect.ValueOf(X)
		if rVal.Kind() != reflect.Struct {
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

		zero := reflect.ValueOf(nil)
		field := rVal.FieldByName(sel.Name)
		if field != zero {
			return field.Interface(), nil
		}
		method := rVal.MethodByName(sel.Name)
		if method != zero {
			return method.Interface(), nil
		}
		return nil, fmt.Errorf("unknown field %#v", sel.Name)

	case *ast.CallExpr:
		fun, err := scope.InterpretExpr(e.Fun)
		if err != nil {
			return nil, err
		}

		args := make([]reflect.Value, len(e.Args))
		for i, arg := range e.Args {
			interpretedArg, err := scope.InterpretExpr(arg)
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
			return InterpretStmt(scope, funV.Def.Body)
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

	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return strconv.Atoi(e.Value)
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
		typ, err := scope.InterpretExpr(e.Type)
		if err != nil {
			return nil, err
		}

		switch t := e.Type.(type) {
		case *ast.ArrayType:
			l := len(e.Elts)
			slice := reflect.MakeSlice(typ.(reflect.Type), l, l)
			for i, elem := range e.Elts {
				elemValue, err := scope.InterpretExpr(elem)
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
					key, err := scope.InterpretExpr(eT.Key)
					if err != nil {
						return nil, err
					}
					val, err := scope.InterpretExpr(eT.Value)
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
		x, err := scope.InterpretExpr(e.X)
		if err != nil {
			return nil, err
		}
		y, err := scope.InterpretExpr(e.Y)
		if err != nil {
			return nil, err
		}
		return ComputeBinaryOp(x, y, e.Op)

	case *ast.UnaryExpr:
		x, err := scope.InterpretExpr(e.X)
		if err != nil {
			return nil, err
		}
		return ComputeUnaryOp(x, e.Op)

	case *ast.ArrayType:
		typ, err := scope.InterpretExpr(e.Elt)
		if err != nil {
			return nil, err
		}
		arrType := reflect.SliceOf(typ.(reflect.Type))
		return arrType, nil

	case *ast.MapType:
		keyType, err := scope.InterpretExpr(e.Key)
		if err != nil {
			return nil, err
		}
		valType, err := scope.InterpretExpr(e.Value)
		if err != nil {
			return nil, err
		}
		mapType := reflect.MapOf(keyType.(reflect.Type), valType.(reflect.Type))
		return mapType, nil

	case *ast.ChanType:
		typeI, err := scope.InterpretExpr(e.Value)
		if err != nil {
			return nil, err
		}
		typ, isType := typeI.(reflect.Type)
		if !isType {
			return nil, fmt.Errorf("chan needs to be passed a type not %T", typ)
		}
		return reflect.ChanOf(reflect.BothDir, typ), nil

	case *ast.IndexExpr:
		X, err := scope.InterpretExpr(e.X)
		if err != nil {
			return nil, err
		}
		i, err := scope.InterpretExpr(e.Index)
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
		low, err := scope.InterpretExpr(e.Low)
		if err != nil {
			return nil, err
		}
		high, err := scope.InterpretExpr(e.High)
		if err != nil {
			return nil, err
		}
		X, err := scope.InterpretExpr(e.X)
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
		return scope.InterpretExpr(e.X)

	case *ast.FuncLit:
		return &Func{e}, nil

	default:
		return nil, fmt.Errorf("unknown EXPR %T", e)
	}
}

// InterpretStmt interprets an ast.Stmt and returns the value.
func InterpretStmt(scope *Scope, stmt ast.Stmt) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		var outFinal interface{}
		for _, stmts := range s.List {
			out, err := InterpretStmt(scope, stmts)
			if err != nil {
				return out, err
			}
			outFinal = out
		}
		return outFinal, nil

	case *ast.ReturnStmt:
		results := make([]interface{}, len(s.Results))
		for i, result := range s.Results {
			out, err := scope.InterpretExpr(result)
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
		define := s.Tok == token.DEFINE
		lhs := make([]string, len(s.Lhs))
		for i, id := range s.Lhs {
			lhsIdent, isIdent := id.(*ast.Ident)
			if !isIdent {
				return nil, fmt.Errorf("%#v assignment is not ident", id)
			}
			lhs[i] = lhsIdent.Name
		}
		rhs := make([]interface{}, len(s.Rhs))
		for i, expr := range s.Rhs {
			val, err := scope.InterpretExpr(expr)
			if err != nil {
				return nil, err
			}
			rhs[i] = val
		}
		if len(rhs) != 1 && len(rhs) != len(lhs) {
			return nil, fmt.Errorf("assignment count mismatch: %d = %d", len(lhs), len(rhs))
		}
		if len(rhs) == 1 && reflect.TypeOf(rhs[0]).Kind() == reflect.Slice {
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

	case *ast.ExprStmt:
		return scope.InterpretExpr(s.X)
	default:
		return nil, fmt.Errorf("unknown STMT %#v", s)
	}
}

// StringToType returns the reflect.Type corresponding to the type string provided. Ex: StringToType("int")
func StringToType(str string) (reflect.Type, error) {
	types := map[string]reflect.Type{
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
	val, present := types[str]
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
