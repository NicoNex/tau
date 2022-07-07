package obj

import (
	"plugin"
	"reflect"
)

type NativePlugin struct {
	*plugin.Plugin
}

func NewNativePlugin(path string) Object {
	p, err := plugin.Open(path)
	if err != nil {
		return NewError(err.Error())
	}
	return &NativePlugin{p}
}

func (n *NativePlugin) Get(name string) (Object, bool) {
	s, err := n.Lookup(name)
	if err != nil {
		return NewError(err.Error()), false
	}

	switch val := reflect.ValueOf(s); val.Kind() {
	case reflect.Func:
		return Builtin(func(a ...Object) (o Object) {
			defer func() {
				if err := recover(); err != nil {
					o = NewError("%v", err)
				}
			}()

			arguments, err := args(val.Type(), a...)
			if err != nil {
				return NewError(err.Error())
			}

			return multiplex(val.Call(arguments))
		}), true

	default:
		return toObject(val), true
	}
}

func (n *NativePlugin) Set(name string, o Object) Object {
	return NewError("cannot assign to native plugin")
}

func (n NativePlugin) String() string {
	return "<native plugin>"
}

func (n NativePlugin) Type() Type {
	return ObjectType
}
