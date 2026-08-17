package pio

import "errors"

type PadPull uint
type PadOutputDisable uint
type PadOutputEnable uint
type PadInputEnable uint
type PadOutput uint
type PadInput uint
type PadState uint

const (
	PadPullDown PadPull = iota
	PadPullBusKeeper
	PadPullNone
	PadPullUp

	PadOutputNotDisabled PadOutputDisable = iota
	PadOutputDisabled

	PadOutputNotEnabled PadOutputEnable = iota
	PadOutputEnabled

	PadInputEnabled PadInputEnable = iota
	PadInputNotEnabled

	PadOutputLow PadOutput = iota
	PadOutputHigh

	PadInputLow PadInput = iota
	PadInputHigh

	PadStateNone PadState = iota
	PadStateDrivenHigh
	PadStateDrivenLow
	PadStatePulledUp
	PadStatePulledDown
)

type PadsGPIOs interface {
	SetOutput(pad uint, output PadOutput) error
	SetOutputEnable(pad uint, outputEnable PadOutputEnable) error
	GetInput(pad uint) (PadInput, error)
}

type PadsSolver interface {
	GetState(pad uint) (PadState, error)
	SetInput(pad uint, input PadInput) error
}

const padCount = 30

type Pads struct {
	pulls [padCount]PadPull

	outputDisables [padCount]PadOutputDisable
	outputEnables  [padCount]PadOutputEnable
	outputs        [padCount]PadOutput

	inputEnables [padCount]PadInputEnable
	inputs       [padCount]PadInput
}

func NewPads() *Pads {
	return &Pads{}
}

var ErrPadOutOfBounds = errors.New("out of bounds")

func (p *Pads) SetOutput(pad uint, output PadOutput) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.outputs[pad] = output
	return nil
}

func (p *Pads) SetOutputEnable(pad uint, outputEnable PadOutputEnable) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.outputEnables[pad] = outputEnable
	return nil
}

func (p *Pads) GetInput(pad uint) (PadInput, error) {
	if pad >= padCount {
		return PadInputLow, ErrPadOutOfBounds
	}
	if p.inputEnables[pad] == PadInputEnabled {
		return p.inputs[pad], nil
	} else {
		return PadInputLow, nil
	}
}

var ErrPadsInvalidPull = errors.New("invalid pull")
var ErrPadsInvalidInput = errors.New("invalid input")

func (p *Pads) GetState(pad uint) (PadState, error) {
	if pad >= padCount {
		return PadStateNone, ErrPadOutOfBounds
	}
	if p.outputEnables[pad] == PadOutputEnabled && p.outputDisables[pad] == PadOutputNotDisabled {
		if p.outputs[pad] == PadOutputHigh {
			return PadStateDrivenHigh, nil
		} else {
			return PadStateDrivenLow, nil
		}
	} else {
		switch p.pulls[pad] {
		case PadPullBusKeeper:
			switch p.inputs[pad] {
			case PadInputHigh:
				return PadStatePulledUp, nil
			case PadInputLow:
				return PadStatePulledDown, nil
			default:
				return PadStateNone, ErrPadsInvalidInput
			}
		case PadPullUp:
			return PadStatePulledUp, nil
		case PadPullDown:
			return PadStatePulledDown, nil
		case PadPullNone:
			return PadStateNone, nil
		default:
			return PadStateNone, ErrPadsInvalidPull
		}
	}
}

func (p *Pads) SetInput(pad uint, input PadInput) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.inputs[pad] = input
	return nil
}

func (p *Pads) SetPull(pad uint, pull PadPull) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.pulls[pad] = pull
	return nil
}

func (p *Pads) GetPull(pad uint) (PadPull, error) {
	if pad >= padCount {
		return PadPullNone, ErrPadOutOfBounds
	}
	return p.pulls[pad], nil
}

func (p *Pads) SetOutputDisable(pad uint, outputDisable PadOutputDisable) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.outputDisables[pad] = outputDisable
	return nil
}

func (p *Pads) GetOutputDisable(pad uint) (PadOutputDisable, error) {
	if pad >= padCount {
		return PadOutputNotDisabled, ErrPadOutOfBounds
	}
	return p.outputDisables[pad], nil
}

func (p *Pads) SetInputEnable(pad uint, inputEnable PadInputEnable) error {
	if pad >= padCount {
		return ErrPadOutOfBounds
	}
	p.inputEnables[pad] = inputEnable
	return nil
}

func (p *Pads) GetInputEnable(pad uint) (PadInputEnable, error) {
	if pad >= padCount {
		return PadInputEnabled, ErrPadOutOfBounds
	}
	return p.inputEnables[pad], nil
}
