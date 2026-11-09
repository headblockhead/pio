package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instr"
)

type inInput struct {
	alreadyStalled            bool
	autopushEnabled           bool
	baseInPin                 uint
	inShiftMovesRight         bool
	inputShiftRegister        uint32
	inputShiftRegisterCounter uint
	outputShiftRegister       uint32
	pinInputs                 uint32
	pushThreshold             uint
	xRegister                 uint32
	xRegisterInitialized      bool
	yRegister                 uint32
	yRegisterInitialized      bool

	fifoRX fifo.Writer
}

type inOutput struct {
	newInputShiftRegister        uint32
	newInputShiftRegisterCounter uint
	shouldStall                  bool
}

var ErrInSourceInvalid = errors.New("in source invalid")

func executeIn(i instr.In, input inInput) (inOutput, error) {
	var output = inOutput{
		newInputShiftRegister:        input.inputShiftRegister,
		newInputShiftRegisterCounter: input.inputShiftRegisterCounter,
		shouldStall:                  false,
	}

	if !input.alreadyStalled {
		var data uint32
		var mask uint32 = (0b1 << i.BitCount()) - 1
		switch i.Source() {
		case instr.InSourcePins:
			data = bits.RotateLeft32(input.pinInputs, -int(input.baseInPin)) & mask
		case instr.InSourceX:
			if !input.xRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			data = (input.xRegister & mask)
		case instr.InSourceY:
			if !input.yRegisterInitialized {
				return output, ErrYRegisterNotInitialized
			}
			data = (input.yRegister & mask)
		case instr.InSourceNull:
			data = 0
		case instr.InSourceISR:
			data = (input.inputShiftRegister & mask)
		case instr.InSourceOSR:
			data = (input.outputShiftRegister & mask)
		default:
			return output, ErrInSourceInvalid
		}
		if input.inShiftMovesRight {
			output.newInputShiftRegister >>= i.BitCount()
			output.newInputShiftRegister |= (data << (32 - i.BitCount()))
		} else {
			output.newInputShiftRegister <<= i.BitCount()
			output.newInputShiftRegister |= data
		}
		output.newInputShiftRegisterCounter += i.BitCount()
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
