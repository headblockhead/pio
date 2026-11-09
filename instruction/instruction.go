package instruction

type Type uint8

func DecodeType(instr uint16) Type {
	return Type(instr >> 13)
}
