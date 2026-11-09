package instruction

const TypeMove Type = 0b101

type MoveDestination uint8
type MoveOperation uint8
type MoveSource uint8

const (
	MoveDestinationPins MoveDestination = 0b000
	MoveDestinationX    MoveDestination = 0b001
	MoveDestinationY    MoveDestination = 0b010
	MoveDestinationEXEC MoveDestination = 0b100
	MoveDestinationPC   MoveDestination = 0b101
	MoveDestinationISR  MoveDestination = 0b110
	MoveDestinationOSR  MoveDestination = 0b111

	MoveOperationNone       MoveOperation = 0b00
	MoveOperationInvert     MoveOperation = 0b01
	MoveOperationBitReverse MoveOperation = 0b10

	MoveSourcePins   MoveSource = 0b000
	MoveSourceX      MoveSource = 0b001
	MoveSourceY      MoveSource = 0b010
	MoveSourceNull   MoveSource = 0b011
	MoveSourceStatus MoveSource = 0b101
	MoveSourceISR    MoveSource = 0b110
	MoveSourceOSR    MoveSource = 0b111
)

type Move struct {
	destination MoveDestination
	operation   MoveOperation
	source      MoveSource
}

func NewMove(destination MoveDestination, operation MoveOperation, source MoveSource) Move {
	return Move{
		destination: destination & 0b111,
		operation:   operation & 0b11,
		source:      source & 0b111,
	}
}

func (m Move) Destination() MoveDestination { return m.destination }
func (m Move) Operation() MoveOperation     { return m.operation }
func (m Move) Source() MoveSource           { return m.source }

func DecodeMove(instr uint16) Move {
	return Move{
		destination: MoveDestination((instr >> 5) & 0b111),
		operation:   MoveOperation((instr >> 3) & 0b11),
		source:      MoveSource(instr & 0b111),
	}
}

func (m Move) Encode() uint16 {
	return uint16(TypeMove)<<13 | uint16(m.destination)<<5 | uint16(m.operation)<<3 | uint16(m.source)
}
