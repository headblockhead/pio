package asm

const InstructionTypeWait InstructionType = 0b001

type WaitSource uint8

const (
	WaitSourceGPIO WaitSource = 0b00
	WaitSourcePin  WaitSource = 0b01
	WaitSourceIRQ  WaitSource = 0b10
)

type Wait struct {
	polarity bool
	source   WaitSource
	index    uint8
}

func NewWait(polarity bool, source WaitSource, index uint8) Wait {
	return Wait{
		polarity: polarity,
		source:   source & 0b11,
		index:    index & 0b11111,
	}
}

func (w Wait) Polarity() bool     { return w.polarity }
func (w Wait) Source() WaitSource { return w.source }
func (w Wait) Index() uint8       { return w.index }

func DecodeWait(instr uint16) Wait {
	return Wait{
		polarity: (instr>>7)&0b1 == 1,
		source:   WaitSource((instr >> 5) & 0b11),
		index:    uint8(instr & 0b11111),
	}
}

func (w Wait) Encode() uint16 {
	var encoded uint16 = uint16(InstructionTypeWait)<<13 | uint16(w.source)<<5 | uint16(w.index)
	if w.polarity {
		encoded |= 0b1 << 7
	}
	return encoded
}
