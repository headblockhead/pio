package instr

type Type uint

func DecodeType(instr uint16) Type {
	return Type(instr >> 13)
}
