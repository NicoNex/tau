package obj

type Pipe chan Object

func NewPipe() Object {
	return Pipe(make(chan Object))
}

func NewPipeBuffered(n int) Object {
	return Pipe(make(chan Object, n))
}

func (p Pipe) Type() Type {
	return PipeType
}

func (p Pipe) String() string {
	return "<pipe>"
}
