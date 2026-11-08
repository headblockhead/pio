package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type WaitInput struct {
	SMIndex   uint8
	PinInputs uint32
	BaseInPin uint8
	IRQInputs uint8
}

type WaitOutput struct {
	ShouldStall    bool
	ShouldClearIRQ bool
	ClearIRQIndex  uint8
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
		relative := (instr.Index()>>4)&0b1 == 1
		irq := instr.Index() & 0b111
		if relative {
			upperBit := irq & 0b100
			lowerBits := (irq + input.SMIndex) & 0b011
			irq = upperBit | lowerBits
		}
		output.ShouldStall = ((input.IRQInputs>>irq)&0b1 == 1) != instr.Polarity()
		if !output.ShouldStall && instr.Polarity() {
			output.ShouldClearIRQ = true
			output.ClearIRQIndex = irq
		}
	default:
		return output, ErrWaitSourceInvalid
	}
	return output, nil
}
