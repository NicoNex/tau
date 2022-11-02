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

//go:generate stringer -type=Opcode
const (
	OpConstant Opcode = iota
	OpTrue
	OpFalse
	OpNull
	OpList
	OpMap
	OpClosure
	OpCurrentClosure

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod

	OpBwAnd
	OpBwOr
	OpBwXor
	OpBwNot
	OpBwLShift
	OpBwRShift

	OpAnd
	OpOr
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterThanEqual
	OpIn

	OpMinus
	OpBang
	OpIndex

	OpCall
	OpConcurrentCall
	OpReturn
	OpReturnValue

	OpJump
	OpJumpNotTruthy

	OpDot
	OpDefine
	OpGetGlobal
	OpSetGlobal
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpGetFree
	OpLoadModule
	OpInterpolate

	OpPop
)

var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpNull:           {"OpNull", []int{}},
	OpList:           {"OpList", []int{2}},
	OpMap:            {"OpMap", []int{2}},
	OpClosure:        {"OpClosure", []int{2, 1}},
	OpCurrentClosure: {"OpCurrentClosure", []int{}},

	OpAdd: {"OpAdd", []int{}},
	OpSub: {"OpSub", []int{}},
	OpMul: {"OpMul", []int{}},
	OpDiv: {"OpDiv", []int{}},
	OpMod: {"OpMod", []int{}},

	OpBwAnd:    {"OpBwAnd", []int{}},
	OpBwOr:     {"OpBwOr", []int{}},
	OpBwXor:    {"OpBwXor", []int{}},
	OpBwNot:    {"OpBwNot", []int{}},
	OpBwLShift: {"OpBwLShift", []int{}},
	OpBwRShift: {"OpBwRshift", []int{}},

	OpAnd:              {"OpAnd", []int{}},
	OpOr:               {"OpOr", []int{}},
	OpEqual:            {"OpEqual", []int{}},
	OpNotEqual:         {"OpNotEqual", []int{}},
	OpGreaterThan:      {"OpGreaterThan", []int{}},
	OpGreaterThanEqual: {"OpGreaterThanEqual", []int{}},
	OpIn:               {"OpIn", []int{}},

	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},
	OpIndex: {"OpIndex", []int{}},

	OpCall:           {"OpCall", []int{1}},
	OpConcurrentCall: {"OpConcurrentCall", []int{1}},
	OpReturn:         {"OpReturn", []int{}},
	OpReturnValue:    {"OpReturnValue", []int{}},

	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},

	OpDot:         {"OpDot", []int{}},
	OpDefine:      {"OpDefine", []int{}},
	OpGetGlobal:   {"OpGetGlobal", []int{2}},
	OpSetGlobal:   {"OpSetGlobal", []int{2}},
	OpGetLocal:    {"OpGetLocal", []int{1}},
	OpSetLocal:    {"OpSetLocal", []int{1}},
	OpGetBuiltin:  {"OpGetBuiltin", []int{1}},
	OpGetFree:     {"OpGetFree", []int{1}},
	OpLoadModule:  {"OpLoadModule", []int{}},
	OpInterpolate: {"OpInterpolate", []int{2, 2}},

	OpPop: {"OpPop", []int{}},
}

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

	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])

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
		case 1:
			instructions[offset] = byte(o)

		case 2:
			binary.BigEndian.PutUint16(instructions[offset:], uint16(o))

		case 4:
			binary.BigEndian.PutUint32(instructions[offset:], uint32(o))

		case 8:
			binary.BigEndian.PutUint64(instructions[offset:], uint64(o))
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
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))

		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))

		case 4:
			operands[i] = int(ReadUint32(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint32(ins Instructions) uint32 {
	return binary.BigEndian.Uint32(ins)
}

func ReadUint64(ins Instructions) uint64 {
	return binary.BigEndian.Uint64(ins)
}
