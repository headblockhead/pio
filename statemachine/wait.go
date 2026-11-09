package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instr"
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

func executeWait(i instr.Wait, input waitInput) (waitOutput, error) {
	var output waitOutput

	switch i.Source() {
	case instr.WaitSourceGPIO:
		output.shouldStall = ((input.pinInputs>>i.Index())&0b1 == 1) != i.Polarity()
	case instr.WaitSourcePin:
		pin := (input.baseInPin + i.Index()) % 32
		output.shouldStall = ((input.pinInputs>>pin)&0b1 == 1) != i.Polarity()
	case instr.WaitSourceIRQ:
		irq := getIRQIndex(i.Index(), input.stateMachineIndex)
		output.shouldStall = ((input.irqInputs>>irq)&0b1 == 1) != i.Polarity()
		if !output.shouldStall && i.Polarity() {
			output.clearIRQsMask = 0b1 << irq
		}
	default:
		return output, ErrWaitSourceInvalid
	}
	return output, nil
}
