package obj

import (
	"fmt"
	"reflect"
)

func toValue(t reflect.Type, o Object) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Bool:
		p, ok := o.(*Boolean)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected bool but %v provided", o.Type())
		}
		return reflect.ValueOf(p == True), nil

	case reflect.Int:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected int but %v provided", o.Type())
		}
		return reflect.ValueOf(int(*i)), nil

	case reflect.Int8:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected int8 but %v provided", o.Type())
		}
		return reflect.ValueOf(int8(*i)), nil

	case reflect.Int16:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected int16 but %v provided", o.Type())
		}
		return reflect.ValueOf(int16(*i)), nil

	case reflect.Int32:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected int32 but %v provided", o.Type())
		}
		return reflect.ValueOf(int32(*i)), nil

	case reflect.Int64:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected int64 but %v provided", o.Type())
		}
		return reflect.ValueOf(int64(*i)), nil

	case reflect.Uint:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected uint but %v provided", o.Type())
		}
		return reflect.ValueOf(uint(*i)), nil

	case reflect.Uint8:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected uint8 but %v provided", o.Type())
		}
		return reflect.ValueOf(uint8(*i)), nil

	case reflect.Uint16:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected uint16 but %v provided", o.Type())
		}
		return reflect.ValueOf(uint16(*i)), nil

	case reflect.Uint32:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected uint32 but %v provided", o.Type())
		}
		return reflect.ValueOf(uint32(*i)), nil

	case reflect.Uint64:
		i, ok := o.(*Integer)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected uint64 but %v provided", o.Type())
		}
		return reflect.ValueOf(uint64(*i)), nil

	case reflect.Uintptr:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'uintptr'")

	case reflect.Float32:
		f, ok := o.(*Float)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected float32 but %v provided", o.Type())
		}
		return reflect.ValueOf(float32(*f)), nil

	case reflect.Float64:
		f, ok := o.(*Float)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected float64 but %v provided", o.Type())
		}
		return reflect.ValueOf(float64(*f)), nil

	case reflect.Complex64:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'complex64'")

	case reflect.Complex128:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'complex128'")

	case reflect.Chan:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'chan'")

	case reflect.Func:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'func'")

	case reflect.Interface:
		switch o.(type) {
		case *Null:
			return reflect.Zero(t), nil

		default:
			return reflect.Zero(t), fmt.Errorf("unsupported type 'interface'")
		}

	case reflect.Pointer:
		switch o.(type) {
		case *Null:
			return reflect.Zero(t), nil

		default:
			ret := reflect.New(t.Elem())
			v, err := toValue(ret.Elem().Type(), o)
			if err != nil {
				return reflect.Zero(t), err
			}
			ret.Set(v)
			return ret, nil
		}

	case reflect.Array:
		l, ok := o.(List)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected list but %v provided", o.Type())
		}

		if len(l) != t.Len() {
			return reflect.Zero(t), fmt.Errorf("length mismatch: expected %d, got %d", t.Len(), len(l))
		}

		innerType := t.Elem()
		arrayType := reflect.ArrayOf(t.Len(), innerType)
		array := reflect.New(arrayType).Elem()
		for i, e := range l {
			v, err := toValue(innerType, e)
			if err != nil {
				return reflect.Zero(t), fmt.Errorf("expected %s but %v provided", innerType.String(), e.Type())
			}
			array.Index(i).Set(v)
		}

		return array, nil

	case reflect.Slice:
		l, ok := o.(List)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected list but %v provided", o.Type())
		}
		innerType := t.Elem()
		slice := reflect.MakeSlice(t, len(l), cap(l))
		for i, e := range l {
			v, err := toValue(innerType, e)
			if err != nil {
				return reflect.Zero(t), fmt.Errorf("expected %s but %v provided", innerType.String(), e.Type())
			}
			slice.Index(i).Set(v)
		}

		return slice, nil

	case reflect.Map:
		m, ok := o.(Map)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected map but %v provided", o.Type())
		}

		keyType := t.Key()
		valType := t.Elem()
		retMap := reflect.MakeMap(t)

		for _, pair := range m {
			key, err := toValue(keyType, pair.Key)
			if err != nil {
				return reflect.Zero(t), fmt.Errorf("expected %s but %v provided", keyType.String(), pair.Key.Type())
			}

			val, err := toValue(valType, pair.Value)
			if err != nil {
				return reflect.Zero(t), fmt.Errorf("expected %s but %v provided", valType.String(), pair.Value.Type())
			}

			retMap.SetMapIndex(key, val)
		}

		return retMap, nil

	case reflect.String:
		s, ok := o.(*String)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected string but %v provided", o.Type())
		}
		return reflect.ValueOf(string(*s)), nil

	case reflect.Struct:
		object, ok := o.(Class)
		if !ok {
			return reflect.Zero(t), fmt.Errorf("expected object but %v provided", o.Type())
		}

		return objToValue(object, t)

	case reflect.UnsafePointer:
		return reflect.Zero(t), fmt.Errorf("unsupported type 'unsafeptr'")

	default:
		return reflect.Zero(t), fmt.Errorf("unsupported type")
	}
}

func args(t reflect.Type, a ...Object) (args []reflect.Value, err error) {
	if t.NumIn() != len(a) {
		return args, fmt.Errorf(
			"arguments mismatch: %d expected, %d provided",
			t.NumIn(),
			len(a),
		)
	}

	args = make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn() && err == nil; i++ {
		args[i], err = toValue(t.In(i), a[i])
	}
	return
}

func objToValue(o Class, t reflect.Type) (reflect.Value, error) {
	s := reflect.New(t).Elem()

	for i := 0; i < t.NumField(); i++ {
		goField := t.Field(i)
		tauField, ok := o.Get(goField.Name)
		if !ok {
			continue
		}

		val, err := toValue(goField.Type, tauField)
		if err != nil {
			return reflect.Zero(t), err
		}

		s.Field(i).Set(val)
	}

	return s, nil
}

func toObject(v reflect.Value) Object {
	switch v.Kind() {
	case reflect.String:
		return NewString(v.String())

	case reflect.Bool:
		return ParseBool(v.Bool())

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return NewInteger(v.Int())

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return NewInteger(int64(v.Uint()))

	case reflect.Float64, reflect.Float32:
		return NewFloat(v.Float())

	case reflect.Slice, reflect.Array:
		l := make(List, v.Len())

		for i := 0; i < v.Len(); i++ {
			l[i] = toObject(v.Index(i))
		}
		return l

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
	case 0:
		return NullObj

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
