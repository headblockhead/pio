package pio

type PIO struct {
	stateMachines     [4]SM
	pins              [32]Pin
	irqs              *IRQs
	instructionMemory *InstructionMemory
}
