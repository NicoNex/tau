package obj

import (
	"io"
	"os"
)

func NewFile(path string, flag int) (Object, error) {
	var ret = Class{NewStore()}

	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return nil, err
	}

	ret.Set("Read", Builtin(func(args ...Object) Object {
		if l := len(args); l != 0 {
			return NewError("Read: wrong number of arguments, expected 0 got %d", l)
		}

		b, err := io.ReadAll(f)
		if err != nil {
			return NewError("Read: %v", err)
		}

		return Bytes(b)
	}))

	ret.Set("Write", Builtin(func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("Write: wrong number of arguments, expected 1 got %d", l)
		}

		switch a := Unwrap(args[0]).(type) {
		case String:
			i, err := io.WriteString(f, string(a))
			if err != nil {
				return NewError("Write: %v", err)
			}
			return Integer(i)

		case Bytes:
			i, err := f.Write([]byte(a))
			if err != nil {
				return NewError("Write: %v", err)
			}
			return Integer(i)

		default:
			return NewError("Write: wrong argument type, expected string or bytes, got %s instead", args[0].Type())
		}
	}))

	ret.Set("Sync", Builtin(func(args ...Object) Object {
		if l := len(args); l != 0 {
			return NewError("Sync: wrong number of arguments, expected 0 got %d", l)
		}

		if err := f.Sync(); err != nil {
			return NewError("Sync: %v", err)
		}
		return NullObj
	}))

	ret.Set("Truncate", Builtin(func(args ...Object) Object {
		if len(args) != 1 {
			return NewError("Truncate: wrong number of arguments, expected 1, got %d", len(args))
		}

		size, ok := Unwrap(args[0]).(Integer)
		if !ok {
			return NewError("Truncate: wrong argument type, expected int, got %s", args[0].Type())
		}

		if err := f.Truncate(int64(size)); err != nil {
			return NewError("Truncate: %v", err)
		}
		return NullObj
	}))

	ret.Set("Close", Builtin(func(args ...Object) Object {
		if l := len(args); l != 0 {
			return NewError("Close: wrong number of arguments, expected 0 got %d", l)
		}

		if err := f.Close(); err != nil {
			return NewError("Close: %v", err)
		}
		return NullObj
	}))

	return ret, nil
}