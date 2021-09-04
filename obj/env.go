package obj

type Env struct {
	outer *Env
	store map[string]*Container
}

func NewEnv() *Env {
	return &Env{nil, make(map[string]*Container)}
}

func NewEnvWrap(e *Env) *Env {
	return &Env{e, make(map[string]*Container)}
}

func (e *Env) Get(n string) (*Container, bool) {
	ret, ok := e.store[n]
	if !ok && e.outer != nil {
		return e.outer.Get(n)
	}
	return ret, ok
}

func (e *Env) Set(n string, o Object) *Container {
	c := &Container{o}
	e.store[n] = c
	return c
}
