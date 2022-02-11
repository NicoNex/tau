package obj

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
