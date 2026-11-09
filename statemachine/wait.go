package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type waitInput struct {
	baseInPin         uint8
	irqInputs         uint8
	pinInputs         uint32
	stateMachineIndex uint8
}

type waitOutput struct {
	clearIRQsMask uint8
	shouldStall   bool
}

var ErrWaitSourceInvalid = errors.New("wait source invalid")

func executeWait(instr instruction.Wait, input waitInput) (waitOutput, error) {
	var output waitOutput

	switch instr.Source() {
	case instruction.WaitSourceGPIO:
		output.shouldStall = ((input.pinInputs>>instr.Index())&0b1 == 1) != instr.Polarity()
	case instruction.WaitSourcePin:
		pin := (input.baseInPin + instr.Index()) % 32
		output.shouldStall = ((input.pinInputs>>pin)&0b1 == 1) != instr.Polarity()
	case instruction.WaitSourceIRQ:
		irq := getIRQIndex(instr.Index(), input.stateMachineIndex)
		output.shouldStall = ((input.irqInputs>>irq)&0b1 == 1) != instr.Polarity()
		if !output.shouldStall && instr.Polarity() {
			output.clearIRQsMask = 0b1 << irq
		}
	default:
		return output, ErrWaitSourceInvalid
	}
	return output, nil
}
