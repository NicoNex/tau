package obj

type Object interface {
	Type() Type
	String() string
}

type Type int

const (
	NULL Type = iota
	ERROR
	INT
	FLOAT
	BOOL
	STRING
	RETURN
	FUNCTION
	BUILTIN
	LIST
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
	RETURN:   "RETURN",
	FUNCTION: "FUNCTION",
	BUILTIN:  "BUILTIN",
	LIST:     "LIST",
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
