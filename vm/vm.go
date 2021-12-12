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
	sp           int
}

const StackSize = 2048

var (
	True  = obj.True
	False = obj.False
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

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		consts:       bytecode.Constants,
		stack:        make([]obj.Object, StackSize),
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
		vm.push(obj.NewInteger(l + r))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		vm.push(obj.NewString(l + r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.NewFloat(l + r))

	default:
		return fmt.Errorf("unsupported operator '+' for types %v and %v", left.Type(), right.Type())
	}

	return nil
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
		vm.push(obj.NewInteger(l - r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.NewFloat(l - r))

	default:
		return fmt.Errorf("unsupported operator '-' for types %v and %v", left.Type(), right.Type())
	}

	return nil
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
		vm.push(obj.NewInteger(l * r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.NewFloat(l * r))

	default:
		return fmt.Errorf("unsupported operator '*' for types %v and %v", left.Type(), right.Type())
	}

	return nil
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
		vm.push(obj.NewInteger(l / r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.NewFloat(l / r))

	default:
		return fmt.Errorf("unsupported operator '/' for types %v and %v", left.Type(), right.Type())
	}

	return nil
}

func (vm *VM) execEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		vm.push(obj.ParseBool(left == right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.ParseBool(l == r))

	default:
		return fmt.Errorf("unsupported operator '==' for types %v and %v", left.Type(), right.Type())
	}

	return nil
}

func (vm *VM) execNotEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		vm.push(obj.ParseBool(left != right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.ParseBool(l != r))

	default:
		return fmt.Errorf("unsupported operator '!=' for types %v and %v", left.Type(), right.Type())
	}

	return nil
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
		vm.push(obj.ParseBool(l > r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		vm.push(obj.ParseBool(l > r))

	default:
		return fmt.Errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}

	return nil
}

// TODO: optimise this function with map[OpCode]func() error
func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			if err := vm.push(vm.consts[constIndex]); err != nil {
				return err
			}

		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}

		case code.OpAdd:
			if err := vm.execAdd(); err != nil {
				return err
			}

		case code.OpSub:
			if err := vm.execSub(); err != nil {
				return err
			}

		case code.OpMul:
			if err := vm.execMul(); err != nil {
				return err
			}

		case code.OpDiv:
			if err := vm.execDiv(); err != nil {
				return err
			}

		case code.OpEqual:
			if err := vm.execEqual(); err != nil {
				return err
			}

		case code.OpNotEqual:
			if err := vm.execNotEqual(); err != nil {
				return err
			}

		case code.OpGreaterThan:
			if err := vm.execGreaterThan(); err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()
		}

	}

	return nil
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
