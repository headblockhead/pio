package instr

import "fmt"

const TypeOut Type = 0b011

type OutDestination uint

const (
	OutDestinationPins               OutDestination = 0b000
	OutDestinationX                  OutDestination = 0b001
	OutDestinationY                  OutDestination = 0b010
	OutDestinationNull               OutDestination = 0b011
	OutDestinationPinDirections      OutDestination = 0b100
	OutDestinationProgramCounter     OutDestination = 0b101
	OutDestinationInputShiftRegister OutDestination = 0b110
	OutDestinationEXEC               OutDestination = 0b111
)

func (o OutDestination) String() string {
	switch o {
	case OutDestinationPins:
		return "pins"
	case OutDestinationX:
		return "x"
	case OutDestinationY:
		return "y"
	case OutDestinationNull:
		return "null"
	case OutDestinationPinDirections:
		return "pindirs"
	case OutDestinationProgramCounter:
		return "pc"
	case OutDestinationInputShiftRegister:
		return "isr"
	case OutDestinationEXEC:
		return "exec"
	default:
		return fmt.Sprintf("instr.OutDestination(%d)", o)
	}
}

type Out struct {
	destination OutDestination
	bitCount    uint
}

func NewOut(destination OutDestination, bitCount uint) Out {
	return Out{
		destination: destination & 0b111,
		bitCount:    bitCount & 0b11111,
	}
}

func (o Out) Destination() OutDestination { return o.destination }
func (o Out) BitCount() uint {
	if o.bitCount == 0 {
		return 32
	}
	return o.bitCount
}

func DecodeOut(instr uint16) Out {
	return Out{
		destination: OutDestination((instr >> 5) & 0b111),
		bitCount:    uint(instr & 0b11111),
	}
}

func (o Out) Encode() uint16 {
	return uint16(TypeOut)<<13 | uint16(o.destination)<<5 | uint16(o.bitCount)
}

func (o Out) String() string {
	return fmt.Sprintf("out    %s, %d", o.destination, o.BitCount())
}
