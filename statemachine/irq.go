package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type IRQInput struct {
	SMIndex        uint8
	AlreadyStalled bool
	IRQInputs      uint8
}

type IRQOutput struct {
	ShouldStall    bool
	ShouldClearIRQ bool
	ShouldSetIRQ   bool
	IRQIndex       uint8
}

var ErrCannotBothClearAndWaitOnAnIRQ = errors.New("cannot both clear and wait on an IRQ")

func ExecuteIRQ(instr instruction.IRQ, input IRQInput) (IRQOutput, error) {
	var output IRQOutput

	relative := (instr.Index()>>4)&0b1 == 1
	irq := instr.Index() & 0b111
	if relative {
		upperBit := irq & 0b100
		lowerBits := (irq + input.SMIndex) & 0b011
		irq = upperBit | lowerBits
	}

	if input.AlreadyStalled {
		output.ShouldStall = ((input.IRQInputs >> irq) & 0b1) == 1
		return output, nil
	}

	output.IRQIndex = irq
	if instr.Clear() {
		if instr.Wait() {
			return output, ErrCannotBothClearAndWaitOnAnIRQ
		}
		output.ShouldClearIRQ = true
	} else {
		output.ShouldSetIRQ = true
		if instr.Wait() {
			output.ShouldStall = true
		}
	}

	return output, nil
}
