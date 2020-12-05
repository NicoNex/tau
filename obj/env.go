package obj

type Env struct {
	outer *Env
	store map[string]Object
}

func NewEnv() *Env {
	return &Env{nil, make(map[string]Object)}
}

func NewEnvWrap(e *Env) *Env {
	return &Env{e, make(map[string]Object)}
}

func (e *Env) Get(n string) (Object, bool) {
	ret, ok := e.store[n]
	if !ok && e.outer != nil {
		return e.outer.Get(n)
	}
	return ret, ok
}

func (e *Env) Set(n string, o Object) Object {
	e.store[n] = o
	return o
}
