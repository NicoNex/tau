package vm

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

const StackSize = 2048

type VM struct {
	consts       []obj.Object
	instructions code.Instructions
	stack        []obj.Object
	sp           int
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

		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			lVal := left.(*obj.Integer).Val()
			rVal := right.(*obj.Integer).Val()
			vm.push(obj.NewInteger(lVal + rVal))

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
