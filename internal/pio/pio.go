package pio

import (
	"github.com/headblockhead/pio/internal/memory"
	"github.com/headblockhead/pio/internal/sm"
)

type PIO struct {
	memory        *memory.Memory
	stateMachines []*sm.SM

	irqs uint8

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
