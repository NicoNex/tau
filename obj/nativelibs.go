package obj

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

var NativeLibs = map[string]map[string]interface{}{
	"strings": {
		"Compare":        strings.Compare,
		"Contains":       strings.Contains,
		"ContainsAny":    strings.ContainsAny,
		"ContainsRune":   strings.ContainsRune,
		"Count":          strings.Count,
		"EqualFold":      strings.EqualFold,
		"Fields":         strings.Fields,
		"FieldsFunc":     strings.FieldsFunc,
		"HasPrefix":      strings.HasPrefix,
		"HasSuffix":      strings.HasSuffix,
		"Index":          strings.Index,
		"IndexAny":       strings.IndexAny,
		"IndexByte":      strings.IndexByte,
		"IndexFunc":      strings.IndexFunc,
		"IndexRune":      strings.IndexRune,
		"Join":           strings.Join,
		"LastIndex":      strings.LastIndex,
		"LastIndexAny":   strings.LastIndexAny,
		"LastIndexByte":  strings.LastIndexByte,
		"LastIndexFunc":  strings.LastIndexFunc,
		"Map":            strings.Map,
		"Repeat":         strings.Repeat,
		"Replace":        strings.Replace,
		"ReplaceAll":     strings.ReplaceAll,
		"Split":          strings.Split,
		"SplitAfter":     strings.SplitAfter,
		"splitAfterN":    strings.SplitAfterN,
		"SplitN":         strings.SplitN,
		"Title":          strings.Title,
		"ToLower":        strings.ToLower,
		"ToLowerSpecial": strings.ToLowerSpecial,
		"ToTitle":        strings.ToTitle,
		"ToTitleSpecial": strings.ToTitleSpecial,
		"ToUpper":        strings.ToUpper,
		"ToUpperSpecial": strings.ToUpperSpecial,
		"ToValidUTF8":    strings.ToValidUTF8,
		"Trim":           strings.Trim,
		"TrimFunc":       strings.TrimFunc,
		"TrimLeft":       strings.TrimLeft,
		"TrimLeftFunc":   strings.TrimLeftFunc,
		"TrimPrefix":     strings.TrimPrefix,
		"TrimRight":      strings.TrimRight,
		"TrimRightFunc":  strings.TrimRightFunc,
		"TrimSpace":      strings.TrimSpace,
		"TrimSuffix":     strings.TrimSuffix,
	},
	"regexp": {
		"MatchString":      regexp.MatchString,
		"QuoteMeta":        regexp.QuoteMeta,
		"Compile":          regexp.Compile,
		"CompilePOSIX":     regexp.CompilePOSIX,
		"MustCompile":      regexp.MustCompile,
		"MustCompilePOSIX": regexp.MustCompilePOSIX,
	},
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

	case reflect.Struct, reflect.Ptr:
		return NewNativeStruct(v.Interface())

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
