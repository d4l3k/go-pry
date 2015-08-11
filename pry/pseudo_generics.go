package pry

import (
	"fmt"
	"go/token"
	"reflect"
)

// ComputeBinaryOp executes the corresponding binary operation (+, -, etc) on two interfaces.
func ComputeBinaryOp(xI, yI interface{}, op token.Token) (interface{}, error) {
	typeX := reflect.TypeOf(xI)
	typeY := reflect.TypeOf(yI)
	if typeX == typeY {
		switch xI.(type) {
		case string:
			x := xI.(string)
			y := yI.(string)
			switch op {
			case token.ADD:
				return x + y, nil
			}
		case int:
			x := xI.(int)
			y := yI.(int)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case int8:
			x := xI.(int8)
			y := yI.(int8)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case int16:
			x := xI.(int16)
			y := yI.(int16)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case int32:
			x := xI.(int32)
			y := yI.(int32)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case int64:
			x := xI.(int64)
			y := yI.(int64)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uint:
			x := xI.(uint)
			y := yI.(uint)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uint8:
			x := xI.(uint8)
			y := yI.(uint8)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uint16:
			x := xI.(uint16)
			y := yI.(uint16)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uint32:
			x := xI.(uint32)
			y := yI.(uint32)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uint64:
			x := xI.(uint64)
			y := yI.(uint64)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case uintptr:
			x := xI.(uintptr)
			y := yI.(uintptr)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.REM:
				return x % y, nil
			case token.AND:
				return x & y, nil
			case token.OR:
				return x | y, nil
			case token.XOR:
				return x ^ y, nil
			case token.AND_NOT:
				return x &^ y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case complex64:
			x := xI.(complex64)
			y := yI.(complex64)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			}
		case complex128:
			x := xI.(complex128)
			y := yI.(complex128)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			}
		case float32:
			x := xI.(float32)
			y := yI.(float32)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case float64:
			x := xI.(float64)
			y := yI.(float64)
			switch op {
			case token.ADD:
				return x + y, nil
			case token.SUB:
				return x - y, nil
			case token.MUL:
				return x * y, nil
			case token.QUO:
				return x / y, nil
			case token.LSS:
				return x < y, nil
			case token.GTR:
				return x > y, nil
			case token.LEQ:
				return x <= y, nil
			case token.GEQ:
				return x >= y, nil
			}
		case bool:
			x := xI.(bool)
			y := yI.(bool)
			switch op {
			// Bool
			case token.LAND:
				return x && y, nil
			case token.LOR:
				return x || y, nil
			}
		}
	}
	yUint, isUint := yI.(uint64)
	if !isUint {
		isUint = true
		switch yV := yI.(type) {
		case int:
			yUint = uint64(yV)
		case int8:
			yUint = uint64(yV)
		case int16:
			yUint = uint64(yV)
		case int32:
			yUint = uint64(yV)
		case int64:
			yUint = uint64(yV)
		case uint8:
			yUint = uint64(yV)
		case uint16:
			yUint = uint64(yV)
		case uint32:
			yUint = uint64(yV)
		case uint64:
			yUint = uint64(yV)
		case float32:
			yUint = uint64(yV)
		case float64:
			yUint = uint64(yV)
		default:
			isUint = false
		}
	}
	if isUint {
		switch xI.(type) {
		case int:
			x := xI.(int)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case int8:
			x := xI.(int8)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case int16:
			x := xI.(int16)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case int32:
			x := xI.(int32)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case int64:
			x := xI.(int64)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uint:
			x := xI.(uint)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uint8:
			x := xI.(uint8)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uint16:
			x := xI.(uint16)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uint32:
			x := xI.(uint32)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uint64:
			x := xI.(uint64)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		case uintptr:
			x := xI.(uintptr)
			switch op {
			// Num, uint
			case token.SHL:
				return x << yUint, nil
			case token.SHR:
				return x >> yUint, nil
			}
		}
	}
	// Anything
	switch op {
	case token.EQL:
		return xI == yI, nil
	case token.NEQ:
		return xI != yI, nil
	}
	return nil, fmt.Errorf("unknown operation %#v between %#v and %#v", op, xI, yI)
}

// ComputeUnaryOp computes the corresponding unary (+x, -x) operation on an interface.
func ComputeUnaryOp(xI interface{}, op token.Token) (interface{}, error) {
	switch xI.(type) {
	case bool:
		x := xI.(bool)
		switch op {
		case token.NOT:
			return !x, nil
		}
	case int:
		x := xI.(int)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case int8:
		x := xI.(int8)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case int16:
		x := xI.(int16)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case int32:
		x := xI.(int32)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case int64:
		x := xI.(int64)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uint:
		x := xI.(uint)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uint8:
		x := xI.(uint8)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uint16:
		x := xI.(uint16)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uint32:
		x := xI.(uint32)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uint64:
		x := xI.(uint64)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case uintptr:
		x := xI.(uintptr)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case float32:
		x := xI.(float32)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case float64:
		x := xI.(float64)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case complex64:
		x := xI.(complex64)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	case complex128:
		x := xI.(complex128)
		switch op {
		case token.ADD:
			return +x, nil
		case token.SUB:
			return -x, nil
		}
	}
	return nil, fmt.Errorf("unknown unary operation %#v on %#v", op, xI)
}
