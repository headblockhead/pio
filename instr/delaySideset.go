package instr

type Delay struct {
	cycles uint8
}

func NewDelay(delayCycles uint8) Delay {
	return Delay{
		cycles: delayCycles & 0b11111,
	}
}

func (d Delay) Cycles() uint8 { return d.cycles }

func DecodeDelay(instr uint16, bitCountSideset uint8) Delay {
	return Delay{
		cycles: uint8(instr>>8) & (0b1<<(5-bitCountSideset) - 1),
	}
}

func (d Delay) Encode() uint16 {
	return uint16(d.cycles) << 8
}

type Sideset struct {
	data   uint8
	active bool
}

func NewSideset(sidesetData uint8) Sideset {
	return Sideset{
		data:   sidesetData & 0b11111,
		active: true,
	}
}

func (s Sideset) Data() uint8  { return s.data }
func (s Sideset) Active() bool { return s.active }

func DecodeSideset(instr uint16, pinCountSideset uint8, sidesetIsOptional bool) Sideset {
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
		data:   uint8(instr>>lsbPosition) & ((0b1 << pinCountSideset) - 1),
		active: true,
	}
}

func (s Sideset) Encode(pinCountSideset uint8, sidesetIsOptional bool) uint16 {
	if sidesetIsOptional {
		return 0b1<<12 | uint16(s.data)<<(12-pinCountSideset)
	}
	return uint16(s.data) << (13 - pinCountSideset)
}
