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
	ARRAY
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
	ARRAY:    "ARRAY",
}

func (t Type) String() string {
	return typrepr[t]
}
