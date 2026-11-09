package instr

func EncodeNOP() uint16 {
	nop := NewMove(MoveDestinationY, MoveOperationNone, MoveSourceY)
	return nop.Encode()
}
