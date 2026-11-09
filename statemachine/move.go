package statemachine

import (
	"errors"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type MoveInput struct {
	AutopullEnabled              bool
	BaseInPin                    uint8
	BaseOutPin                   uint8
	InputShiftRegister           uint32
	OutWriteEnableBitIndex       uint8
	OutWriteEnableUsed           bool
	OutputShiftRegister          uint32
	PinCountOut                  uint8
	PinInputs                    uint32
	StatusValueComparisonLevel   uint8
	StatusValueUsesRXFIFO        bool
	StickyOutSetAssertionEnabled bool
	XRegister                    uint32
	XRegisterInitialized         bool
	YRegister                    uint32
	YRegisterInitialized         bool

	FIFORX fifo.Writer
	FIFOTX fifo.Reader
}

type MoveOutput struct {
	NewEXECdInstruction bool
	EXECdInstruction    uint16

	Jumped            bool
	NewProgramCounter uint8

	WrittenToInputShiftRegister  bool
	NewInputShiftRegister        uint32
	NewInputShiftRegisterCounter uint8

	WrittenToOutputShiftRegister  bool
	NewOutputShiftRegister        uint32
	NewOutputShiftRegisterCounter uint8

	NewPinOutputEnables     uint32
	NewPinOutputEnablesMask uint32
	NewPinOutputs           uint32
	NewPinOutputsMask       uint32
	WrittenToXRegister      bool
	NewXRegister            uint32
	WrittenToYRegister      bool
	NewYRegister            uint32
}

var ErrMoveSourceInvalid = errors.New("move source invalid")
var ErrCannotMoveFromOSRWhileAutopullEnabled = errors.New("cannot move from OSR while autopull enabled")
var ErrMoveOperationInvalid = errors.New("move operation invalid")
var ErrMoveDestinationInvalid = errors.New("move destination invalid")

func ExecuteMove(instr instruction.Move, input MoveInput) (MoveOutput, error) {
	var output MoveOutput

	var sourceData uint32
	switch instr.Source() {
	case instruction.MoveSourcePins:
		sourceData = bits.RotateLeft32(input.PinInputs, -int(input.BaseInPin))
	case instruction.MoveSourceX:
		if !input.XRegisterInitialized && instr.Destination() != instruction.MoveDestinationX {
			return output, ErrXRegisterNotInitialized
		}
		sourceData = input.XRegister
	case instruction.MoveSourceY:
		if !input.YRegisterInitialized && instr.Destination() != instruction.MoveDestinationY {
			return output, ErrYRegisterNotInitialized
		}
		sourceData = input.YRegister
	case instruction.MoveSourceNull:
		sourceData = 0
	case instruction.MoveSourceStatus:
		var fifoLevel uint
		if input.StatusValueUsesRXFIFO {
			fifoLevel = input.FIFORX.Level()
		} else {
			fifoLevel = input.FIFOTX.Level()
		}
		if uint8(fifoLevel) < input.StatusValueComparisonLevel {
			sourceData = 0xFFFFFFFF
		} else {
			sourceData = 0x00000000
		}
	case instruction.MoveSourceISR:
		sourceData = input.InputShiftRegister
	case instruction.MoveSourceOSR:
		if input.AutopullEnabled {
			return output, ErrCannotMoveFromOSRWhileAutopullEnabled
		} else {
			sourceData = input.OutputShiftRegister
		}
	default:
		return output, ErrMoveSourceInvalid
	}

	var modifiedData uint32
	switch instr.Operation() {
	case instruction.MoveOperationNone:
		modifiedData = sourceData
	case instruction.MoveOperationInvert:
		modifiedData = sourceData ^ 0xFFFFFFFF
	case instruction.MoveOperationBitReverse:
		modifiedData = bits.Reverse32(sourceData)
	default:
		return output, ErrMoveOperationInvalid
	}

	switch instr.Destination() {
	case instruction.MoveDestinationPins:
		outWriteEnable := ((modifiedData >> input.OutWriteEnableBitIndex) & 0b1) == 1
		writePins := true
		if input.OutWriteEnableUsed && !outWriteEnable {
			writePins = false
			if input.StickyOutSetAssertionEnabled {
				output.NewPinOutputEnables = 0
				output.NewPinOutputEnablesMask = 0
				output.NewPinOutputs = 0
				output.NewPinOutputsMask = 0
			}
		}
		if writePins {
			output.NewPinOutputsMask = bits.RotateLeft32((0b1<<input.PinCountOut)-1, int(input.BaseOutPin))
			output.NewPinOutputs = bits.RotateLeft32(modifiedData, int(input.BaseOutPin))
		}
	case instruction.MoveDestinationX:
		if instr.Source() != instruction.MoveSourceX {
			output.WrittenToXRegister = true
			output.NewXRegister = modifiedData
		}
	case instruction.MoveDestinationY:
		if instr.Source() != instruction.MoveSourceY {
			output.WrittenToYRegister = true
			output.NewYRegister = modifiedData
		}
	case instruction.MoveDestinationEXEC:
		output.NewEXECdInstruction = true
		output.EXECdInstruction = uint16(modifiedData)
	case instruction.MoveDestinationPC:
		output.Jumped = true
		output.NewProgramCounter = uint8(modifiedData % 32)
	case instruction.MoveDestinationISR:
		output.WrittenToInputShiftRegister = true
		output.NewInputShiftRegister = modifiedData
		output.NewInputShiftRegisterCounter = 0
	case instruction.MoveDestinationOSR:
		output.WrittenToOutputShiftRegister = true
		output.NewOutputShiftRegister = modifiedData
		output.NewOutputShiftRegisterCounter = 0
	default:
		return output, ErrMoveDestinationInvalid
	}

	return output, nil
}
