package item

import "fmt"

type Item struct {
	Typ Type
	Val string
	Pos int
}

func (i Item) Is(t Type) bool {
	return i.Typ == t
}

func (i Item) String() string {
	if i.Is(ERROR) {
		return i.Val
	}
	if len(i.Val) > 10 {
		return fmt.Sprintf("%.10q...", i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}
