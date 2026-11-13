package instr

import "fmt"

const TypeSet Type = 0b111

type SetDestination uint

const (
	SetDestinationPins          SetDestination = 0b000
	SetDestinationX             SetDestination = 0b001
	SetDestinationY             SetDestination = 0b010
	SetDestinationPinDirections SetDestination = 0b100
)

func (s SetDestination) String() string {
	switch s {
	case SetDestinationPins:
		return "pins"
	case SetDestinationX:
		return "x"
	case SetDestinationY:
		return "y"
	case SetDestinationPinDirections:
		return "pindirs"
	default:
		return "ERR"
	}
}

type Set struct {
	destination SetDestination
	data        uint
}

func NewSet(destination SetDestination, data uint) Set {
	return Set{
		destination: destination & 0b111,
		data:        data & 0b11111,
	}
}

func (s Set) Destination() SetDestination { return s.destination }
func (s Set) Data() uint                  { return s.data }

func DecodeSet(instruction uint16) Set {
	return Set{
		destination: SetDestination((instruction >> 5) & 0b111),
		data:        uint(instruction & 0b11111),
	}
}

func (s Set) Encode() uint16 {
	return uint16(TypeSet)<<13 | uint16(s.destination)<<5 | uint16(s.data)
}

func (s Set) String() string {
	return fmt.Sprintf("set %s, %d", s.destination, s.data)
}
