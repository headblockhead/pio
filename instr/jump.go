package instr

const TypeJump Type = 0b000

type JumpCondition uint8

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

type Jump struct {
	condition JumpCondition
	address   uint8
}

func NewJump(condition JumpCondition, address uint8) Jump {
	return Jump{
		condition: condition & 0b111,
		address:   address & 0b11111,
	}
}

func (j Jump) Condition() JumpCondition { return j.condition }
func (j Jump) Address() uint8           { return j.address }

func DecodeJump(instr uint16) Jump {
	return Jump{
		condition: JumpCondition((instr >> 5) & 0b111),
		address:   uint8(instr & 0b11111),
	}
}

func (j Jump) Encode() uint16 {
	return uint16(TypeJump)<<13 | uint16(j.condition)<<5 | uint16(j.address)
}
