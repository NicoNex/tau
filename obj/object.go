package obj

type Object interface {
	Type() Type
	String() string
}

type KeyHash struct {
	Type  Type
	Value uint64
}

type Hashable interface {
	KeyHash() KeyHash
}

type Type int

const (
	NULL Type = iota
	ERROR
	INT
	FLOAT
	BOOL
	STRING
	CLASS
	RETURN
	FUNCTION
	BUILTIN
	LIST
	MAP
)

var (
	NullObj = NewNull()
	True    = NewBoolean(true)
	False   = NewBoolean(false)
)

var typrepr = map[Type]string{
	NULL:     "NULL",
	ERROR:    "ERROR",
	INT:      "INTEGER",
	FLOAT:    "FLOAT",
	BOOL:     "BOOLEAN",
	STRING:   "STRING",
	CLASS:    "CLASS",
	RETURN:   "RETURN",
	FUNCTION: "FUNCTION",
	BUILTIN:  "BUILTIN",
	LIST:     "LIST",
	MAP:      "MAP",
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
