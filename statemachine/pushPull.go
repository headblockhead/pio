package statemachine

import (
	"fmt"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type PushPullInput struct {
	AutopullEnabled            bool
	InputShiftRegister         uint32
	InputShiftRegisterCounter  uint8
	OutputShiftRegister        uint32
	OutputShiftRegisterCounter uint8
	PullThreshold              uint8
	PushThreshold              uint8
	XRegister                  uint32
	XRegisterInitialized       bool

	FIFOTX fifo.Reader
	FIFORX fifo.Writer
}

type PushPullOutput struct {
	NewInputShiftRegister         uint32
	NewInputShiftRegisterCounter  uint8
	NewOutputShiftRegister        uint32
	NewOutputShiftRegisterCounter uint8
	ShouldStall                   bool
}

func ExecutePushPull(instr instruction.PushPull, input PushPullInput) (PushPullOutput, error) {
	var output PushPullOutput = PushPullOutput{
		NewInputShiftRegister:         input.InputShiftRegister,
		NewInputShiftRegisterCounter:  input.InputShiftRegisterCounter,
		NewOutputShiftRegister:        input.OutputShiftRegister,
		NewOutputShiftRegisterCounter: input.OutputShiftRegisterCounter,
	}

	if instr.IsPull() {
		shouldPull := (!instr.IfThreshold() && !input.AutopullEnabled) || (input.OutputShiftRegisterCounter >= input.PullThreshold)
		output.ShouldStall = instr.Block() && shouldPull && input.FIFOTX.IsEmpty()
		if shouldPull && !input.FIFOTX.IsEmpty() {
			osr, err := input.FIFOTX.Read()
			if err != nil {
				return output, fmt.Errorf("error reading TX fifo for pull: %w", err)
			}
			output.NewOutputShiftRegister = osr
			output.NewOutputShiftRegisterCounter = 0
		} else if shouldPull && !instr.Block() {
			if !input.XRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			output.NewOutputShiftRegister = input.XRegister
			output.NewOutputShiftRegisterCounter = 0
		}
	} else {
		shouldPush := !instr.IfThreshold() || (input.InputShiftRegisterCounter >= input.PushThreshold)
		output.ShouldStall = instr.Block() && shouldPush && input.FIFORX.IsFull()
		if shouldPush && !input.FIFORX.IsFull() {
			err := input.FIFORX.Write(input.InputShiftRegister)
			if err != nil {
				return output, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			output.NewInputShiftRegister = 0
			output.NewInputShiftRegisterCounter = 0
		}
	}
	return output, nil
}
