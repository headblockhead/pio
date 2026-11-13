package instr

import "fmt"

type Type uint

func DecodeType(instr uint16) Type {
	return Type(instr >> 13)
}

func String(instr uint16, pinCountSideset uint, sidesetIsOptional bool) string {
	t := DecodeType(instr)
	var decoded fmt.Stringer
	switch t {
	case TypeJump:
		decoded = DecodeJump(instr)
	case TypeWait:
		decoded = DecodeWait(instr)
	case TypeIn:
		decoded = DecodeIn(instr)
	case TypeOut:
		decoded = DecodeOut(instr)
	case TypePushPull:
		decoded = DecodePushPull(instr)
	case TypeMove:
		decoded = DecodeMove(instr)
	case TypeIRQ:
		decoded = DecodeIRQ(instr)
	case TypeSet:
		decoded = DecodeSet(instr)
	default:
		return "ERR"
	}
	sideset := DecodeSideset(instr, pinCountSideset, sidesetIsOptional)
	bitCountSideset := pinCountSideset
	if sidesetIsOptional {
		bitCountSideset++
	}
	delay := DecodeDelay(instr, bitCountSideset)
	return fmt.Sprintf("%s %s %s", decoded, sideset, delay)
}
