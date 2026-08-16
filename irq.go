package pio

import "errors"

type IRQState bool

const (
	IRQSet     IRQState = true
	IRQCleared          = false
)

type IRQSetClearReader interface {
	Set(irq uint) error
	Clear(irq uint) error
	Read(irq uint) (state IRQState, err error)
}

type IRQ struct {
	irqs [8]IRQState
}

func NewIRQ() *IRQ {
	return &IRQ{
		irqs: [8]IRQState{},
	}
}

var IRQOutOfBounds = errors.New("out of bounds")

func (i *IRQ) Set(irq uint) error {
	if irq >= 8 {
		return IRQOutOfBounds
	}
	i.irqs[irq] = IRQSet
	return nil
}

func (i *IRQ) Clear(irq uint) error {
	if irq >= 8 {
		return IRQOutOfBounds
	}
	i.irqs[irq] = IRQCleared
	return nil
}

func (i *IRQ) Read(irq uint) (state IRQState, err error) {
	if irq >= 8 {
		return false, IRQOutOfBounds
	}
	return i.irqs[irq], nil
}
