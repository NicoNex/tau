package obj

import "reflect"

type NativeModule struct {
	name    string
	methods map[string]interface{}
}

func NewNativeModule(name string) Object {
	lib, ok := NativeLibs[name]
	if !ok {
		return NewError("import error: cannot find module with name %q", name)
	}

	return &NativeModule{
		name:    name,
		methods: lib,
	}
}

func (n *NativeModule) Get(name string) (Object, bool) {
	fn, ok := n.methods[name]
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

func (n *NativeModule) Set(name string, o Object) Object {
	return NewError("cannot assign to native module")
}

func (n NativeModule) Type() Type {
	return ClassType
}

func (n NativeModule) String() string {
	return "<native module>"
}
