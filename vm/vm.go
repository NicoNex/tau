package vm

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type VM struct {
	consts       []obj.Object
	instructions code.Instructions
	stack        []obj.Object
	globals      []obj.Object
	sp           int
}

const (
	StackSize  = 2048
	GlobalSize = 65536
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
	return &VM{
		instructions: bytecode.Instructions,
		consts:       bytecode.Constants,
		stack:        make([]obj.Object, StackSize),
		globals:      make([]obj.Object, GlobalSize),
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []obj.Object) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		consts:       bytecode.Constants,
		stack:        make([]obj.Object, StackSize),
		globals:      s,
	}
}

func (vm *VM) LastPoppedStackElem() obj.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) execAdd() error {
	var (
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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

func (vm *VM) execEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
	)

	return vm.push(obj.ParseBool(isTruthy(left) && isTruthy(right)))
}

func (vm *VM) execOr() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	return vm.push(obj.ParseBool(isTruthy(left) || isTruthy(right)))
}

func (vm *VM) execGreaterThan() error {
	var (
		right = vm.pop()
		left  = vm.pop()
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
		right = vm.pop()
		left  = vm.pop()
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

func (vm *VM) execBang() error {
	var right = vm.pop()

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
	var right = vm.pop()

	switch r := right.(type) {
	case *obj.Integer:
		return vm.push(obj.NewInteger(-r.Val()))

	case *obj.Float:
		return vm.push(obj.NewFloat(-r.Val()))

	default:
		return fmt.Errorf("unsupported prefix operator '-' for type %v", r.Type())
	}
}

// TODO: optimise this function with map[OpCode]func() error
func (vm *VM) Run() (err error) {
	for ip := 0; ip < len(vm.instructions) && err == nil; ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			if err := vm.push(vm.consts[constIndex]); err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			if cond := vm.pop(); !isTruthy(cond) {
				ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err = vm.push(vm.globals[globalIndex])

		case code.OpNull:
			err = vm.push(Null)

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
