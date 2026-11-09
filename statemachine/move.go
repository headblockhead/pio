package statemachine

import (
	"errors"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instr"
)

type moveInput struct {
	autopullEnabled              bool
	baseInPin                    uint8
	baseOutPin                   uint8
	inputShiftRegister           uint32
	outWriteEnableBitIndex       uint8
	outWriteEnableUsed           bool
	outputShiftRegister          uint32
	pinCountOut                  uint8
	pinInputs                    uint32
	statusValueComparisonLevel   uint8
	statusValueUsesRXFIFO        bool
	stickyOutSetAssertionEnabled bool
	xRegister                    uint32
	xRegisterInitialized         bool
	yRegister                    uint32
	yRegisterInitialized         bool

	fifoRX fifo.Writer
	fifoTX fifo.Reader
}

type moveOutput struct {
	newEXECdInstruction bool
	execdInstruction    uint16

	jumped            bool
	newProgramCounter uint8

	writtenToInputShiftRegister  bool
	newInputShiftRegister        uint32
	newInputShiftRegisterCounter uint8

	writtenToOutputShiftRegister  bool
	newOutputShiftRegister        uint32
	newOutputShiftRegisterCounter uint8

	newPinOutputEnables     uint32
	newPinOutputEnablesMask uint32
	newPinOutputs           uint32
	newPinOutputsMask       uint32
	writtenToXRegister      bool
	newXRegister            uint32
	writtenToYRegister      bool
	newYRegister            uint32
}

var ErrMoveSourceInvalid = errors.New("move source invalid")
var ErrCannotMoveFromOSRWhileAutopullEnabled = errors.New("cannot move from OSR while autopull enabled")
var ErrMoveOperationInvalid = errors.New("move operation invalid")
var ErrMoveDestinationInvalid = errors.New("move destination invalid")

func executeMove(i instr.Move, input moveInput) (moveOutput, error) {
	var output moveOutput

	var sourceData uint32
	switch i.Source() {
	case instr.MoveSourcePins:
		sourceData = bits.RotateLeft32(input.pinInputs, -int(input.baseInPin))
	case instr.MoveSourceX:
		if !input.xRegisterInitialized && i.Destination() != instr.MoveDestinationX {
			return output, ErrXRegisterNotInitialized
		}
		sourceData = input.xRegister
	case instr.MoveSourceY:
		if !input.yRegisterInitialized && i.Destination() != instr.MoveDestinationY {
			return output, ErrYRegisterNotInitialized
		}
		sourceData = input.yRegister
	case instr.MoveSourceNull:
		sourceData = 0
	case instr.MoveSourceStatus:
		var fifoLevel uint
		if input.statusValueUsesRXFIFO {
			fifoLevel = input.fifoRX.Level()
		} else {
			fifoLevel = input.fifoTX.Level()
		}
		if uint8(fifoLevel) < input.statusValueComparisonLevel {
			sourceData = 0xFFFFFFFF
		} else {
			sourceData = 0x00000000
		}
	case instr.MoveSourceISR:
		sourceData = input.inputShiftRegister
	case instr.MoveSourceOSR:
		if input.autopullEnabled {
			return output, ErrCannotMoveFromOSRWhileAutopullEnabled
		} else {
			sourceData = input.outputShiftRegister
		}
	default:
		return output, ErrMoveSourceInvalid
	}

	var modifiedData uint32
	switch i.Operation() {
	case instr.MoveOperationNone:
		modifiedData = sourceData
	case instr.MoveOperationInvert:
		modifiedData = sourceData ^ 0xFFFFFFFF
	case instr.MoveOperationBitReverse:
		modifiedData = bits.Reverse32(sourceData)
	default:
		return output, ErrMoveOperationInvalid
	}

	switch i.Destination() {
	case instr.MoveDestinationPins:
		outWriteEnable := ((modifiedData >> input.outWriteEnableBitIndex) & 0b1) == 1
		writePins := true
		if input.outWriteEnableUsed && !outWriteEnable {
			writePins = false
			if input.stickyOutSetAssertionEnabled {
				output.newPinOutputEnables = 0
				output.newPinOutputEnablesMask = 0
				output.newPinOutputs = 0
				output.newPinOutputsMask = 0
			}
		}
		if writePins {
			output.newPinOutputsMask = bits.RotateLeft32((0b1<<input.pinCountOut)-1, int(input.baseOutPin))
			output.newPinOutputs = bits.RotateLeft32(modifiedData, int(input.baseOutPin))
		}
	case instr.MoveDestinationX:
		if i.Source() != instr.MoveSourceX {
			output.writtenToXRegister = true
			output.newXRegister = modifiedData
		}
	case instr.MoveDestinationY:
		if i.Source() != instr.MoveSourceY {
			output.writtenToYRegister = true
			output.newYRegister = modifiedData
		}
	case instr.MoveDestinationEXEC:
		output.newEXECdInstruction = true
		output.execdInstruction = uint16(modifiedData)
	case instr.MoveDestinationPC:
		output.jumped = true
		output.newProgramCounter = uint8(modifiedData % 32)
	case instr.MoveDestinationISR:
		output.writtenToInputShiftRegister = true
		output.newInputShiftRegister = modifiedData
		output.newInputShiftRegisterCounter = 0
	case instr.MoveDestinationOSR:
		output.writtenToOutputShiftRegister = true
		output.newOutputShiftRegister = modifiedData
		output.newOutputShiftRegisterCounter = 0
	default:
		return output, ErrMoveDestinationInvalid
	}

	return output, nil
}
