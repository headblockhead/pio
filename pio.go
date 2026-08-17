package pio

type PIO struct {
	stateMachines     [4]SM
	irqs              *IRQs
	instructionMemory *InstructionMemory
}
