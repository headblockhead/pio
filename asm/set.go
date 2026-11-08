package asm

const InstructionTypeSet InstructionType = 0b111

type SetDestination uint8

const (
	SetDestinationPins          SetDestination = 0b000
	SetDestinationX             SetDestination = 0b001
	SetDestinationY             SetDestination = 0b010
	SetDestinationPinDirections SetDestination = 0b100
)

type Set struct {
	destination SetDestination
	data        uint8
}

func NewSet(destination SetDestination, data uint8) Set {
	return Set{
		destination: destination & 0b111,
		data:        data & 0b11111,
	}
}

func (s Set) Destination() SetDestination { return s.destination }
func (s Set) Data() uint8                 { return s.data }

func DecodeSet(instruction uint16) Set {
	return Set{
		destination: SetDestination((instruction >> 5) & 0b111),
		data:        uint8(instruction & 0b11111),
	}
}

func (s Set) Encode() uint16 {
	return uint16(InstructionTypeSet)<<13 | uint16(s.destination)<<5 | uint16(s.data)
}
