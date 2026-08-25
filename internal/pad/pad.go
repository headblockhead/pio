package pad

import "github.com/headblockhead/pio/internal/connection"

func padState(shouldOutput bool, outputValue bool, pullUp bool, pullDown bool) connection.State {
	if shouldOutput {
		if outputValue {
			return connection.StateOutHigh
		} else {
			return connection.StateOutLow
		}
	} else {
		if pullUp && pullDown {
			return connection.StateBusKeeper
		}
		if pullUp {
			return connection.StatePullUp
		}
		if pullDown {
			return connection.StatePullDown
		}
		return connection.StateNone
	}
}

type Pad struct {
	id string

	pullUp         bool
	pullDown       bool
	outputDisabled bool
	inputEnabled   bool

	outputEnabled bool
	output        bool

	stateHistory []connection.State
	inputHistory []bool
}

func NewPad(id string) *Pad {
	return &Pad{
		id: id,

		pullDown:     true,
		inputEnabled: true,

		stateHistory: make([]connection.State, 2),
		inputHistory: make([]bool, 2),
	}
}

func (p *Pad) Tick() {
	for i := 1; i < len(p.stateHistory); i++ {
		p.stateHistory[i] = p.stateHistory[i-1]
	}
	p.stateHistory[0] = padState(p.outputEnabled && !p.outputDisabled, p.output, p.pullUp, p.pullDown)

	for i := 1; i < len(p.inputHistory); i++ {
		p.inputHistory[i] = p.inputHistory[i-1]
	}
	// p.inputHistory[0] is updated by SetInput.
}

type PadConfigurator interface {
	SetPullUp(bool)
	SetPullDown(bool)
	SetOutputDisabled(bool)
	SetInputEnabled(bool)
	SetOutputDelayCycles(uint)
	SetInputDelayCycles(uint)
}

func (p *Pad) Configurator() PadConfigurator {
	return p
}

func (p *Pad) SetPullUp(pu bool) {
	p.pullUp = pu
}
func (p *Pad) SetPullDown(pd bool) {
	p.pullDown = pd
}
func (p *Pad) SetOutputDisabled(d bool) {
	p.outputDisabled = d
}
func (p *Pad) SetInputEnabled(e bool) {
	p.inputEnabled = e
}
func (p *Pad) SetOutputDelayCycles(c uint) {
	prev := len(p.stateHistory)
	p.stateHistory = p.stateHistory[:c+1]
	// fill in any new space
	for i := prev; i < int(c+1); i++ {
		p.stateHistory[i] = p.stateHistory[prev-1]
	}
}
func (p *Pad) SetInputDelayCycles(c uint) {
	prev := len(p.inputHistory)
	p.inputHistory = p.inputHistory[:c+1]
	// fill in any new space
	for i := prev; i < int(c+1); i++ {
		p.inputHistory[i] = p.inputHistory[prev-1]
	}
}

type PadController interface {
	SetOutputEnabled(bool)
	SetOutput(bool)
	GetInput() bool
}

func (p *Pad) Controller() PadController {
	return p
}

func (p *Pad) SetOutputEnabled(e bool) {
	p.outputEnabled = e
}
func (p *Pad) SetOutput(o bool) {
	p.output = o
}
func (p *Pad) GetInput() bool {
	return p.inputHistory[len(p.inputHistory)-1]
}

func (p *Pad) Connection() connection.Connection {
	return p
}

func (p *Pad) ID() string {
	return p.id
}
func (p *Pad) GetState() connection.State {
	return p.stateHistory[len(p.stateHistory)-1]
}
func (p *Pad) SetInput(input bool) {
	p.inputHistory[0] = input
}
