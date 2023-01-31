package obj

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
	BytesType                // bytes
	ObjectType               // object
	FunctionType             // function
	ClosureType              // closure
	BuiltinType              // builtin
	ListType                 // list
	MapType                  // map
	PipeType                 // pipe
	ReturnType               // return
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

func AssertTypes(o Object, types ...Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func IsPrimitive(o Object) bool {
	return AssertTypes(o, BoolType, ErrorType, FloatType, IntType, StringType)
}

func ToFloat(l, r Object) (Object, Object) {
	if i, ok := l.(Integer); ok {
		l = Float(i)
	}
	if i, ok := r.(Integer); ok {
		r = Float(i)
	}
	return l, r
}

func IsTruthy(o Object) bool {
	switch val := o.(type) {
	case *Boolean:
		return o == True
	case Integer:
		return val != 0
	case Float:
		return val != 0
	case *Null:
		return false
	default:
		return true
	}
}

var (
	ErrFileNotFound   = errors.New("file not found")
	ErrNoFileProvided = errors.New("no file provided")
)

func ImportLookup(taupath string) (string, error) {
	dir, file := filepath.Split(taupath)

	if file == "" {
		return "", ErrNoFileProvided
	}

	if filepath.Ext(file) == "" {
		file += ".tau"
	}

	path := filepath.Join(dir, file)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join("/lib", "tau", dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("%s: %w", path, err)
		}
	}

	return path, nil
}
