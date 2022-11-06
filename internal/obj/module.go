package obj

import "strings"

type Module struct {
	Exported   Store
	Unexported Store
}

func NewModule() *Module {
	return &Module{
		Exported:   NewStore(),
		Unexported: NewStore(),
	}
}

func (m *Module) Get(n string) (Object, bool) {
	ret, ok := m.Exported[n]
	return ret, ok
}

func (m *Module) Set(n string, o Object) Object {
	if _, ok := m.Unexported[n]; ok {
		return NewError("cannot assign to unexported field")
	}
	m.Exported[n] = o
	return o
}

func (m Module) Type() Type {
	return ObjectType
}

func (m Module) String() string {
	var buf strings.Builder

	buf.WriteRune('{')
	i := 0
	for k, v := range m.Exported {
		buf.WriteString(k)
		buf.WriteString(": ")
		if s, ok := v.(String); ok {
			buf.WriteString(s.Quoted())
		} else {
			buf.WriteString(v.String())
		}
		if i < len(m.Exported)-1 {
			buf.WriteString(", ")
		}

		i++
	}

	buf.WriteRune('}')
	return buf.String()
}
