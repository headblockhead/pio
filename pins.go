package pio

import "errors"

type PinOutput uint
type PinInput uint
type PinOutputEnable uint

const (
	PinOutputLow PinOutput = iota
	PinOutputHigh

	PinOutputNotEnabled PinOutputEnable = iota
	PinOutputEnabled

	PinInputLow PinInput = iota
	PinInputHigh
)

type PinsSMs interface {
	SetOutput(pin uint, output PinOutput) error
	SetOutputEnable(pin uint, outputEnable PinOutputEnable) error
	GetInput(pin uint) (PinInput, error)
}

type PinsGPIOs interface {
	GetOutput(pin uint) (PinOutput, error)
	GetOutputEnable(pin uint) (PinOutputEnable, error)
	SetInput(pin uint, input PinInput) error
}

const pinCount = 32

type Pins struct {
	outputs       [pinCount]PinOutput
	outputEnables [pinCount]PinOutputEnable
	inputs        [pinCount]PinInput
}

func NewPins() *Pins {
	return &Pins{}
}

var ErrPinOutOfBounds = errors.New("out of bounds")

func (p *Pins) SetOutput(pin uint, output PinOutput) error {
	if pin >= pinCount {
		return ErrPinOutOfBounds
	}
	p.outputs[pin] = output
	return nil
}

func (p *Pins) SetOutputEnable(pin uint, outputEnable PinOutputEnable) error {
	if pin >= pinCount {
		return ErrPinOutOfBounds
	}
	p.outputEnables[pin] = outputEnable
	return nil
}

func (p *Pins) GetInput(pin uint) (PinInput, error) {
	if pin >= pinCount {
		return PinInputLow, ErrPinOutOfBounds
	}
	return p.inputs[pin], nil
}

func (p *Pins) GetOutput(pin uint) (PinOutput, error) {
	if pin >= pinCount {
		return PinOutputLow, ErrPinOutOfBounds
	}
	return p.outputs[pin], nil
}

func (p *Pins) GetOutputEnable(pin uint) (PinOutputEnable, error) {
	if pin >= pinCount {
		return PinOutputNotEnabled, ErrPinOutOfBounds
	}
	return p.outputEnables[pin], nil
}

func (p *Pins) SetInput(pin uint, input PinInput) error {
	if pin >= pinCount {
		return ErrPinOutOfBounds
	}
	p.inputs[pin] = input
	return nil
}
