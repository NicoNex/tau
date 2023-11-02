package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Import struct {
	name  Node
	parse parseFn
	pos   int
}

func NewImport(name Node, parse parseFn, pos int) Node {
	return &Import{
		name:  name,
		parse: parse,
		pos:   pos,
	}
}

func (i Import) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.Import: not a constant expression")
}

func (i Import) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.name.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpLoadModule)
	c.Bookmark(i.pos)
	return
}

func (i Import) String() string {
	return fmt.Sprintf("import(%q)", i.name.String())
}

func (i Import) IsConstExpression() bool {
	return false
}
