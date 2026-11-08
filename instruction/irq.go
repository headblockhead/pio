package instruction

const InstructionTypeIRQ Type = 0b110

type IRQ struct {
	clear bool
	wait  bool
	index uint8
}

func NewIRQ(clear bool, wait bool, index uint8) IRQ {
	return IRQ{
		clear: clear,
		wait:  wait,
		index: index & 0b11111,
	}
}

func (i IRQ) Clear() bool  { return i.clear }
func (i IRQ) Wait() bool   { return i.wait }
func (i IRQ) Index() uint8 { return i.index }

func DecodeIRQ(instr uint16) IRQ {
	return IRQ{
		clear: (instr>>6)&0b1 == 1,
		wait:  (instr>>5)&0b1 == 1,
		index: uint8(instr & 0b11111),
	}
}

func (i IRQ) Encode() uint16 {
	var encoded uint16 = uint16(InstructionTypeIRQ)<<13 | uint16(i.index)
	if i.clear {
		encoded |= 0b1 << 6
	}
	if i.wait {
		encoded |= 0b1 << 5
	}
	return encoded
}
