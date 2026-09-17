package pio

import (
	"fmt"

	"github.com/headblockhead/pio/internal/memory"
	"github.com/headblockhead/pio/internal/sm"
)

type PIO struct {
	memory        *memory.Memory
	stateMachines []*sm.SM

	irqs uint8

	pinInputs uint32

	pinOutputEnables     uint32
	pinOutputEnablesMask uint32
	pinOutputs           uint32
	pinOutputMask        uint32
	pinSidesets          uint32
	pinSidesetsMask      uint32
}

func NewPIO(memorySize uint, numberOfSMs uint) *PIO {
	p := &PIO{}
	p.memory = memory.NewMemory(memorySize)
	for i := range numberOfSMs {
		p.stateMachines[i] = sm.NewSM(i, p.memory.Reader())
	}
	return p
}

func (p *PIO) Tick() error {
	for i, sm := range p.stateMachines {
		err := sm.Tick()
		if err != nil {
			return fmt.Errorf("error ticking state machine %d: %w", i, err)
		}
	}

	return nil
}
