package main

import "github.com/NicoNex/tau/internal/obj"

// builtinDoc is a builtin's signature and the one line that says what it
// does. The names come from obj.Builtins so the list cannot drift from the
// runtime; only the prose lives here.
type builtinDoc struct {
	signature string
	summary   string
}

var builtinDocs = map[string]builtinDoc{
	"len":     {"len(x)", "The number of elements in a string, list, map or bytes value."},
	"println": {"println(...)", "Writes its arguments separated by spaces, then a newline."},
	"print":   {"print(...)", "Writes its arguments separated by spaces, without a newline."},
	"input":   {"input(prompt...)", "Reads one line from standard input and returns it without the newline."},
	"string":  {"string(x)", "The string representation of any value."},
	"error":   {"error(msg)", "Builds an error value carrying msg."},
	"type":    {"type(x)", "The name of the type of x, as a string."},
	"int":     {"int(x)", "Converts a number or string to an integer."},
	"float":   {"float(x)", "Converts a number or string to a float."},
	"exit":    {"exit(code...)", "Stops the program with the given status, zero by default."},
	"append":  {"append(list, ...)", "Returns the list with the given elements added at the end."},
	"new":     {"new()", "A new empty object, the value a module builds itself from."},
	"failed":  {"failed(x)", "Reports whether x is an error value."},
	"plugin":  {"plugin(path)", "Loads a shared object and returns the object exposing its symbols."},
	"pipe":    {"pipe(capacity...)", "A channel that goroutines send to and receive from."},
	"send":    {"send(pipe, value)", "Sends a value on a pipe, blocking until it is taken."},
	"recv":    {"recv(pipe)", "Receives the next value from a pipe, blocking until one arrives."},
	"close":   {"close(pipe)", "Closes a pipe, so that no further value can be sent on it."},
	"hex":     {"hex(n)", "The hexadecimal representation of an integer, as a string."},
	"oct":     {"oct(n)", "The octal representation of an integer, as a string."},
	"bin":     {"bin(n)", "The binary representation of an integer, as a string."},
	"slice":   {"slice(x, start, end)", "The part of a string, list or bytes value between start and end."},
	"keys":    {"keys(m)", "The list of the keys of a map."},
	"delete":  {"delete(m, key)", "Removes a key from a map."},
	"bytes":   {"bytes(x)", "The bytes of a string, or a bytes value of the given length."},
	"import":  {"import(path)", "Loads a tau module and returns the object holding its exported names."},
}

// builtins is the runtime's own list, in a form the completion can range
// over. `import` is a keyword to the lexer but reads as a builtin call, so
// it is offered alongside them.
var builtins = func() []string {
	out := make([]string, 0, len(obj.Builtins)+1)
	out = append(out, obj.Builtins[:]...)
	return append(out, "import")
}()

// keywords are the words the lexer reserves.
var keywords = []string{
	"fn", "if", "else", "for", "return", "break", "continue",
	"true", "false", "null", "import", "tau",
}

var keywordSet = func() map[string]bool {
	m := make(map[string]bool, len(keywords))
	for _, k := range keywords {
		m[k] = true
	}
	return m
}()

func isBuiltin(name string) bool {
	_, ok := builtinDocs[name]
	return ok
}
