package instr

import "fmt"

const TypeIn Type = 0b010

type InSource uint

const (
	InSourcePins InSource = 0b000
	InSourceX    InSource = 0b001
	InSourceY    InSource = 0b010
	InSourceNull InSource = 0b011
	InSourceISR  InSource = 0b110
	InSourceOSR  InSource = 0b111
)

func (s InSource) String() string {
	switch s {
	case InSourcePins:
		return "pins"
	case InSourceX:
		return "x"
	case InSourceY:
		return "y"
	case InSourceNull:
		return "null"
	case InSourceISR:
		return "isr"
	case InSourceOSR:
		return "osr"
	default:
		return "ERR"
	}
}

type In struct {
	source   InSource
	bitCount uint
}

func NewIn(source InSource, bitCount uint) In {
	return In{
		source:   source & 0b111,
		bitCount: bitCount & 0b11111,
	}
}

func (i In) Source() InSource { return i.source }
func (i In) BitCount() uint {
	if i.bitCount == 0 {
		return 32
	}
	return i.bitCount
}

func DecodeIn(instr uint16) In {
	return In{
		source:   InSource((instr >> 5) & 0b111),
		bitCount: uint(instr & 0b11111),
	}
}

func (i In) Encode() uint16 {
	return uint16(TypeIn)<<13 | uint16(i.source)<<5 | uint16(i.bitCount)
}

func (i In) String() string {
	return fmt.Sprintf("in %s, %d", i.source, i.BitCount())
}
