package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type inInput struct {
	alreadyStalled            bool
	autopushEnabled           bool
	baseInPin                 uint8
	inShiftMovesRight         bool
	inputShiftRegister        uint32
	inputShiftRegisterCounter uint8
	outputShiftRegister       uint32
	pinInputs                 uint32
	pushThreshold             uint8
	xRegister                 uint32
	xRegisterInitialized      bool
	yRegister                 uint32
	yRegisterInitialized      bool

	fifoRX fifo.Writer
}

type inOutput struct {
	newInputShiftRegister        uint32
	newInputShiftRegisterCounter uint8
	shouldStall                  bool
}

var ErrInSourceInvalid = errors.New("in source invalid")

func executeIn(instr instruction.In, input inInput) (inOutput, error) {
	var output inOutput = inOutput{
		newInputShiftRegister:        input.inputShiftRegister,
		newInputShiftRegisterCounter: input.inputShiftRegisterCounter,
		shouldStall:                  false,
	}

	if !input.alreadyStalled {
		var data uint32
		var mask uint32 = (0b1 << instr.BitCount()) - 1
		switch instr.Source() {
		case instruction.InSourcePins:
			data = bits.RotateLeft32(input.pinInputs, -int(input.baseInPin)) & mask
		case instruction.InSourceX:
			if !input.xRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			data = (input.xRegister & mask)
		case instruction.InSourceY:
			if !input.yRegisterInitialized {
				return output, ErrYRegisterNotInitialized
			}
			data = (input.yRegister & mask)
		case instruction.InSourceNull:
			data = 0
		case instruction.InSourceISR:
			data = (input.inputShiftRegister & mask)
		case instruction.InSourceOSR:
			data = (input.outputShiftRegister & mask)
		default:
			return output, ErrInSourceInvalid
		}
		if input.inShiftMovesRight {
			output.newInputShiftRegister >>= instr.BitCount()
			output.newInputShiftRegister |= (data << (32 - instr.BitCount()))
		} else {
			output.newInputShiftRegister <<= instr.BitCount()
			output.newInputShiftRegister |= data
		}
		output.newInputShiftRegisterCounter += instr.BitCount()
		if output.newInputShiftRegisterCounter > 32 {
			output.newInputShiftRegisterCounter = 32
		}
	}

	shouldPush := input.autopushEnabled && (output.newInputShiftRegisterCounter >= input.pushThreshold)
	output.shouldStall = shouldPush && input.fifoRX.IsFull()
	if shouldPush && !input.fifoRX.IsFull() {
		err := input.fifoRX.Write(output.newInputShiftRegister)
		if err != nil {
			return output, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
		}
		output.newInputShiftRegister = 0
		output.newInputShiftRegisterCounter = 0
	}

	return output, nil
}
