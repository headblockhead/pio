package gpio

import (
	"errors"
	"fmt"
)

type Override uint

const (
	OverrideNone Override = iota
	OverrideInvert
	OverrideAlways0
	OverrideAlways1
)

var ErrOverrideInvalid = errors.New("override invalid")

func (o Override) ApplyTo(v bool) (bool, error) {
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
	return v, fmt.Errorf("%w: %d", ErrOverrideInvalid, o)
}
