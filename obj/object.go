package obj

type Object interface {
	Type() Type
	String() string
}

type Getter interface {
	Object() Object
}

type Setter interface {
	Set(Object) Object
}

type GetSetter interface {
	Object
	Getter
	Setter
}

type MapGetSetter interface {
	Get(string) (Object, bool)
	Set(string, Object) Object
}

type KeyHash struct {
	Type  Type
	Value uint64
}

type Hashable interface {
	KeyHash() KeyHash
}

type setter interface {
	Set(string, Object) Object
}

//go:generate stringer -linecomment -type=Type
type Type int

const (
	NullType     Type = iota // null
	ErrorType                // error
	IntType                  // int
	FloatType                // float
	BoolType                 // bool
	StringType               // string
	ObjectType               // object
	ReturnType               // return
	FunctionType             // function
	ClosureType              // closure
	BuiltinType              // builtin
	ListType                 // list
	MapType                  // map
	ContinueType             // continue
	BreakType                // break
)

var (
	NullObj     = NewNull()
	True        = NewBoolean(true)
	False       = NewBoolean(false)
	ContinueObj = NewContinue()
	BreakObj    = NewBreak()
)

func ParseBool(b bool) Object {
	if b {
		return True
	}
	return False
}

func Unwrap(o Object) Object {
	if g, ok := o.(Getter); ok {
		return g.Object()
	}
	return o
}
