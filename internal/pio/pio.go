package pio

import "github.com/headblockhead/pio/internal/memory"
import "github.com/headblockhead/pio/internal/sm"

type PIO struct {
	memoryReader  *memory.Memory
	stateMachines []*sm.SM

	irqs uint8

	pinOutputs           uint32
	pinOutputMask        uint32
	pinOutputEnables     uint32
	pinOutputEnablesMask uint32
	pinSidesets          uint32
	pinSidesetsMask      uint32
}
