package obj

import "path/filepath"

type Moduler interface {
	Module() *Module
}

type Env struct {
	Outer *Env
	Store Store
	file  string
	dir   string
}

func NewEnv(path string) *Env {
	dir, file := filepath.Split(path)

	return &Env{
		Outer: nil,
		Store: NewStore(),
		file:  file,
		dir:   dir,
	}
}

func NewEnvWrap(e *Env) *Env {
	return &Env{
		Outer: e,
		Store: NewStore(),
		file:  e.file,
		dir:   e.dir,
	}
}

func (e *Env) File() string {
	return e.file
}

func (e *Env) Dir() string {
	return e.dir
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
	return e.Store.Module()
}