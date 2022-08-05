package obj

type Continue struct{}

func NewContinue() Object {
	return new(Continue)
}

func (c Continue) String() string {
	return "continue"
}

func (c Continue) Type() Type {
	return ContinueType
}
