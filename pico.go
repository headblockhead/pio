package pio

import "errors"

type PicoChip uint

const (
	ChipRP2040 PicoChip = iota
)

type Pico interface {
	ID() string

	// Advance the system clock by one tick.
	Tick() error

	PadConnection(pad uint) (Connection, error)

	SetPadPullUp(pad uint, pu bool) error
	SetPadPullDown(pad uint, pd bool) error
	SetPadOutputDisabled(pad uint, od bool) error
	SetPadInputEnabled(pad uint, ie bool) error
	SetPadInputSynchroniserEnabled(pad uint, ise bool) error

	SetGPIOOutputEnableOverride(gpio uint, o Override) error
	SetGPIOOutputOverride(gpio uint, o Override) error
	SetGPIOInputOverride(gpio uint, o Override) error
	SetGPIOAssignment(gpio uint, pio uint) error

	// SetSM...
}

var ErrInvalidPicoChip = errors.New("invalid chip type")

func NewPico(chip PicoChip, id string) (Pico, error) {
	switch chip {
	case ChipRP2040:
		return newRP2040(id), nil
	default:
		return nil, ErrInvalidPicoChip
	}
}
