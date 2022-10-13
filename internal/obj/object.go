package obj

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

var (
	Input             string
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

type TauError struct {
	pos int
	msg string
}

func Errorf(pos int, f string, a ...any) TauError {
	line := extractLine(Input, pos)

	return TauError{
		pos: pos,
		msg: fmt.Sprintf(
			"Error at line %d:\n    %s\n    %s\n%s",
			line.number,
			line,
			arrow(line.pos),
			fmt.Sprintf(f, a...),
		),
	}
}

func (t TauError) Error() string {
	return t.msg
}

func (t TauError) Pos() int {
	return t.pos
}

type line struct {
	start  int
	end    int
	number int
	pos    int
	str    string
}

func (l line) String() string {
	return l.str
}

func extractLine(input string, pos int) line {
	s, e := startLine(input, pos), endLine(input, pos)
	l := input[s:e]
	str := strings.TrimLeft(l, " \t")

	return line{
		start:  s,
		end:    e,
		number: lineNo(input, pos),
		pos:    pos - s - (len(l) - len(str)),
		str:    str,
	}
}

func startLine(s string, pos int) (beg int) {
	for i := pos - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return i + 1
		}
	}
	return
}

func endLine(s string, pos int) (end int) {
	end = len(s)
	for i := pos; i < len(s); i++ {
		if s[i] == '\n' {
			return i
		}
	}
	return
}

func lineNo(s string, pos int) int {
	var cnt = 1

	for _, b := range s[:pos] {
		if b == '\n' {
			cnt++
		}
	}

	return cnt
}

func arrow(pos int) string {
	var s = make([]byte, pos+1)

	for i := range s {
		if i == pos {
			s[i] = '^'
		} else {
			s[i] = ' '
		}
	}
	return string(s)
}
