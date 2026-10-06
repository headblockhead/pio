package pio

import (
	"errors"
	"fmt"

	"github.com/headblockhead/pio/internal/memory"
	"github.com/headblockhead/pio/internal/sm"
)

type Observer interface {
	MemoryObserver() memory.Observer
	SMObserver(uint) (sm.Observer, error)

	IRQs() uint8
	PinInputs() uint32

	PinOutputEnables() uint32
	PinOutputs() uint32
}

type Configurator interface {
	MemoryWriter() memory.Writer
	SMConfigurator(uint) (sm.Configurator, error)
}

type Controller interface {
	SetPinInputs(uint32)

	Tick() error

	PinOutputEnables() uint32
	PinOutputs() uint32
	IRQs() uint8
}

type PIO struct {
	memory        *memory.Memory
	stateMachines []*sm.SM

	irqs uint8

	pinInputs uint32

	pinOutputEnables uint32
	pinOutputs       uint32
}

func NewPIO(memorySize uint, numberOfSMs uint) *PIO {
	p := &PIO{}
	p.memory = memory.NewMemory(memorySize)
	p.stateMachines = make([]*sm.SM, numberOfSMs)
	for i := range numberOfSMs {
		p.stateMachines[i] = sm.NewSM(i, p.memory.Reader())
	}
	return p
}

func (p *PIO) Observer() Observer {
	return p
}

func (p *PIO) MemoryObserver() memory.Observer { return p.memory.Observer() }

var ErrSMOutOfRange = errors.New("SM out of range")

func (p *PIO) SMObserver(i uint) (sm.Observer, error) {
	if i >= uint(len(p.stateMachines)) {
		return nil, ErrSMOutOfRange
	}
	return p.stateMachines[i].Observer(), nil
}

func (p *PIO) IRQs() uint8              { return p.irqs }
func (p *PIO) PinInputs() uint32        { return p.pinInputs }
func (p *PIO) PinOutputEnables() uint32 { return p.pinOutputEnables }
func (p *PIO) PinOutputs() uint32       { return p.pinOutputs }

func (p *PIO) Configurator() Configurator {
	return p
}

func (p *PIO) MemoryWriter() memory.Writer { return p.memory.Writer() }

func (p *PIO) SMConfigurator(i uint) (sm.Configurator, error) {
	if i >= uint(len(p.stateMachines)) {
		return nil, ErrSMOutOfRange
	}
	return p.stateMachines[i].Configurator(), nil
}

func (p *PIO) Controller() Controller {
	return p
}

func (p *PIO) SetPinInputs(pinInputs uint32) { p.pinInputs = pinInputs }

func (p *PIO) Tick() error {
	inputIRQs := p.irqs
	inputsPins := p.pinInputs

	for i, sm := range p.stateMachines {
		c := sm.Controller()
		c.SetIRQInputs(inputIRQs)
		c.SetPinInputs(inputsPins)

		err := c.Tick()
		if err != nil {
			return fmt.Errorf("SM %d: %w", i, err)
		}

		p.pinOutputEnables &= ^c.PinOutputEnablesMask()
		p.pinOutputEnables |= (c.PinOutputEnables() & c.PinOutputEnablesMask())
		p.pinOutputs &= ^c.PinOutputsMask()
		p.pinOutputs |= (c.PinOutputs() & c.PinOutputsMask())
		if c.SidesetControlsPinDirection() {
			p.pinOutputEnables &= ^c.PinSidesetsMask()
			p.pinOutputEnables |= (c.PinSidesets() & c.PinSidesetsMask())
		} else {
			p.pinOutputs &= ^c.PinSidesetsMask()
			p.pinOutputs |= (c.PinSidesets() & c.PinSidesetsMask())
		}
		p.irqs &= ^c.IRQWritesMask()
		p.irqs |= (c.IRQWrites() & c.IRQWritesMask())
	}

	return nil
}
