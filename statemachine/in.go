package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type InInput struct {
	AlreadyStalled            bool
	AutopushEnabled           bool
	BaseInPin                 uint8
	InShiftMovesRight         bool
	InputShiftRegister        uint32
	InputShiftRegisterCounter uint8
	OutputShiftRegister       uint32
	PinInputs                 uint32
	PushThreshold             uint8
	XRegister                 uint32
	XRegisterInitialized      bool
	YRegister                 uint32
	YRegisterInitialized      bool

	FIFORX fifo.Writer
}

type InOutput struct {
	NewInputShiftRegister        uint32
	NewInputShiftRegisterCounter uint8
	ShouldStall                  bool
}

var ErrInSourceInvalid = errors.New("in source invalid")

func ExecuteIn(instr instruction.In, input InInput) (InOutput, error) {
	var output InOutput = InOutput{
		NewInputShiftRegister:        input.InputShiftRegister,
		NewInputShiftRegisterCounter: input.InputShiftRegisterCounter,
		ShouldStall:                  false,
	}

	if !input.AlreadyStalled {
		var data uint32
		var mask uint32 = (0b1 << instr.BitCount()) - 1
		switch instr.Source() {
		case instruction.InSourcePins:
			data = bits.RotateLeft32(input.PinInputs, -int(input.BaseInPin)) & mask
		case instruction.InSourceX:
			if !input.XRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			data = (input.XRegister & mask)
		case instruction.InSourceY:
			if !input.YRegisterInitialized {
				return output, ErrYRegisterNotInitialized
			}
			data = (input.YRegister & mask)
		case instruction.InSourceNull:
			data = 0
		case instruction.InSourceISR:
			data = (input.InputShiftRegister & mask)
		case instruction.InSourceOSR:
			data = (input.OutputShiftRegister & mask)
		default:
			return output, ErrInSourceInvalid
		}
		if input.InShiftMovesRight {
			output.NewInputShiftRegister >>= instr.BitCount()
			output.NewInputShiftRegister |= (data << (32 - instr.BitCount()))
		} else {
			output.NewInputShiftRegister <<= instr.BitCount()
			output.NewInputShiftRegister |= data
		}
		output.NewInputShiftRegisterCounter += instr.BitCount()
		if output.NewInputShiftRegisterCounter > 32 {
			output.NewInputShiftRegisterCounter = 32
		}
	}

	shouldPush := input.AutopushEnabled && (input.InputShiftRegisterCounter >= input.PushThreshold)
	output.ShouldStall = shouldPush && input.FIFORX.IsFull()
	if shouldPush && !input.FIFORX.IsFull() {
		err := input.FIFORX.Write(output.NewInputShiftRegister)
		if err != nil {
			return output, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
		}
		output.NewInputShiftRegister = 0
		output.NewInputShiftRegisterCounter = 0
	}

	return output, nil
}
