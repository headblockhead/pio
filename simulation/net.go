package simulation

import (
	"errors"
	"fmt"
)

type NetObserver interface {
	Component

	LogicLevel() bool
	Connections() map[ComponentIdentifier]Connection
}

type NetConfigurator interface {
	SetLabel(string)
	Connect(Connection) error
	Disconnect(Connection) error
}

type NetOperator interface {
	Solve() error
}

type Net struct {
	id          ComponentIdentifier
	label       string
	logicLevel  bool
	connections map[ComponentIdentifier]Connection
}

func NewNet(label string) *Net {
	return &Net{
		id:          NewComponentIdentifier(),
		label:       label,
		logicLevel:  false,
		connections: make(map[ComponentIdentifier]Connection),
	}
}

func (n *Net) Component() Component {
	return n
}

func (n *Net) ID() ComponentIdentifier { return n.id }
func (n *Net) Label() string           { return n.label }

func (n *Net) Observer() NetObserver {
	return n
}

func (n *Net) LogicLevel() bool                                { return n.logicLevel }
func (n *Net) Connections() map[ComponentIdentifier]Connection { return n.connections }

func (n *Net) Configurator() NetConfigurator {
	return n
}

func (n *Net) SetLabel(label string) { n.label = label }

var ErrNetConnectionAlreadyConnected = errors.New("already connected")

func (n *Net) Connect(c Connection) error {
	if _, exists := n.connections[c.ID()]; exists {
		return fmt.Errorf("connection [%v]: %w", c.Label(), ErrNetConnectionAlreadyConnected)
	}
	n.connections[c.ID()] = c
	return nil
}

var ErrNetConnectionNotCurrentlyConnected = errors.New("not currently connected")

func (n *Net) Disconnect(c Connection) error {
	if _, exists := n.connections[c.ID()]; !exists {
		return fmt.Errorf("connection [%v]: %w", c.Label(), ErrNetConnectionNotCurrentlyConnected)
	}
	delete(n.connections, c.ID())
	return nil
}

func (n *Net) Operator() NetOperator {
	return n
}

var ErrNetSolveConnectionStateInvalid = errors.New("connection state invalid")
var ErrNetSolveConflictingDrive = errors.New("conflicting drive")
var ErrNetSolveConflictingPulls = errors.New("conflicting pulls")
var ErrNetSolveFloating = errors.New("floating")

func (n *Net) Solve() error {
	previousLogicLevel := n.logicLevel

	var drivenHighBy Connection
	drivenHigh := false
	var drivenLowBy Connection
	drivenLow := false
	hasBusKeeper := false
	var pulledUpBy Connection
	pulledUp := false
	var pulledDownBy Connection
	pulledDown := false

	for _, c := range n.connections {
		state := c.State()
		switch state {
		case ConnectionStateNone:
			// nothing
		case ConnectionStateOutHigh:
			drivenHighBy = c
			drivenHigh = true
		case ConnectionStateOutLow:
			drivenLowBy = c
			drivenLow = true
		case ConnectionStateBusKeeper:
			hasBusKeeper = true
		case ConnectionStatePullUp:
			pulledUpBy = c
			pulledUp = true
		case ConnectionStatePullDown:
			pulledDownBy = c
			pulledDown = true
		default:
			return fmt.Errorf("connection [%v]: %w: %d", c.Label(), ErrNetSolveConnectionStateInvalid, state)
		}
	}

	if drivenHigh && drivenLow {
		return fmt.Errorf("connection [%v] and connection [%v]: %w", drivenHighBy.Label(), drivenLowBy.Label(), ErrNetSolveConflictingDrive)
	}
	if pulledUp && pulledDown {
		return fmt.Errorf("connection [%v] and connection [%v]: %w", pulledUpBy.Label(), pulledDownBy.Label(), ErrNetSolveConflictingPulls)
	}

	if drivenHigh {
		n.logicLevel = true
	}
	if drivenLow {
		n.logicLevel = false
	}
	if !drivenHigh && !drivenLow {
		if pulledUp {
			n.logicLevel = true
		}
		if pulledDown {
			n.logicLevel = false
		}
		if !pulledUp && !pulledDown {
			if hasBusKeeper {
				n.logicLevel = previousLogicLevel
			} else {
				return ErrNetSolveFloating
			}
		}
	}

	for _, c := range n.connections {
		c.SetInput(n.logicLevel)
	}

	return nil
}
