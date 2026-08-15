package pio

type PadGPIO interface {
	SetOutputEnable(bool)
	GetOutputEnable() bool
	SetOutput(bool)
	GetOutput() bool
	GetInput() bool
}

type PadControl interface {
	SetPullUpEnable(bool)
	GetPullUpEnable() bool
	SetPullDownEnable(bool)
	GetPullDownEnable() bool
}

type Pad struct {
	spm SolvablePadManipulator

	outputEnable   bool
	outputValue    bool
	pullUpEnable   bool
	pullDownEnable bool
}

func NewPad(spm SolvablePadManipulator) *Pad {
	return &Pad{
		spm: spm,

		outputEnable: false,
		outputValue:  false,
	}
}

func (p *Pad) SetOutputEnable(outputEnable bool) {
	p.outputEnable = outputEnable
}
func (p *Pad) GetOutputEnable() bool {
	return p.outputEnable
}
func (p *Pad) SetOutput(output bool) {
	p.outputValue = output
}
func (p *Pad) GetOutput() bool {
	return p.outputValue
}
func (p *Pad) GetInput() bool {
	switch p.spm.GetInputState() {
	case PadInputStateLow:
		return false
	case PadInputStateHigh:
		return true
	case PadInputStateFloating:
		return false
	}
	return false
}

func (p *Pad) SetPullUpEnable(pullUpEnable bool) {
	p.pullUpEnable = pullUpEnable
}
func (p *Pad) GetPullUpEnable() bool {
	return p.pullUpEnable
}
func (p *Pad) SetPullDownEnable(pullDownEnable bool) {
	p.pullDownEnable = pullDownEnable
}
func (p *Pad) GetPullDownEnable() bool {
	return p.pullDownEnable
}

func (p *Pad) Update() {
	if p.GetOutputEnable() == true {
		if p.GetOutput() == true {
			p.spm.SetOutputState(PadOutputStateDrivenHigh)
		} else {
			p.spm.SetOutputState(PadOutputStateDrivenLow)
		}
	} else {
		pulledUp := p.GetPullUpEnable()
		pulledDown := p.GetPullDownEnable()
		if pulledUp && pulledDown {
			if p.GetInput() == true {
				p.spm.SetOutputState(PadOutputStatePulledUp)
			} else {
				p.spm.SetOutputState(PadOutputStatePulledDown)
			}
		} else if pulledUp {
			p.spm.SetOutputState(PadOutputStatePulledUp)
		} else if pulledDown {
			p.spm.SetOutputState(PadOutputStatePulledDown)
		} else {
			p.spm.SetOutputState(PadOutputStateNone)
		}
	}
}
