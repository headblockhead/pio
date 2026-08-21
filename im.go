package pio

import "errors"

const imSize = 32

type im struct {
	memory [imSize]uint16
}

func newIM() *im {
	return &im{}
}

var ErrIMOutOfBounds = errors.New("out of bounds")

func (i *im) read(address uint) (uint16, error) {
	if address >= imSize {
		return 0, ErrIMOutOfBounds
	}
	return i.memory[address], nil
}

func (i *im) write(address uint, value uint16) error {
	if address >= imSize {
		return ErrIMOutOfBounds
	}
	i.memory[address] = value
	return nil
}
