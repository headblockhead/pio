package pio

import "errors"

type Override uint

const (
	OverrideNone Override = iota
	OverrideInvert
	OverrideAlways0
	OverrideAlways1
)

var ErrOverrideInvalid = errors.New("invalid override")

func (o Override) applyTo(v bool) (bool, error) {
	switch o {
	case OverrideNone:
		return v, nil
	case OverrideInvert:
		return !v, nil
	case OverrideAlways0:
		return false, nil
	case OverrideAlways1:
		return true, nil
	}
	return false, ErrOverrideInvalid
}

type gpio struct {
	outputEnableOverride Override
	outputOverride       Override
	inputOverride        Override
	pioAssignment        uint
}

func newGPIO() *gpio {
	return &gpio{}
}
