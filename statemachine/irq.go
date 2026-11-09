package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instr"
)

type irqInput struct {
	alreadyStalled    bool
	irqInputs         uint8
	stateMachineIndex uint8
}

type irqOutput struct {
	clearIRQsMask uint8
	setIRQsMask   uint8
	shouldStall   bool
}

func getIRQIndex(index uint8, stateMachineIndex uint8) uint8 {
	relative := (index>>4)&0b1 == 1
	irq := index & 0b111
	if relative {
		upperBit := irq & 0b100
		lowerBits := (irq + stateMachineIndex) & 0b011
		irq = upperBit | lowerBits
	}
	return irq
}

var ErrCannotBothClearAndWaitOnAnIRQ = errors.New("cannot both clear and wait on an IRQ")

func executeIRQ(i instr.IRQ, input irqInput) (irqOutput, error) {
	var output irqOutput

	irq := getIRQIndex(i.Index(), input.stateMachineIndex)

	if input.alreadyStalled {
		output.shouldStall = ((input.irqInputs >> irq) & 0b1) == 1
		return output, nil
	}

	if i.Clear() {
		if i.Wait() {
			return output, ErrCannotBothClearAndWaitOnAnIRQ
		}
		output.clearIRQsMask = 0b1 << irq
	} else {
		output.setIRQsMask = 0b1 << irq
		if i.Wait() {
			output.shouldStall = true
		}
	}

	return output, nil
}
