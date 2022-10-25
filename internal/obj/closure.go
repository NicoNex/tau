package obj

import "fmt"

type Closure struct {
	Fn   *CompiledFunction
	Free []Object
}

func NewClosure(fn *CompiledFunction, free []Object) *Closure {
	return &Closure{Fn: fn, Free: free}
}

func (c *Closure) String() string {
	return fmt.Sprintf("closure[%p]", c)
}

func (c *Closure) Type() Type {
	return ClosureType
}
