package obj

type Object interface {
	Type() Type
	String() string
}

type MapGetter interface {
	Get(string) (Object, bool)
}

type MapSetter interface {
	Set(string, Object) Object
}

type MapGetSetter interface {
	MapGetter
	MapSetter
}

type Getter interface {
	Object() Object
}

type Setter interface {
	Set(Object) Object
}

type GetSetter interface {
	Getter
	Setter
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

type Type int

const (
	NullType Type = iota
	ErrorType
	IntType
	FloatType
	BoolType
	StringType
	ClassType
	ReturnType
	FunctionType
	ClosureType
	BuiltinType
	ListType
	MapType
	ContinueType
	BreakType
)

var (
	NullObj     = NewNull()
	True        = NewBoolean(true)
	False       = NewBoolean(false)
	ContinueObj = NewContinue()
	BreakObj    = NewBreak()
)

var typrepr = map[Type]string{
	NullType:     "null",
	ErrorType:    "error",
	IntType:      "int",
	FloatType:    "float",
	BoolType:     "bool",
	StringType:   "string",
	ClassType:    "class",
	ReturnType:   "return",
	FunctionType: "function",
	ClosureType:  "closure",
	BuiltinType:  "builtin",
	ListType:     "list",
	MapType:      "map",
	ContinueType: "continue",
	BreakType:    "break",
}

func (t Type) String() string {
	return typrepr[t]
}

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
