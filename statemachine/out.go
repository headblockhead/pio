package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instr"
)

type outInput struct {
	autopullEnabled              bool
	baseOutPin                   uint8
	outShiftMovesRight           bool
	outWriteEnableBitIndex       uint8
	outWriteEnableUsed           bool
	outputShiftRegister          uint32
	outputShiftRegisterCounter   uint8
	pinCountOut                  uint8
	pullThreshold                uint8
	stickyOutSetAssertionEnabled bool

	fifoTX fifo.Reader
}

type outOutput struct {
	newEXECdInstruction bool
	execdInstruction    uint16

	jumped            bool
	newProgramCounter uint8

	writtenToInputShiftRegister  bool
	newInputShiftRegister        uint32
	newInputShiftRegisterCounter uint8

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
	shouldStall             bool
}

var ErrOutDestinationInvalid = errors.New("out destination invalid")

func executeOut(i instr.Out, input outInput) (outOutput, error) {
	var output = outOutput{
		newOutputShiftRegister:        input.outputShiftRegister,
		newOutputShiftRegisterCounter: input.outputShiftRegisterCounter,
	}

	shouldPull := input.autopullEnabled && (input.outputShiftRegisterCounter >= input.pullThreshold)
	if shouldPull {
		output.shouldStall = true
		if !input.fifoTX.IsEmpty() {
			osr, err := input.fifoTX.Read()
			if err != nil {
				return output, fmt.Errorf("error reading TX fifo for autopull: %w", err)
			}
			output.newOutputShiftRegister = osr
			output.newOutputShiftRegisterCounter = 0
		}
		return output, nil
	}

	var data uint32
	var mask uint32 = (0b1 << i.BitCount()) - 1
	if input.outShiftMovesRight {
		data = input.outputShiftRegister & mask
		output.newOutputShiftRegister >>= i.BitCount()
	} else {
		data = (input.outputShiftRegister >> (32 - i.BitCount()))
		output.newOutputShiftRegister <<= i.BitCount()
	}
	output.newOutputShiftRegisterCounter += i.BitCount()
	if output.newOutputShiftRegisterCounter > 32 {
		output.newOutputShiftRegisterCounter = 32
	}

	isPinOperation := i.Destination() == instr.OutDestinationPins || i.Destination() == instr.OutDestinationPinDirections

	outWriteEnable := ((data >> input.outWriteEnableBitIndex) & 0b1) == 1
	writePins := true
	if isPinOperation && input.outWriteEnableUsed && !outWriteEnable {
		writePins = false
		if input.stickyOutSetAssertionEnabled {
			output.newPinOutputEnables = 0
			output.newPinOutputEnablesMask = 0
			output.newPinOutputs = 0
			output.newPinOutputsMask = 0
		}
	}

	var pinData = bits.RotateLeft32(data, int(input.baseOutPin))
	var pinMask = bits.RotateLeft32((0b1<<input.pinCountOut)-1, int(input.baseOutPin))

	switch i.Destination() {
	case instr.OutDestinationPins:
		if writePins {
			output.newPinOutputs = pinData
			output.newPinOutputsMask = pinMask
		}
	case instr.OutDestinationX:
		output.newXRegister = data
		output.writtenToXRegister = true
	case instr.OutDestinationY:
		output.newYRegister = data
		output.writtenToYRegister = true
	case instr.OutDestinationNull:
		// discards data
	case instr.OutDestinationPinDirections:
		if writePins {
			output.newPinOutputEnables = pinData
			output.newPinOutputEnablesMask = pinMask
		}
	case instr.OutDestinationProgramCounter:
		output.jumped = true
		output.newProgramCounter = uint8(data % 32)
	case instr.OutDestinationInputShiftRegister:
		output.writtenToInputShiftRegister = true
		output.newInputShiftRegister = data
		output.newInputShiftRegisterCounter = i.BitCount()
	case instr.OutDestinationEXEC:
		output.newEXECdInstruction = true
		output.execdInstruction = uint16(data)
	default:
		return output, ErrOutDestinationInvalid
	}

	return output, nil
}
