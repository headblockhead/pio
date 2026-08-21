package pio

type PadState uint

const (
	PadNone PadState = iota
	PadOutHigh
	PadOutLow
	PadBusKeeper
	PadPullUp
	PadPullDown
)

type pad struct {
	id string

	pullUp         bool
	pullDown       bool
	outputDisabled bool
	inputEnabled   bool

	outputEnabled bool
	output        bool

	inputSynchroniserEnabled bool

	state         PadState
	nextState     PadState
	nextNextState PadState

	input                 bool
	previousInput         bool
	previousPreviousInput bool
}

func newPad(id string) *pad {
	return &pad{
		id:           id,
		pullDown:     true,
		inputEnabled: true,
	}
}

func (p *pad) tick() {
	p.state = p.nextState
	p.nextState = p.nextNextState
	p.nextNextState = p.computeState()

	p.previousPreviousInput = p.previousInput
	p.previousInput = p.input
	// p.inputValue is updated by SetInput through the Connection interface.
}

func (p *pad) computeState() PadState {
	if !p.outputEnabled || p.outputDisabled {
		if p.pullUp && p.pullDown {
			return PadBusKeeper
		}
		if p.pullUp {
			return PadPullUp
		}
		if p.pullDown {
			return PadPullDown
		}
		return PadNone
	}
	if p.output {
		return PadOutHigh
	}
	return PadOutLow
}

func (p *pad) getInput() bool {
	if !p.inputEnabled {
		return false
	}
	if p.inputSynchroniserEnabled {
		return p.previousPreviousInput
	}
	return p.previousInput
}

// to implement Connection

func (p *pad) ID() string {
	return p.id
}

func (p *pad) GetState() PadState {
	if p.inputSynchroniserEnabled {
		return p.state
	}
	return p.nextState
}

func (p *pad) SetInput(input bool) {
	p.input = input
}
