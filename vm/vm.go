package vm

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type VM struct {
	consts     []obj.Object
	stack      []obj.Object
	globals    []obj.Object
	sp         int
	frames     []*Frame
	frameIndex int
}

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

var (
	True  = obj.True
	False = obj.False
	Null  = obj.NullObj
)

func assertTypes(o obj.Object, types ...obj.Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func toFloat(l, r obj.Object) (obj.Object, obj.Object) {
	if i, ok := l.(*obj.Integer); ok {
		l = obj.NewFloat(float64(*i))
	}
	if i, ok := r.(*obj.Integer); ok {
		r = obj.NewFloat(float64(*i))
	}
	return l, r
}

func isTruthy(o obj.Object) bool {
	switch val := o.(type) {
	case *obj.Boolean:
		return o == obj.True
	case *obj.Integer:
		return val.Val() != 0
	case *obj.Float:
		return val.Val() != 0
	case *obj.Null:
		return false
	default:
		return true
	}
}

func New(bytecode *compiler.Bytecode) *VM {
	vm := &VM{
		consts:     bytecode.Constants,
		stack:      make([]obj.Object, StackSize),
		globals:    make([]obj.Object, GlobalSize),
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
	}

	fn := &obj.Function{Instructions: bytecode.Instructions}
	vm.frames[0] = NewFrame(&obj.Closure{Fn: fn}, 0)
	return vm
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []obj.Object) *VM {
	vm := &VM{
		consts:     bytecode.Constants,
		stack:      make([]obj.Object, StackSize),
		globals:    s,
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
	}
	fn := &obj.Function{Instructions: bytecode.Instructions}
	vm.frames[0] = NewFrame(&obj.Closure{Fn: fn}, 0)
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) LastPoppedStackElem() obj.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) execDot() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch l := left.(type) {
	case obj.Class:
		return vm.push(obj.NewGetSetter(l, right.String()))

	case obj.GetSetter:
		o := l.Object()
		return vm.push(obj.NewGetSetter(o.(obj.Class), right.String()))

	default:
		return fmt.Errorf("%v object has no attribute %s", left.Type(), right)
	}
}

func (vm *VM) execDefine() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	l, ok := left.(obj.Setter)
	if !ok {
		return fmt.Errorf("cannot assign to type %v", left.Type())
	}
	return vm.push(l.Set(right))
}

func (vm *VM) execAdd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.NewInteger(l + r))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return vm.push(obj.NewString(l + r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.NewFloat(l + r))

	default:
		return fmt.Errorf("unsupported operator '+' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execSub() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.NewInteger(l - r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.NewFloat(l - r))

	default:
		return fmt.Errorf("unsupported operator '-' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execMul() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.NewInteger(l * r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.NewFloat(l * r))

	default:
		return fmt.Errorf("unsupported operator '*' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execDiv() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.NewInteger(l / r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.NewFloat(l / r))

	default:
		return fmt.Errorf("unsupported operator '/' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execMod() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '%%' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return fmt.Errorf("can't divide by 0")
	}
	return vm.push(obj.NewInteger(l % r))
}

func (vm *VM) execBwAnd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '&' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(l & r))
}

func (vm *VM) execBwOr() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '|' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(l | r))
}

func (vm *VM) execBwXor() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '^' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(l ^ r))
}

func (vm *VM) execBwNot() error {
	var left = obj.Unwrap(vm.pop())

	if !assertTypes(left, obj.IntType) {
		return fmt.Errorf("unsupported operator '~' for type %v", left.Type())
	}

	l := left.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(^l))
}

func (vm *VM) execBwLShift() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '<<' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(l << r))
}

func (vm *VM) execBwRShift() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '>>' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return vm.push(obj.NewInteger(l >> r))
}

func (vm *VM) execEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left == right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.ParseBool(l == r))

	default:
		return fmt.Errorf("unsupported operator '==' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execNotEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left != right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.ParseBool(l != r))

	default:
		return fmt.Errorf("unsupported operator '!=' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execAnd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	return vm.push(obj.ParseBool(isTruthy(left) && isTruthy(right)))
}

func (vm *VM) execIn() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return fmt.Errorf("unsupported operator 'in' for type %v", left.Type())
	}
	if !assertTypes(right, obj.ListType, obj.StringType) {
		return fmt.Errorf("unsupported operator 'in' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return vm.push(obj.ParseBool(strings.Contains(r, l)))

	case assertTypes(right, obj.ListType):
		for _, o := range right.(obj.List).Val() {
			if !assertTypes(left, o.Type()) {
				continue
			}
			if assertTypes(left, obj.BoolType, obj.NullType) && left == o {
				return vm.push(obj.True)
			}

			switch l := left.(type) {
			case *obj.String:
				r := o.(*obj.String)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}

			case *obj.Integer:
				r := o.(*obj.Integer)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}

			case *obj.Float:
				r := o.(*obj.Float)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}
			}
		}
		return vm.push(obj.False)

	default:
		return fmt.Errorf(
			"invalid operation %v in %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (vm *VM) execOr() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	return vm.push(obj.ParseBool(isTruthy(left) || isTruthy(right)))
}

func (vm *VM) execGreaterThan() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.ParseBool(l > r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.ParseBool(l > r))

	default:
		return fmt.Errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execGreaterThanEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return vm.push(obj.ParseBool(l >= r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return vm.push(obj.ParseBool(l >= r))

	default:
		return fmt.Errorf("unsupported operator '>=' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execIndex() error {
	var (
		index = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.ListType) && assertTypes(index, obj.IntType):
		l := left.(obj.List)
		i := int(index.(*obj.Integer).Val())

		if i < 0 || i >= len(l) {
			return fmt.Errorf("index out of range")
		}
		return vm.push(l[i])

	case assertTypes(left, obj.StringType) && assertTypes(index, obj.IntType):
		s := left.(*obj.String).Val()
		i := int(index.(*obj.Integer).Val())

		if i < 0 || i >= len(s) {
			return fmt.Errorf("index out of range")
		}
		return vm.push(obj.NewString(string(s[i])))

	case assertTypes(left, obj.MapType) && assertTypes(index, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
		m := left.(obj.Map)
		k := index.(obj.Hashable)
		return vm.push(m.Get(k.KeyHash()).Value)

	default:
		return fmt.Errorf("invalid index operator for types %v and %v", left.Type(), index.Type())
	}
}

func (vm *VM) execBang() error {
	var right = obj.Unwrap(vm.pop())

	switch b := right.(type) {
	case *obj.Boolean:
		return vm.push(obj.ParseBool(!b.Val()))

	case *obj.Null:
		return vm.push(True)

	default:
		return vm.push(False)
	}
}

func (vm *VM) execMinus() error {
	var right = obj.Unwrap(vm.pop())

	switch r := right.(type) {
	case *obj.Integer:
		return vm.push(obj.NewInteger(-r.Val()))

	case *obj.Float:
		return vm.push(obj.NewFloat(-r.Val()))

	default:
		return fmt.Errorf("unsupported prefix operator '-' for type %v", r.Type())
	}
}

func (vm *VM) execReturnValue() error {
	retVal := obj.Unwrap(vm.pop())
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1

	return vm.push(retVal)
}

func (vm *VM) execReturn() error {
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1

	return vm.push(Null)
}

func (vm *VM) execCurrentClosure() error {
	return vm.push(vm.currentFrame().cl)
}

func (vm *VM) execCall(nargs int) error {
	callee := vm.stack[vm.sp-1-nargs]

	switch callee := callee.(type) {
	case *obj.Closure:
		return vm.callClosure(callee, nargs)
	case obj.Builtin:
		return vm.callBuiltin(callee, nargs)
	case obj.Getter:
		o := callee.Object()
		fn, ok := o.(*obj.Closure)
		if !ok {
			return fmt.Errorf("calling non-function")
		}
		return vm.callClosure(fn, nargs)
	default:
		return fmt.Errorf("calling non-function")
	}
}

func (vm *VM) buildList(start, end int) obj.Object {
	var elements = make([]obj.Object, end-start)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	return obj.NewList(elements...)
}

func (vm *VM) buildMap(start, end int) (obj.Object, error) {
	var m = obj.NewMap()

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := obj.MapPair{Key: key, Value: val}
		mapKey, ok := key.(obj.Hashable)
		if !ok {
			return nil, fmt.Errorf("invalid map key type %v", key.Type())
		}
		m.Set(mapKey.KeyHash(), pair)
	}

	return m, nil
}

func (vm *VM) callClosure(cl *obj.Closure, nargs int) error {
	if nargs != cl.Fn.NumParams {
		return fmt.Errorf("wrong number of arguments: expected %d, got %d", cl.Fn.NumParams, nargs)
	}

	frame := NewFrame(cl, vm.sp-nargs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals
	return nil
}

func (vm *VM) callBuiltin(fn obj.Builtin, nargs int) error {
	args := vm.stack[vm.sp-nargs : vm.sp]
	res := fn(args...)
	vm.sp = vm.sp - nargs - 1

	if res == nil {
		return vm.push(Null)
	}
	return vm.push(res)
}

func (vm *VM) pushClosure(constIdx, numFree int) error {
	constant := vm.consts[constIdx]
	fn, ok := constant.(*obj.Function)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]obj.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree
	return vm.push(&obj.Closure{Fn: fn, Free: free})
}

func (vm *VM) Run() (err error) {
	var (
		ip  int
		ins code.Instructions
		op  code.Opcode
	)

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 && err == nil {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2

			if err := vm.push(vm.consts[constIndex]); err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if cond := obj.Unwrap(vm.pop()); !isTruthy(cond) {
				vm.currentFrame().ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.peek()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.globals[globalIndex])

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.peek()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err = vm.push(vm.stack[frame.basePointer+int(localIndex)])

		case code.OpList:
			nElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			list := vm.buildList(vm.sp-nElements, vm.sp)
			vm.sp = vm.sp - nElements
			err = vm.push(list)

		case code.OpMap:
			nElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			mapObj, err := vm.buildMap(vm.sp-nElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - nElements
			err = vm.push(mapObj)

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			err = vm.execCall(int(numArgs))

		case code.OpGetBuiltin:
			idx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			def := obj.Builtins[idx]
			err = vm.push(def.Builtin)

		case code.OpClosure:
			constIdx := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			err = vm.pushClosure(int(constIdx), int(numFree))

		case code.OpGetFree:
			freeIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			closure := vm.currentFrame().cl
			err = vm.push(closure.Free[freeIdx])

		case code.OpDot:
			err = vm.execDot()

		case code.OpDefine:
			err = vm.execDefine()

		case code.OpCurrentClosure:
			err = vm.execCurrentClosure()

		case code.OpReturn:
			err = vm.execReturn()

		case code.OpReturnValue:
			err = vm.execReturnValue()

		case code.OpNull:
			err = vm.push(Null)

		case code.OpIndex:
			err = vm.execIndex()

		case code.OpTrue:
			err = vm.push(True)

		case code.OpFalse:
			err = vm.push(False)

		case code.OpAdd:
			err = vm.execAdd()

		case code.OpSub:
			err = vm.execSub()

		case code.OpMul:
			err = vm.execMul()

		case code.OpDiv:
			err = vm.execDiv()

		case code.OpMod:
			err = vm.execMod()

		case code.OpBwAnd:
			err = vm.execBwAnd()

		case code.OpBwOr:
			err = vm.execBwOr()

		case code.OpBwXor:
			err = vm.execBwXor()

		case code.OpBwLShift:
			err = vm.execBwLShift()

		case code.OpBwRShift:
			err = vm.execBwRShift()

		case code.OpEqual:
			err = vm.execEqual()

		case code.OpNotEqual:
			err = vm.execNotEqual()

		case code.OpGreaterThan:
			err = vm.execGreaterThan()

		case code.OpGreaterThanEqual:
			err = vm.execGreaterThanEqual()

		case code.OpAnd:
			err = vm.execAnd()

		case code.OpIn:
			err = vm.execIn()

		case code.OpOr:
			err = vm.execOr()

		case code.OpBang:
			err = vm.execBang()

		case code.OpMinus:
			err = vm.execMinus()

		case code.OpPop:
			vm.pop()
		}

	}
	return
}

func (vm *VM) push(o obj.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() obj.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) peek() obj.Object {
	return vm.stack[vm.sp-1]
}
