package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type WaitInput struct {
	BaseInPin         uint8
	IRQInputs         uint8
	PinInputs         uint32
	StateMachineIndex uint8
}

type WaitOutput struct {
	ClearIRQsMask uint8
	ShouldStall   bool
}

var ErrWaitSourceInvalid = errors.New("wait source invalid")

func ExecuteWait(instr instruction.Wait, input WaitInput) (WaitOutput, error) {
	var output WaitOutput

	switch instr.Source() {
	case instruction.WaitSourceGPIO:
		output.ShouldStall = ((input.PinInputs>>instr.Index())&0b1 == 1) != instr.Polarity()
	case instruction.WaitSourcePin:
		pin := (input.BaseInPin + instr.Index()) % 32
		output.ShouldStall = ((input.PinInputs>>pin)&0b1 == 1) != instr.Polarity()
	case instruction.WaitSourceIRQ:
		irq := GetIRQIndex(instr.Index(), input.StateMachineIndex)
		output.ShouldStall = ((input.IRQInputs>>irq)&0b1 == 1) != instr.Polarity()
		if !output.ShouldStall && instr.Polarity() {
			output.ClearIRQsMask = 0b1 << irq
		}
	default:
		return output, ErrWaitSourceInvalid
	}
	return output, nil
}
