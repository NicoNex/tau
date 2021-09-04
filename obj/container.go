package obj

type Container struct {
	o Object
}

func (c Container) Type() Type {
	return c.o.Type()
}

func (c Container) String() string {
	return c.o.String()
}

func (c *Container) Set(o Object) Object {
	c.o = o
	return o
}

func (c Container) Object() Object {
	return c.o
}
