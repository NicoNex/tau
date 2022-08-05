package ast

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Map map[Node]Node

func NewMap(pairs ...[2]Node) Node {
	var m = make(map[Node]Node)

	for _, p := range pairs {
		m[p[0]] = p[1]
	}
	return Map(m)
}

func (m Map) Eval(env *obj.Env) obj.Object {
	var ret = obj.NewMap()

	for key, val := range m {
		k := obj.Unwrap(key.Eval(env))
		if takesPrecedence(k) {
			return k
		}

		h, ok := k.(obj.Hashable)
		if !ok {
			return obj.NewError("invalid map key type %v", k.Type())
		}

		v := obj.Unwrap(val.Eval(env))
		if takesPrecedence(v) {
			return v
		}

		ret.Set(h.KeyHash(), obj.MapPair{Key: k, Value: v})
	}

	return ret
}

func (m Map) String() string {
	var buf strings.Builder
	var i = 1

	buf.WriteString("{")
	for k, v := range m {
		var key, val string

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

		if i < len(m) {
			buf.WriteString(", ")
		}
		i += 1
	}
	buf.WriteString("}")
	return buf.String()
}

func (m Map) Compile(c *compiler.Compiler) (position int, err error) {
	var keys []Node

	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for _, k := range keys {
		if position, err = k.Compile(c); err != nil {
			return
		}

		if position, err = m[k].Compile(c); err != nil {
			return
		}
	}

	return c.Emit(code.OpMap, len(m)*2), nil
}
