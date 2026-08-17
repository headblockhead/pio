package pio

import "errors"
import "fmt"

type CircuitSolver interface {
	Solve() error
}

type Circuit struct {
	connections []ConnectionCircuit
}

func NewCircuit() *Circuit {
	return &Circuit{}
}

func (c *Circuit) AddConnection(connection ConnectionCircuit) error {
	c.connections = append(c.connections, connection)
	return nil
}

var ErrCircuitConflictingDrive = errors.New("a connection is driven low and high simultaneously, creating a short circuit")
var ErrCircuitConflictingPullups = errors.New("a connection is pulled up and down simultaneously, which is likely unintentional")
var ErrCircuitPadFloating = errors.New("a connection is left floating, which is likely unintentional")
var ErrCircuitInvalidConnectionPadState = errors.New("invalid connection pad state")

func (c *Circuit) Solve() error {
	drivenHigh := false
	drivenLow := false
	pulledUp := false
	pulledDown := false
	for i, connection := range c.connections {
		state, err := connection.GetState()
		if err != nil {
			return fmt.Errorf("error getting state of connection %d: %w", i, err)
		}
		switch state {
		case PadStateDrivenHigh:
			if drivenLow {
				return ErrCircuitConflictingDrive
			}
			drivenHigh = true
		case PadStateDrivenLow:
			if drivenHigh {
				return ErrCircuitConflictingDrive
			}
			drivenLow = true
		case PadStatePulledUp:
			if pulledDown {
				return ErrCircuitConflictingPullups
			}
			pulledUp = true
		case PadStatePulledDown:
			if pulledUp {
				return ErrCircuitConflictingPullups
			}
			pulledDown = true
		default:
			return ErrCircuitInvalidConnectionPadState
		}
	}
	var solvedValue PadInput
	if drivenHigh {
		solvedValue = PadInputHigh
	} else if drivenLow {
		solvedValue = PadInputLow
	} else if pulledUp {
		solvedValue = PadInputHigh
	} else if pulledDown {
		solvedValue = PadInputLow
	} else {
		return ErrCircuitPadFloating
	}
	for i, connection := range c.connections {
		err := connection.SetInput(solvedValue)
		if err != nil {
			return fmt.Errorf("error setting input of connection %d: %w", i, err)
		}
	}
	return nil
}
