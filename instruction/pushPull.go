package instruction

const InstructionTypePushPull Type = 0b100

type PushPull struct {
	isPull      bool
	ifThreshold bool
	block       bool
}

func NewPushPull(isPull bool, ifThreshold bool, block bool) *PushPull {
	return &PushPull{
		isPull:      isPull,
		ifThreshold: ifThreshold,
		block:       block,
	}
}

func (p PushPull) IsPull() bool      { return p.isPull }
func (p PushPull) IfThreshold() bool { return p.ifThreshold }
func (p PushPull) Block() bool       { return p.block }

func DecodePushPull(instr uint16) PushPull {
	return PushPull{
		isPull:      (instr>>7)&1 == 1,
		ifThreshold: (instr>>6)&1 == 1,
		block:       (instr>>5)&1 == 1,
	}
}

func (p PushPull) Encode() uint16 {
	var encoded uint16 = uint16(InstructionTypePushPull) << 13
	if p.isPull {
		encoded |= 0b1 << 7
	}
	if p.ifThreshold {
		encoded |= 0b1 << 6
	}
	if p.block {
		encoded |= 0b1 << 5
	}
	return encoded
}
