package pioblock

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/headblockhead/pio/memory"
	"github.com/headblockhead/pio/simulation"
	"github.com/headblockhead/pio/statemachine"
)

type Observer interface {
	MemoryObserver() memory.Observer
	StateMachineObserver(uint) (statemachine.Observer, error)

	IRQs() uint8
	PinInputs() uint32

	PinOutputEnables() uint32
	PinOutputs() uint32
}

type Configurator interface {
	SetLabel(string)
	MemoryWriter() memory.Writer
	StateMachineConfigurator(uint) (statemachine.Configurator, error)

	SetPinOutput(uint, bool) error
	SetPinOutputEnable(uint, bool) error
}

type Operator interface {
	simulation.Ticker

	SetPinInputs(uint32)

	PinOutputEnables() uint32
	PinOutputs() uint32
	IRQs() uint8
}

type PIOBlock struct {
	id    simulation.ComponentIdentifier
	label string

	memory        *memory.Memory
	stateMachines []*statemachine.StateMachine

	irqs uint8

	pinInputs uint32

	pinOutputEnables uint32
	pinOutputs       uint32
}

func New(label string, memorySize uint, numberOfStateMachines uint) *PIOBlock {
	p := &PIOBlock{
		id:            simulation.NewComponentIdentifier(),
		label:         label,
		memory:        memory.New(memorySize),
		stateMachines: make([]*statemachine.StateMachine, numberOfStateMachines),

		irqs: 0,

		pinInputs: 0,

		pinOutputEnables: 0,
		pinOutputs:       0,
	}

	for i := range numberOfStateMachines {
		p.stateMachines[i] = statemachine.New(i, strconv.Itoa(int(i)), p.memory.Reader())
	}

	return p
}

func (p *PIOBlock) StateMachine(i uint) (*statemachine.StateMachine, error) {
	if i >= uint(len(p.stateMachines)) {
		return nil, fmt.Errorf("%d: %w", i, ErrOutOfRange)
	}
	return p.stateMachines[i], nil
}

func (p *PIOBlock) Observer() Observer {
	return p
}

func (p *PIOBlock) MemoryObserver() memory.Observer { return p.memory.Observer() }

var ErrOutOfRange = errors.New("out of range")

func (p *PIOBlock) StateMachineObserver(i uint) (statemachine.Observer, error) {
	return p.StateMachine(i)
}

func (p *PIOBlock) IRQs() uint8              { return p.irqs }
func (p *PIOBlock) PinInputs() uint32        { return p.pinInputs }
func (p *PIOBlock) PinOutputEnables() uint32 { return p.pinOutputEnables }
func (p *PIOBlock) PinOutputs() uint32       { return p.pinOutputs }

func (p *PIOBlock) Configurator() Configurator {
	return p
}

func (p *PIOBlock) SetLabel(label string)       { p.label = label }
func (p *PIOBlock) MemoryWriter() memory.Writer { return p.memory.Writer() }
func (p *PIOBlock) StateMachineConfigurator(i uint) (statemachine.Configurator, error) {
	return p.StateMachine(i)
}

func (p *PIOBlock) SetPinOutput(i uint, v bool) error {
	if i > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrOutOfRange, i)
	}
	if v {
		p.pinOutputs |= (0b1 << i)
	} else {
		p.pinOutputs &= ^(0b1 << i)
	}
	return nil
}

func (p *PIOBlock) SetPinOutputEnable(i uint, v bool) error {
	if i > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrOutOfRange, i)
	}
	if v {
		p.pinOutputEnables |= (0b1 << i)
	} else {
		p.pinOutputEnables &= ^(0b1 << i)
	}
	return nil
}

func (p *PIOBlock) Operator() Operator {
	return p
}

func (p *PIOBlock) ID() simulation.ComponentIdentifier { return p.id }
func (p *PIOBlock) Label() string                      { return p.label }
func (p *PIOBlock) SetPinInputs(pinInputs uint32)      { p.pinInputs = pinInputs }

func (p *PIOBlock) Tick() error {
	inputIRQs := p.irqs
	inputsPins := p.pinInputs

	for _, sm := range p.stateMachines {
		o := sm.Operator()
		o.SetIRQInputs(inputIRQs)
		o.SetPinInputs(inputsPins)

		err := o.Tick()
		if err != nil {
			return fmt.Errorf("state machine [%v]: %w", sm.Label(), err)
		}

		p.pinOutputEnables &= ^o.PinOutputEnablesMask()
		p.pinOutputEnables |= (o.PinOutputEnables() & o.PinOutputEnablesMask())
		p.pinOutputs &= ^o.PinOutputsMask()
		p.pinOutputs |= (o.PinOutputs() & o.PinOutputsMask())
		if o.SidesetControlsPinDirection() {
			p.pinOutputEnables &= ^o.PinSidesetsMask()
			p.pinOutputEnables |= (o.PinSidesets() & o.PinSidesetsMask())
		} else {
			p.pinOutputs &= ^o.PinSidesetsMask()
			p.pinOutputs |= (o.PinSidesets() & o.PinSidesetsMask())
		}
		p.irqs &= ^o.IRQWritesMask()
		p.irqs |= (o.IRQWrites() & o.IRQWritesMask())
	}

	return nil
}
