package connection

import (
	"errors"
	"fmt"
)

type net struct {
	isHigh      bool
	connections map[string]Connection
}

func NewNet() *net {
	return &net{}
}

var ErrAlreadyConnected = errors.New("already connected")

func (n *net) connect(c Connection) error {
	_, exists := n.connections[c.ID()]
	if exists {
		return ErrAlreadyConnected
	}
	n.connections[c.ID()] = c
	return nil
}

var ErrInvalidConnectionState = errors.New("invalid connection state")
var ErrConflictingDrive = errors.New("net is driven with conflicting values")
var ErrConflictingPullups = errors.New("net is pulled with conflicting pulls")
var ErrFloating = errors.New("net is floating")

func (n *net) solve() error {
	previousState := n.isHigh

	drivenHigh := false
	drivenLow := false
	hasBusKeeper := false
	pulledUp := false
	pulledDown := false

	for _, c := range n.connections {
		state := c.GetState()
		switch state {
		case StateNone:
			// nothing
		case StateOutHigh:
			drivenHigh = true
		case StateOutLow:
			drivenLow = true
		case StateBusKeeper:
			hasBusKeeper = true
		case StatePullUp:
			pulledUp = true
		case StatePullDown:
			pulledDown = true
		default:
			return fmt.Errorf("connection %s: %w", c.ID(), ErrInvalidConnectionState)
		}
	}

	if drivenHigh && drivenLow {
		return ErrConflictingDrive
	}
	if pulledUp && pulledDown {
		return ErrConflictingPullups
	}

	if drivenHigh {
		n.isHigh = true
	}
	if drivenLow {
		n.isHigh = false
	}
	if !drivenHigh && !drivenLow {
		if pulledUp {
			n.isHigh = true
		}
		if pulledDown {
			n.isHigh = false
		}
		if !pulledUp && !pulledDown {
			if hasBusKeeper {
				n.isHigh = previousState
			} else {
				return ErrFloating
			}
		}
	}

	for _, c := range n.connections {
		c.SetInput(n.isHigh)
	}

	return nil
}
