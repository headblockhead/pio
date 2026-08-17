package pio

import "errors"

type IRQState uint

const (
	IRQCleared IRQState = iota
	IRQSet
)

type IRQSMs interface {
	Set(irq uint) error
	Clear(irq uint) error
	Read(irq uint) (state IRQState, err error)
}

const irqCount = 8

type IRQs struct {
	irqs [irqCount]IRQState
}

func NewIRQs() *IRQs {
	return &IRQs{}
}

var ErrIRQOutOfBounds = errors.New("out of bounds")

func (i *IRQs) Set(irq uint) error {
	if irq >= irqCount {
		return ErrIRQOutOfBounds
	}
	i.irqs[irq] = IRQSet
	return nil
}

func (i *IRQs) Clear(irq uint) error {
	if irq >= irqCount {
		return ErrIRQOutOfBounds
	}
	i.irqs[irq] = IRQCleared
	return nil
}

func (i *IRQs) Read(irq uint) (state IRQState, err error) {
	if irq >= irqCount {
		return IRQCleared, ErrIRQOutOfBounds
	}
	return i.irqs[irq], nil
}
