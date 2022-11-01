package obj

import (
	"unicode"
	"unicode/utf8"
)

type Store map[string]Object

func NewStore() Store {
	return make(Store)
}

func (s Store) Get(n string) (Object, bool) {
	ret, ok := s[n]
	return ret, ok
}

func (s Store) Set(n string, o Object) Object {
	s[n] = o
	return o
}

func (s Store) Module() *Module {
	m := NewModule()

	for n, o := range s {
		if env, ok := o.(Moduler); ok {
			o = env.Module()
		}

		if isExported(n) {
			m.Exported[n] = o
		} else {
			m.Unexported[n] = o
		}
	}

	return m
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}