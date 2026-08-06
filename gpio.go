package pio

import "errors"

type GPIOReadWriter interface {
	SetOutputEnable(gpio uint, outputEnable bool) error
	SetOutput(gpio uint, output bool) error
	GetInput(gpio uint) (value bool, err error)
}

type GPIOStateSyncer interface {
	GetState() (outputs [32]bool, outputEnables [32]bool)
	SetInputs(inputs [32]bool)
}

type GPIO struct {
	outputs       [32]bool
	outputEnables [32]bool
	inputs        [32]bool
}

var GPIOOutOfBounds = errors.New("Out of bounds")

func (g *GPIO) SetOutputEnable(gpio uint, outputEnable bool) error {
	if gpio >= 32 {
		return GPIOOutOfBounds
	}
	g.outputEnables[gpio] = outputEnable
	return nil
}

func (g *GPIO) SetOutput(gpio uint, output bool) error {
	if gpio >= 32 {
		return GPIOOutOfBounds
	}
	g.outputs[gpio] = output
	return nil
}

func (g *GPIO) GetInput(gpio uint) (value bool, err error) {
	if gpio >= 32 {
		return false, GPIOOutOfBounds
	}
	return g.inputs[gpio], nil
}

func (g *GPIO) GetState() (outputs [32]bool, outputEnables [32]bool) {
	return g.outputs, g.outputEnables
}

func (g *GPIO) SetInputs(inputs [32]bool) {
	g.inputs = inputs
}
