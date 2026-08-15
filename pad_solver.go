package pio

import "errors"

type PadInputState int
type PadOutputState int

const (
	PadInputStateFloating PadInputState = iota
	PadInputStateLow
	PadInputStateHigh

	PadOutputStateNone PadOutputState = iota
	PadOutputStateDrivenHigh
	PadOutputStateDrivenLow
	PadOutputStatePulledUp
	PadOutputStatePulledDown
)

type SolvablePadManipulator interface {
	GetInputState() PadInputState
	SetOutputState(PadOutputState)
}

type SolvablePad struct {
	inputState  PadInputState
	outputState PadOutputState
}

func (sp *SolvablePad) SetInputState(inputState PadInputState) {
	sp.inputState = inputState
}
func (sp *SolvablePad) GetInputState() PadInputState {
	return sp.inputState
}
func (sp *SolvablePad) SetOutputState(outputState PadOutputState) {
	sp.outputState = outputState
}
func (sp *SolvablePad) GetOutputState() PadOutputState {
	return sp.outputState
}

type PadSolver struct {
	connectedPads []SolvablePad
}

var PadSolverErrorConflictingDrive = errors.New("pin is driven both low and high, which produced a short circuit")

func (ps *PadSolver) Update() error {
	solvedValue := PadInputStateFloating
	drivenHigh := false
	drivenLow := false
	pulledUp := false
	pulledDown := false
	for _, p := range ps.connectedPads {
		switch p.GetOutputState() {
		case PadOutputStateDrivenHigh:
			if drivenLow {
				return PadSolverErrorConflictingDrive
			}
			drivenHigh = true
		case PadOutputStateDrivenLow:
			if drivenHigh {
				return PadSolverErrorConflictingDrive
			}
			drivenLow = true
		case PadOutputStatePulledUp:
			pulledUp = true
		case PadOutputStatePulledDown:
			pulledDown = true
		}
	}
	if drivenHigh {
		solvedValue = PadInputStateHigh
	} else if drivenLow {
		solvedValue = PadInputStateLow
	} else if pulledUp && !pulledDown {
		solvedValue = PadInputStateHigh
	} else if pulledDown && !pulledUp {
		solvedValue = PadInputStateLow
	} else {
		solvedValue = PadInputStateFloating
	}
	for _, p := range ps.connectedPads {
		p.SetInputState(solvedValue)
	}
	return nil
}
