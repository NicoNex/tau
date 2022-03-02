package obj

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var libs = map[string]map[string]interface{}{
	"strings": {
		"compare":        strings.Compare,
		"contains":       strings.Contains,
		"containsAny":    strings.ContainsAny,
		"containsRune":   strings.ContainsRune,
		"count":          strings.Count,
		"equalFold":      strings.EqualFold,
		"fields":         strings.Fields,
		"fieldsFunc":     strings.FieldsFunc,
		"hasPrefix":      strings.HasPrefix,
		"hasSuffix":      strings.HasSuffix,
		"index":          strings.Index,
		"indexAny":       strings.IndexAny,
		"indexByte":      strings.IndexByte,
		"indexFunc":      strings.IndexFunc,
		"indexRune":      strings.IndexRune,
		"join":           strings.Join,
		"lastIndex":      strings.LastIndex,
		"lastIndexAny":   strings.LastIndexAny,
		"lastIndexByte":  strings.LastIndexByte,
		"lastIndexFunc":  strings.LastIndexFunc,
		"map":            strings.Map,
		"repeat":         strings.Repeat,
		"replace":        strings.Replace,
		"replaceAll":     strings.ReplaceAll,
		"split":          strings.Split,
		"splitAfter":     strings.SplitAfter,
		"splitAfterN":    strings.SplitAfterN,
		"splitN":         strings.SplitN,
		"title":          strings.Title,
		"toLower":        strings.ToLower,
		"toLowerSpecial": strings.ToLowerSpecial,
		"toTitle":        strings.ToTitle,
		"toTitleSpecial": strings.ToTitleSpecial,
		"toUpper":        strings.ToUpper,
		"toUpperSpecial": strings.ToUpperSpecial,
		"toValidUTF8":    strings.ToValidUTF8,
		"trim":           strings.Trim,
		"trimFunc":       strings.TrimFunc,
		"trimLeft":       strings.TrimLeft,
		"trimLeftFunc":   strings.TrimLeftFunc,
		"trimPrefix":     strings.TrimPrefix,
		"trimRight":      strings.TrimRight,
		"trimRightFunc":  strings.TrimRightFunc,
		"trimSpace":      strings.TrimSpace,
		"trimSuffix":     strings.TrimSuffix,
	},
}

type Module struct {
	name    string
	methods map[string]interface{}
}

func NewModule(name string) Object {
	lib, ok := libs[name]
	if !ok {
		return NewError("import error: cannot find module with name %q", name)
	}

	return &Module{
		name:    name,
		methods: lib,
	}
}

func (m *Module) Get(n string) (Object, bool) {
	fn, ok := m.methods[n]
	if !ok {
		return NullObj, false
	}

	return Builtin(func(a ...Object) (o Object) {
		defer func() {
			if err := recover(); err != nil {
				o = NewError("%v", err)
			}
		}()

		res := reflect.ValueOf(fn).Call(args(a...))
		return multiplex(res)
	}), true
}

func (m *Module) Set(n string, o Object) Object {
	return NewError("cannot assign to module")
}

func (m Module) Type() Type {
	return ClassType
}

func (m Module) String() string {
	var buf strings.Builder

	l := len(m.methods)
	buf.WriteString("{")
	i := 0
	for m, _ := range m.methods {
		buf.WriteString(fmt.Sprintf("%s: <builtin function>", m))
		if i < l-1 {
			buf.WriteString(", ")
		}
		i++
	}
	buf.WriteString("}")
	return buf.String()
}

func toStringSlicePrimitive(list List) ([]string, error) {
	ret := make([]string, len(list))

	for i, val := range list {
		s, ok := val.(*String)
		if !ok {
			return []string{}, errors.New("list must only contain strings")
		}
		ret[i] = s.Val()
	}
	return ret, nil
}

func toInt64SlicePrimitive(list List) ([]int64, error) {
	ret := make([]int64, len(list))

	for i, val := range list {
		s, ok := val.(*Integer)
		if !ok {
			return []int64{}, errors.New("list must only contain integers")
		}
		ret[i] = s.Val()
	}
	return ret, nil
}

func toFloat64SlicePrimitive(list List) ([]float64, error) {
	ret := make([]float64, len(list))

	for i, val := range list {
		s, ok := val.(*Float)
		if !ok {
			return []float64{}, errors.New("list must only contain floats")
		}
		ret[i] = s.Val()
	}
	return ret, nil
}

func toBoolSlicePrimitive(list List) ([]bool, error) {
	ret := make([]bool, len(list))

	for i, val := range list {
		s, ok := val.(*Boolean)
		if !ok {
			return []bool{}, errors.New("list must only contain floats")
		}
		ret[i] = s.Val()
	}
	return ret, nil
}

func toPrimitive(o Object) interface{} {
	switch casted := o.(type) {
	case *Integer:
		return casted.Val()
	case *String:
		return casted.Val()
	case *Float:
		return casted.Val()
	case *Error:
		return errors.New(casted.Val())
	case *Boolean:
		return casted.Val()
	case List:
		ret := make([]interface{}, len(casted))
		for i, val := range casted {
			ret[i] = toPrimitive(val)
		}
		return ret
	default:
		return nil
	}
}

func args(a ...Object) (args []reflect.Value) {
	args = make([]reflect.Value, len(a))

	for i, arg := range a {
		args[i] = reflect.ValueOf(toPrimitive(arg))
	}
	return
}

func toList(v reflect.Value) (list List) {
	list = make(List, v.Len())

	for i := 0; i < v.Len(); i++ {
		list[i] = toObject(v.Index(i))
	}
	return
}

func toObject(v reflect.Value) Object {
	switch v.Kind() {
	case reflect.String:
		return NewString(v.String())

	case reflect.Bool:
		return NewBoolean(v.Bool())

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return NewInteger(v.Int())

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return NewInteger(int64(v.Uint()))

	case reflect.Float64, reflect.Float32:
		return NewFloat(v.Float())

	case reflect.Slice, reflect.Array:
		return toList(v)

	case reflect.Interface:
		err, ok := v.Interface().(error)
		if ok && err != nil {
			return NewError(err.Error())
		} else if err == nil {
			return NullObj
		}
		fallthrough

	default:
		return NewError("unsupported type %T", v.Interface())
	}
}

func isError(o Object) bool {
	if o.Type() == ErrorType {
		return true
	}
	return false
}

func multiplex(values []reflect.Value) Object {
	switch l := len(values); l {
	case 1:
		return toObject(values[0])

	case 2:
		last := toObject(values[1])
		if isError(last) {
			return last
		}
		return toObject(values[0])

	default:
		last := toObject(values[l-1])
		if isError(last) {
			return last
		}
		list := make(List, l-1)
		for i, v := range values[:l-1] {
			list[i] = toObject(v)
		}
		return list
	}
}
