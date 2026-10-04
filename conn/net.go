package conn

import (
	"errors"
	"fmt"
)

type net struct {
	isHigh      bool
	connections map[string]Connection
}

func newNet() *net {
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

var ErrNotCurrentlyConnected = errors.New("not currently connected")

func (n *net) disconnect(c Connection) error {
	_, exists := n.connections[c.ID()]
	if !exists {
		return ErrNotCurrentlyConnected
	}
	delete(n.connections, c.ID())
	return nil
}

var ErrSolveConnectionStateInvalid = errors.New("connection state invalid")
var ErrSolveConflictingDrive = errors.New("conflicting drive")
var ErrSolveConflictingPulls = errors.New("conflicting pulls")
var ErrSolveFloating = errors.New("floating")

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
			return fmt.Errorf("connection %s: %w: %d", c.ID(), ErrSolveConnectionStateInvalid, state)
		}
	}

	if drivenHigh && drivenLow {
		return ErrSolveConflictingDrive
	}
	if pulledUp && pulledDown {
		return ErrSolveConflictingPulls
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
				return ErrSolveFloating
			}
		}
	}

	for _, c := range n.connections {
		c.SetInput(n.isHigh)
	}

	return nil
}
