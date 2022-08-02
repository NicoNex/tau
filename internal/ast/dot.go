package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Dot struct {
	l Node
	r Node
}

func NewDot(l, r Node) Node {
	return Dot{l, r}
}

func (d Dot) Eval(env *obj.Env) obj.Object {
	var left = obj.Unwrap(d.l.Eval(env))

	if takesPrecedence(left) {
		return left
	}

	switch l := left.(type) {
	case obj.Class:
		return &obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				return l.Get(d.r.String())
			},

			SetFunc: func(o obj.Object) obj.Object {
				return l.Set(d.r.String(), o)
			},
		}

	case obj.GetSetter:
		return l

	default:
		return obj.NewError("%v object has no attribute %s", left.Type(), d.r)
	}
}

func (d Dot) String() string {
	return fmt.Sprintf("%v.%v", d.l, d.r)
}

func (d Dot) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = d.l.Compile(c); err != nil {
		return
	}
	if _, ok := d.r.(Identifier); !ok {
		return position, fmt.Errorf("expected identifier with dot operator, got %T", d.r)
	}
	position = c.Emit(code.OpConstant, c.AddConstant(obj.NewString(d.r.String())))
	return c.Emit(code.OpDot), nil
}
