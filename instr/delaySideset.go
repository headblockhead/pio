package instr

import "fmt"

type Delay struct {
	cycles uint
}

func NewDelay(delayCycles uint) Delay {
	return Delay{
		cycles: delayCycles & 0b11111,
	}
}

func (d Delay) Cycles() uint { return d.cycles }

func DecodeDelay(instr uint16, bitCountSideset uint) Delay {
	return Delay{
		cycles: uint(instr>>8) & (0b1<<(5-bitCountSideset) - 1),
	}
}

func (d Delay) Encode() uint16 {
	return uint16(d.cycles) << 8
}

func (d Delay) String() string {
	if d.cycles == 0 {
		return ""
	}
	return fmt.Sprintf("[%d]", d.cycles)
}

type Sideset struct {
	data   uint
	active bool
}

func NewSideset(sidesetData uint) Sideset {
	return Sideset{
		data:   sidesetData & 0b11111,
		active: true,
	}
}

func (s Sideset) Data() uint   { return s.data }
func (s Sideset) Active() bool { return s.active }

func DecodeSideset(instr uint16, pinCountSideset uint, sidesetIsOptional bool) Sideset {
	if pinCountSideset == 0 {
		return Sideset{active: false}
	}
	lsbPosition := 13 - pinCountSideset
	if sidesetIsOptional {
		if (instr>>12)&0b1 != 1 {
			return Sideset{
				active: false,
			}
		}
		lsbPosition = 12 - pinCountSideset
	}
	return Sideset{
		data:   uint(instr>>lsbPosition) & ((0b1 << pinCountSideset) - 1),
		active: true,
	}
}

func (s Sideset) Encode(pinCountSideset uint, sidesetIsOptional bool) uint16 {
	if sidesetIsOptional {
		return 0b1<<12 | uint16(s.data)<<(12-pinCountSideset)
	}
	return uint16(s.data) << (13 - pinCountSideset)
}

func (s Sideset) String() string {
	if !s.active {
		return ""
	}
	return fmt.Sprintf("side %2d", s.data)
}
