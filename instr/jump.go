package instr

import "fmt"

const TypeJump Type = 0b000

type JumpCondition uint

const (
	JumpAlways                JumpCondition = 0b000
	JumpXZero                 JumpCondition = 0b001
	JumpXNonZeroThenDecrement JumpCondition = 0b010
	JumpYZero                 JumpCondition = 0b011
	JumpYNonZeroThenDecrement JumpCondition = 0b100
	JumpXNotEqualY            JumpCondition = 0b101
	JumpPin                   JumpCondition = 0b110
	JumpOSRENotEmpty          JumpCondition = 0b111
)

func (j JumpCondition) String() string {
	switch j {
	case JumpAlways:
		return ""
	case JumpXZero:
		return "!x"
	case JumpXNonZeroThenDecrement:
		return "x--"
	case JumpYZero:
		return "!y"
	case JumpYNonZeroThenDecrement:
		return "y--"
	case JumpXNotEqualY:
		return "x!=y"
	case JumpPin:
		return "pin"
	case JumpOSRENotEmpty:
		return "!osre"
	default:
		return "ERR"
	}
}

type Jump struct {
	condition JumpCondition
	address   uint
}

func NewJump(condition JumpCondition, address uint) Jump {
	return Jump{
		condition: condition & 0b111,
		address:   address & 0b11111,
	}
}

func (j Jump) Condition() JumpCondition { return j.condition }
func (j Jump) Address() uint            { return j.address }

func DecodeJump(instr uint16) Jump {
	return Jump{
		condition: JumpCondition((instr >> 5) & 0b111),
		address:   uint(instr & 0b11111),
	}
}

func (j Jump) Encode() uint16 {
	return uint16(TypeJump)<<13 | uint16(j.condition)<<5 | uint16(j.address)
}

func (j Jump) String() string {
	if j.Condition() != JumpAlways {
		return fmt.Sprintf("jmp %s %d", j.Condition(), j.address)
	}
	return fmt.Sprintf("jmp %d", j.address)
}
