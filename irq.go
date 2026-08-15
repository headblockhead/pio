package pio

import "errors"

type IRQReadWriter interface {
	Set(irq uint) error
	Clear(irq uint) error
	Read(irq uint) (state bool, err error)
}

type IRQ struct {
	irqs [8]bool
}

func NewIRQ() *IRQ {
	return &IRQ{
		irqs: [8]bool{},
	}
}

var IRQOutOfBounds = errors.New("out of bounds")

func (i *IRQ) Set(irq uint) error {
	if irq >= 8 {
		return IRQOutOfBounds
	}
	i.irqs[irq] = true
	return nil
}

func (i *IRQ) Clear(irq uint) error {
	if irq >= 8 {
		return IRQOutOfBounds
	}
	i.irqs[irq] = false
	return nil
}

func (i *IRQ) Read(irq uint) (state bool, err error) {
	if irq >= 8 {
		return false, IRQOutOfBounds
	}
	return i.irqs[irq], nil
}
