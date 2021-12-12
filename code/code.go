package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	Instructions []byte
	Opcode       byte
)

type Definition struct {
	Name          string
	OperandWidths []int
}

const (
	OpConstant Opcode = iota
	OpTrue
	OpFalse

	OpAdd
	OpSub
	OpMul
	OpDiv

	OpEqual
	OpNotEqual
	OpGreaterThan

	OpMinus
	OpBang

	OpPop
)

var (
	definitions = map[Opcode]*Definition{
		OpConstant: {"OpConstant", []int{2}},
		OpTrue:     {"OpTrue", []int{}},
		OpFalse:    {"OpFalse", []int{}},

		OpAdd: {"OpAdd", []int{}},
		OpSub: {"OpSub", []int{}},
		OpMul: {"OpMul", []int{}},
		OpDiv: {"OpDiv", []int{}},

		OpEqual:       {"OpEqual", []int{}},
		OpNotEqual:    {"OpNotEqual", []int{}},
		OpGreaterThan: {"OpGreaterThan", []int{}},

		OpMinus: {"OpMinus", []int{}},
		OpBang:  {"OpBang", []int{}},

		OpPop: {"OpPop", []int{}},
	}
)

func (ins Instructions) String() string {
	var out bytes.Buffer

	for i := 0; i < len(ins); {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += read + 1
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	var opCount = len(def.OperandWidths)

	if len(operands) != opCount {
		return fmt.Sprintf("error: operand len %d does not match defined %d\n", len(operands), opCount)
	}

	switch opCount {
	case 0:
		return def.Name

	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])

	default:
		return fmt.Sprintf("error: unhandled operand count for %s\n", def.Name)
	}
}

func Lookup(op byte) (*Definition, error) {
	if d, ok := definitions[Opcode(op)]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("opcode %d undefined", op)
}

// Returns the resulting bytecode from parsing the instruction.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionsLen := 1
	for _, w := range def.OperandWidths {
		instructionsLen += w
	}

	instructions := make([]byte, instructionsLen)
	instructions[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]

		switch width {
		case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))
		}
		offset += width
	}

	return instructions
}

// Decodes the operands of a bytecode instruction.
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
