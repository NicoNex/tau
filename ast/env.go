package ast

// type Env struct {
// 	store map[string]Object
// 	outer *Env
// }

// func NewEnv() *Env {
// 	return &Env{store: make(map[string]Object), outer: nil}
// }

// func NewEnclosedEnv(outer *Env) *Env {
// 	return &Env{store: make(map[string]Object), outer: outer}
// }

// func (e *Env) Get(name string) (Object, bool) {
// 	ret, ok := e.store[name]
// 	if !ok && e.outer != nil {
// 		ret, ok = e.outer.Get(name)
// 	}
// 	return ret, ok
// }

// func (e *Env) Set(name string, val Object) Object {
// 	e.store[name] = val
// 	return val
// }
