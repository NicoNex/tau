package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Less struct {
	l Node
	r Node
}

func NewLess(l, r Node) Node {
	return Less{l, r}
}

func (l Less) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(l.l.Eval(env))
		right = obj.Unwrap(l.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '<' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '<' for type %v", right.Type())
	}

	if assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType) {
		le := left.(*obj.Integer).Val()
		ri := right.(*obj.Integer).Val()
		return obj.ParseBool(le < ri)
	}

	left, right = toFloat(left, right)
	le := left.(*obj.Float).Val()
	ri := right.(*obj.Float).Val()
	return obj.ParseBool(le < ri)
}

func (l Less) String() string {
	return fmt.Sprintf("(%v < %v)", l.l, l.r)
}

func (l Less) Compile(c *compiler.Compiler) (position int, err error) {
	// the order of the compilation of the operands is inverted because we reuse
	// the code.OpGreaterThan OpCode.
	if position, err = l.r.Compile(c); err != nil {
		return
	}
	if position, err = l.l.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpGreaterThan), nil
}
