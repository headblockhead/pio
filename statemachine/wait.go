package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type WaitInput struct {
}

type WaitOutput struct {
}

var ErrWaitSourceInvalid = errors.New("wait source invalid")

func ExecuteWait(instr instruction.Wait, input WaitInput, output WaitOutput) error {
	switch instr.Source() {
	case instruction.WaitSourceGPIO:
		stalled = ((sm.pinInputs>>index)&0b1 == 1) != polarity
	case instruction.WaitSourcePin:
		pin := (sm.baseInPin + index) % 32
		stalled = ((sm.pinInputs>>pin)&0b1 == 1) != polarity
	case instruction.WaitSourceIRQ:
		relative := (index>>4)&0b1 == 1
		irq := index & 0b111
		if relative {
			upperBit := irq & 0b100
			lowerBits := (irq + sm.index) & 0b011
			irq = upperBit | lowerBits
		}
		stalled = ((sm.irqInputs>>irq)&0b1 == 1) != polarity
		if !stalled && polarity {
			var clearIRQMask uint8 = (0b1 << irq)
			sm.irqWrites &= ^clearIRQMask
			sm.irqWritesMask |= clearIRQMask
		}
	default:
		return false, ErrWaitSourceInvalid
	}
	return stalled, nil
}
