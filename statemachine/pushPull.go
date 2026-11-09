package statemachine

import (
	"fmt"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instr"
)

type pushPullInput struct {
	autopullEnabled            bool
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	pullThreshold              uint
	pushThreshold              uint
	xRegister                  uint32
	xRegisterInitialized       bool

	fifoTX fifo.Reader
	fifoRX fifo.Writer
}

type pushPullOutput struct {
	newInputShiftRegister         uint32
	newInputShiftRegisterCounter  uint
	newOutputShiftRegister        uint32
	newOutputShiftRegisterCounter uint
	shouldStall                   bool
}

var ErrPushNoblockOnFullFIFO = fmt.Errorf("push noblock instruction on full RX FIFO caused data loss")

func executePushPull(i instr.PushPull, input pushPullInput) (pushPullOutput, error) {
	var output = pushPullOutput{
		newInputShiftRegister:         input.inputShiftRegister,
		newInputShiftRegisterCounter:  input.inputShiftRegisterCounter,
		newOutputShiftRegister:        input.outputShiftRegister,
		newOutputShiftRegisterCounter: input.outputShiftRegisterCounter,
	}

	if i.IsPull() {
		shouldPull := (!i.IfThreshold() && !input.autopullEnabled) || (input.outputShiftRegisterCounter >= input.pullThreshold)
		output.shouldStall = i.Block() && shouldPull && input.fifoTX.IsEmpty()
		if shouldPull && !input.fifoTX.IsEmpty() {
			osr, err := input.fifoTX.Read()
			if err != nil {
				return output, fmt.Errorf("error reading TX fifo for pull: %w", err)
			}
			output.newOutputShiftRegister = osr
			output.newOutputShiftRegisterCounter = 0
		} else if shouldPull && !i.Block() {
			if !input.xRegisterInitialized {
				return output, ErrXRegisterNotInitialized
			}
			output.newOutputShiftRegister = input.xRegister
			output.newOutputShiftRegisterCounter = 0
		}
	} else {
		shouldPush := !i.IfThreshold() || (input.inputShiftRegisterCounter >= input.pushThreshold)
		output.shouldStall = i.Block() && shouldPush && input.fifoRX.IsFull()
		if shouldPush && !input.fifoRX.IsFull() {
			err := input.fifoRX.Write(input.inputShiftRegister)
			if err != nil {
				return output, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			output.newInputShiftRegister = 0
			output.newInputShiftRegisterCounter = 0
		} else if shouldPush && !i.Block() {
			return output, ErrPushNoblockOnFullFIFO
		}
	}
	return output, nil
}
