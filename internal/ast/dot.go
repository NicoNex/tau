package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Dot struct {
	l   Node
	r   Node
	pos int
}

func NewDot(l, r Node, pos int) Node {
	return Dot{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (d Dot) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.Dot: not a constant expression")
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
	position = c.Emit(code.OpDot)
	c.Bookmark(d.pos)
	return
}

// CompileDefine assumes the dot operation is for defining a value.
func (d Dot) CompileDefine(c *compiler.Compiler) (position int, err error) {
	if position, err = d.l.Compile(c); err != nil {
		return
	}
	if _, ok := d.r.(Identifier); !ok {
		return position, fmt.Errorf("expected identifier with dot operator, got %T", d.r)
	}
	position = c.Emit(code.OpConstant, c.AddConstant(obj.NewString(d.r.String())))
	c.Bookmark(d.pos)
	return
}

func (d Dot) IsConstExpression() bool {
	return false
}
