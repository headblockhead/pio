package statemachine

import (
	"fmt"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type pushPullInput struct {
	autopullEnabled            bool
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint8
	outputShiftRegister        uint32
	outputShiftRegisterCounter uint8
	pullThreshold              uint8
	pushThreshold              uint8
	xRegister                  uint32
	xRegisterInitialized       bool

	fifoTX fifo.Reader
	fifoRX fifo.Writer
}

type pushPullOutput struct {
	newInputShiftRegister         uint32
	newInputShiftRegisterCounter  uint8
	newOutputShiftRegister        uint32
	newOutputShiftRegisterCounter uint8
	shouldStall                   bool
}

func executePushPull(instr instruction.PushPull, input pushPullInput) (pushPullOutput, error) {
	var output pushPullOutput = pushPullOutput{
		newInputShiftRegister:         input.inputShiftRegister,
		newInputShiftRegisterCounter:  input.inputShiftRegisterCounter,
		newOutputShiftRegister:        input.outputShiftRegister,
		newOutputShiftRegisterCounter: input.outputShiftRegisterCounter,
	}

	if instr.IsPull() {
		shouldPull := (!instr.IfThreshold() && !input.autopullEnabled) || (input.outputShiftRegisterCounter >= input.pullThreshold)
		output.shouldStall = instr.Block() && shouldPull && input.fifoTX.IsEmpty()
		if shouldPull && !input.fifoTX.IsEmpty() {
			osr, err := input.fifoTX.Read()
			if err != nil {
				return output, fmt.Errorf("error reading TX fifo for pull: %w", err)
			}
			output.newOutputShiftRegister = osr
			output.newOutputShiftRegisterCounter = 0
		} else if shouldPull && !instr.Block() {
			if !input.xRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			output.newOutputShiftRegister = input.xRegister
			output.newOutputShiftRegisterCounter = 0
		}
	} else {
		shouldPush := !instr.IfThreshold() || (input.inputShiftRegisterCounter >= input.pushThreshold)
		output.shouldStall = instr.Block() && shouldPush && input.fifoRX.IsFull()
		if shouldPush && !input.fifoRX.IsFull() {
			err := input.fifoRX.Write(input.inputShiftRegister)
			if err != nil {
				return output, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			output.newInputShiftRegister = 0
			output.newInputShiftRegisterCounter = 0
		}
	}
	return output, nil
}
