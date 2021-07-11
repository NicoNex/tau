package ast

import (
	"fmt"
	"github.com/NicoNex/tau/obj"
	"strings"
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
		k := key.Eval(env)
		if isError(k) {
			return k
		}

		h, ok := k.(obj.Hashable)
		if !ok {
			return obj.NewError("invalid map key type %v", k.Type())
		}

		v := val.Eval(env)
		if isError(v) {
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
		buf.WriteString(fmt.Sprintf("%v: %v", k, v))

		if i < len(m) {
			buf.WriteString(", ")
		}
		i += 1
	}
	buf.WriteString("}")
	return buf.String()
}
