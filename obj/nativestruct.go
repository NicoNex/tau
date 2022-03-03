package obj

import "reflect"

type NativeStruct struct {
	s interface{}
}

func NewNativeStruct(s interface{}) Object {
	return &NativeStruct{s}
}

func (n *NativeStruct) Get(name string) (Object, bool) {
	return Builtin(func(a ...Object) (o Object) {
		defer func() {
			if err := recover(); err != nil {
				o = NewError("%v", err)
			}
		}()

		res := reflect.ValueOf(n.s).MethodByName(name).Call(args(a...))
		return multiplex(res)
	}), true
}

func (n *NativeStruct) Set(name string, o Object) Object {
	return NewError("cannot assign to native struct")
}

func (n NativeStruct) String() string {
	return "<native struct>"
}

func (n NativeStruct) Type() Type {
	return ClassType
}
