package obj

type Break struct{}

func NewBreak() Object {
	return new(Break)
}

func (b Break) String() string {
	return "break"
}

func (b Break) Type() Type {
	return BreakType
}
