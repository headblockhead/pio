package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type IRQInput struct {
	AlreadyStalled    bool
	IRQInputs         uint8
	StateMachineIndex uint8
}

type IRQOutput struct {
	ClearIRQsMask uint8
	SetIRQsMask   uint8
	ShouldStall   bool
}

func GetIRQIndex(index uint8, stateMachineIndex uint8) uint8 {
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

func ExecuteIRQ(instr instruction.IRQ, input IRQInput) (IRQOutput, error) {
	var output IRQOutput

	irq := GetIRQIndex(instr.Index(), input.StateMachineIndex)

	if input.AlreadyStalled {
		output.ShouldStall = ((input.IRQInputs >> irq) & 0b1) == 1
		return output, nil
	}

	if instr.Clear() {
		if instr.Wait() {
			return output, ErrCannotBothClearAndWaitOnAnIRQ
		}
		output.ClearIRQsMask = 0b1 << irq
	} else {
		output.SetIRQsMask = 0b1 << irq
		if instr.Wait() {
			output.ShouldStall = true
		}
	}

	return output, nil
}
