package pad

import (
	"errors"
	"fmt"
	"slices"

	"git.edwardh.dev/pio/simulation"
)

type Observer interface {
	simulation.Component

	PulledUp() bool
	PulledDown() bool
	OutputDisabled() bool
	InputEnabled() bool

	OutputEnabled() bool
	Output() bool

	StateHistory() []simulation.ConnectionState
	InputHistory() []bool
}

type Configurator interface {
	SetLabel(string)

	SetPullUp(bool)
	SetPullDown(bool)
	SetOutputDisabled(bool)
	SetInputEnabled(bool)
	SetOutputDelayCycles(uint) error
	SetInputDelayCycles(uint) error
}

type Operator interface {
	simulation.Component
	simulation.Ticker

	SetOutputEnabled(bool)
	SetOutput(bool)
	GetInput() bool
}

func padState(shouldOutput bool, outputValue bool, pullUp bool, pullDown bool) simulation.ConnectionState {
	if shouldOutput {
		if outputValue {
			return simulation.ConnectionStateOutHigh
		} else {
			return simulation.ConnectionStateOutLow
		}
	} else {
		if pullUp && pullDown {
			return simulation.ConnectionStateBusKeeper
		}
		if pullUp {
			return simulation.ConnectionStatePullUp
		}
		if pullDown {
			return simulation.ConnectionStatePullDown
		}
		return simulation.ConnectionStateNone
	}
}

type Pad struct {
	id    simulation.ComponentIdentifier
	label string

	pullUp         bool
	pullDown       bool
	outputDisabled bool
	inputEnabled   bool

	outputEnabled bool
	output        bool

	stateHistory []simulation.ConnectionState
	inputHistory []bool
}

func New(label string) *Pad {
	return &Pad{
		id:    simulation.NewComponentIdentifier(),
		label: label,

		pullUp:         false,
		pullDown:       true,
		outputDisabled: false,
		inputEnabled:   true,

		outputEnabled: false,
		output:        false,

		stateHistory: []simulation.ConnectionState{simulation.ConnectionStatePullDown, simulation.ConnectionStatePullDown},
		inputHistory: []bool{false, false},
	}
}

func (p *Pad) Component() simulation.Component {
	return p
}

func (p *Pad) ID() simulation.ComponentIdentifier { return p.id }
func (p *Pad) Label() string                      { return p.label }

func (p *Pad) Observer() Observer {
	return p
}

func (p *Pad) PulledUp() bool                             { return p.pullUp }
func (p *Pad) PulledDown() bool                           { return p.pullDown }
func (p *Pad) OutputDisabled() bool                       { return p.outputDisabled }
func (p *Pad) InputEnabled() bool                         { return p.inputEnabled }
func (p *Pad) OutputEnabled() bool                        { return p.outputEnabled }
func (p *Pad) Output() bool                               { return p.output }
func (p *Pad) StateHistory() []simulation.ConnectionState { return p.stateHistory }
func (p *Pad) InputHistory() []bool                       { return p.inputHistory }

func (p *Pad) Configurator() Configurator {
	return p
}

func (p *Pad) SetLabel(label string)                 { p.label = label }
func (p *Pad) SetPullUp(pullUp bool)                 { p.pullUp = pullUp }
func (p *Pad) SetPullDown(pullDown bool)             { p.pullDown = pullDown }
func (p *Pad) SetOutputDisabled(outputDisabled bool) { p.outputDisabled = outputDisabled }
func (p *Pad) SetInputEnabled(inputEnabled bool)     { p.inputEnabled = inputEnabled }

var ErrInvalidDelayCycles = errors.New("invalid delay cycles")

func updateHistoryLength[S ~[]E, E any](history S, cycles uint) (S, error) {
	if cycles == 0 {
		return history, fmt.Errorf("%w: %d, should be > 0", ErrInvalidDelayCycles, cycles)
	}
	currentLength := len(history)
	if int(cycles) > currentLength {
		newestValue := history[0]
		numberOfNewValues := int(cycles) - currentLength
		stretchedValues := slices.Repeat([]E{newestValue}, numberOfNewValues)
		history = append(stretchedValues, history...)
	} else if int(cycles) < len(history) {
		history = history[:cycles]
	}
	return history, nil
}

func (p *Pad) SetOutputDelayCycles(c uint) error {
	var err error
	p.stateHistory, err = updateHistoryLength(p.stateHistory, c)
	return err
}

func (p *Pad) SetInputDelayCycles(c uint) error {
	var err error
	p.inputHistory, err = updateHistoryLength(p.inputHistory, c)
	return err
}

func (p *Pad) Operator() Operator {
	return p
}

func (p *Pad) SetOutputEnabled(outputEnabled bool) { p.outputEnabled = outputEnabled }
func (p *Pad) SetOutput(output bool)               { p.output = output }
func (p *Pad) GetInput() bool                      { return p.inputHistory[len(p.inputHistory)-1] }

func (p *Pad) Tick() error {
	for i := len(p.stateHistory) - 1; i > 0; i-- {
		p.stateHistory[i] = p.stateHistory[i-1]
	}
	p.stateHistory[0] = padState(p.outputEnabled && !p.outputDisabled, p.output, p.pullUp, p.pullDown)

	for i := len(p.inputHistory) - 1; i > 0; i-- {
		p.inputHistory[i] = p.inputHistory[i-1]
	}
	// p.inputHistory[0] is updated by SetInput.

	return nil
}

func (p *Pad) Connection() simulation.Connection {
	return p
}

func (p *Pad) State() simulation.ConnectionState { return p.stateHistory[len(p.stateHistory)-1] }
func (p *Pad) SetInput(logicLevel bool)          { p.inputHistory[0] = logicLevel }
