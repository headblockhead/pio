package pio

import "errors"

type InstructionMemoryWriter interface {
	Write(address uint, value uint16) error
}

type InstructionMemoryReader interface {
	Read(address uint) (value uint16, err error)
}

type InstructionMemory struct {
	memory [32]uint16
}

func NewInstructionMemory() *InstructionMemory {
	return &InstructionMemory{
		memory: [32]uint16{},
	}
}

var IMOutOfBounds = errors.New("Out of bounds")

func (im *InstructionMemory) Read(address uint) (value uint16, err error) {
	if address >= 32 {
		return 0, IMOutOfBounds
	}
	return im.memory[address], nil
}

func (im *InstructionMemory) Write(address uint, value uint16) error {
	if address >= 32 {
		return IMOutOfBounds
	}
	im.memory[address] = value
	return nil
}
