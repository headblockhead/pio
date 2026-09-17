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

var ErrNetConnectionAlreadyConnected = errors.New("connection is already connected to this net")

func (n *net) connect(c Connection) error {
	_, exists := n.connections[c.ID()]
	if exists {
		return ErrNetConnectionAlreadyConnected
	}
	n.connections[c.ID()] = c
	return nil
}

var ErrNetConnectionNotCurrentlyConnected = errors.New("connection is not currently connected to this net")

func (n *net) disconnect(c Connection) error {
	_, exists := n.connections[c.ID()]
	if !exists {
		return ErrNetConnectionNotCurrentlyConnected
	}
	delete(n.connections, c.ID())
	return nil
}

var ErrSolveInvalidConnectionState = errors.New("invalid connection state")
var ErrSolveConflictingDrive = errors.New("net is driven with conflicting values by at least two connections")
var ErrSolveConflictingPullups = errors.New("net is pulled with conflicting pulls by at least two connections")
var ErrSolveFloating = errors.New("net is floating, which is likely unintentional")

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
			return fmt.Errorf("connection %s: %w", c.ID(), ErrSolveInvalidConnectionState)
		}
	}

	if drivenHigh && drivenLow {
		return ErrSolveConflictingDrive
	}
	if pulledUp && pulledDown {
		return ErrSolveConflictingPullups
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
