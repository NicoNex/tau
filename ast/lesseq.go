package ast

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type LessEq struct {
	l Node
	r Node
}

func NewLessEq(l, r Node) Node {
	return LessEq{l, r}
}

func (l LessEq) Eval(env *obj.Env) obj.Object {
	var (
		left  = l.l.Eval(env)
		right = l.r.Eval(env)
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '<=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '<=' for type %v", right.Type())
	}

	if assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType) {
		le := left.(*obj.Integer).Val()
		ri := right.(*obj.Integer).Val()
		return obj.ParseBool(le <= ri)
	}

	left, right = toFloat(left, right)
	le := left.(*obj.Float).Val()
	ri := right.(*obj.Float).Val()
	return obj.ParseBool(le <= ri)
}

func (l LessEq) String() string {
	return fmt.Sprintf("(%v <= %v)", l.l, l.r)
}

func (l LessEq) Compile(c *compiler.Compiler) (position int, err error) {
	l.r.Compile(c)
	l.l.Compile(c)
	return c.Emit(code.OpGreaterThanEqual), nil
}
