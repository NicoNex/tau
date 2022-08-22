package obj

import (
	"unicode"
	"unicode/utf8"
)

type Moduler interface {
	Module() *Module
}

type Env struct {
	Outer *Env
	Store map[string]Object
}

func NewEnv() *Env {
	return &Env{nil, make(map[string]Object)}
}

func NewEnvWrap(e *Env) *Env {
	return &Env{e, make(map[string]Object)}
}

func (e *Env) Get(n string) (Object, bool) {
	ret, ok := e.Store[n]
	if !ok && e.Outer != nil {
		return e.Outer.Get(n)
	}
	return ret, ok
}

func (e *Env) Set(n string, o Object) Object {
	e.Store[n] = o
	return o
}

func (e *Env) Module() *Module {
	m := NewModule()

	for n, o := range e.Store {
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
