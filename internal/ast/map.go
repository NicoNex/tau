package ast

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Map struct {
	m   [][2]Node
	pos int
}

func NewMap(pos int, pairs ...[2]Node) Node {
	return Map{
		m:   pairs,
		pos: pos,
	}
}

func (m Map) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.Index: not a constant expression")
}

func (m Map) String() string {
	var (
		buf strings.Builder
		i   = 1
	)

	buf.WriteString("{")
	for _, pair := range m.m {
		var (
			k   = pair[0]
			v   = pair[1]
			key string
			val string
		)

		if s, ok := k.(String); ok {
			key = s.Quoted()
		} else {
			key = k.String()
		}

		if s, ok := v.(String); ok {
			val = s.Quoted()
		} else {
			val = v.String()
		}

		buf.WriteString(fmt.Sprintf("%s: %s", key, val))

		if i < len(m.m) {
			buf.WriteString(", ")
		}
		i += 1
	}
	buf.WriteString("}")
	return buf.String()
}

func (m Map) Compile(c *compiler.Compiler) (position int, err error) {
	sort.Slice(m.m, func(i, j int) bool {
		return m.m[i][0].String() < m.m[j][0].String()
	})

	for _, pair := range m.m {
		if position, err = pair[0].Compile(c); err != nil {
			return
		}

		if position, err = pair[1].Compile(c); err != nil {
			return
		}
	}

	position = c.Emit(code.OpMap, len(m.m)*2)
	c.Bookmark(m.pos)
	return
}

func (m Map) IsConstExpression() bool {
	return false
}
