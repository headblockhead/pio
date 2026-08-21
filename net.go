package pio

import (
	"errors"
	"fmt"
)

type Connection interface {
	ID() string
	GetState() PadState
	SetInput(bool)
}

type net struct {
	state       bool
	connections map[string]Connection
}

var ErrInvalidPadState = errors.New("invalid pad state")
var ErrConflictingDrive = errors.New("net is driven with conflicting values")
var ErrConflictingPullups = errors.New("net is pulled with conflicting pulls")
var ErrFloating = errors.New("net is floating")

func (n *net) solve() error {
	previousState := n.state

	drivenHigh := false
	drivenLow := false
	hasBusKeeper := false
	pulledUp := false
	pulledDown := false

	for _, c := range n.connections {
		state := c.GetState()
		switch state {
		case PadNone:
			// nothing
		case PadOutHigh:
			drivenHigh = true
		case PadOutLow:
			drivenLow = true
		case PadBusKeeper:
			hasBusKeeper = true
		case PadPullUp:
			pulledUp = true
		case PadPullDown:
			pulledDown = true
		default:
			return fmt.Errorf("connection %s: %w", c.ID(), ErrInvalidPadState)
		}
	}

	if drivenHigh && drivenLow {
		return ErrConflictingDrive
	}
	if pulledUp && pulledDown {
		return ErrConflictingPullups
	}

	if drivenHigh {
		n.state = true
	}
	if drivenLow {
		n.state = false
	}
	if !drivenHigh && !drivenLow {
		if pulledUp {
			n.state = true
		}
		if pulledDown {
			n.state = false
		}
		if !pulledUp && !pulledDown {
			if hasBusKeeper {
				n.state = previousState
			} else {
				return ErrFloating
			}
		}
	}

	for _, c := range n.connections {
		c.SetInput(n.state)
	}

	return nil
}
