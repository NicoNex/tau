package obj

import "fmt"

type Closure struct {
	Fn   *Function
	Free []Object
}

func NewClosure(fn *Function, free []Object) *Closure {
	return &Closure{Fn: fn, Free: free}
}

func (c *Closure) String() string {
	return fmt.Sprintf("closure[%p]", c)
}

func (c *Closure) Type() Type {
	return ClosureType
}
