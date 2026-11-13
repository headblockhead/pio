package memory

import (
	"fmt"

	"git.edwardh.dev/pio/instr"
)

type Program struct {
	instructions []uint16
	hasOrigin    bool
	origin       uint
}

func NewProgram(instructions []uint16) *Program {
	return &Program{
		instructions: instructions,
		hasOrigin:    false,
		origin:       0,
	}
}

func NewProgramWithOrigin(instructions []uint16, origin uint) *Program {
	return &Program{
		instructions: instructions,
		hasOrigin:    true,
		origin:       origin,
	}
}

func (p *Program) Instructions() []uint16 {
	return p.instructions
}

func (p *Program) HasOrigin() bool {
	return p.hasOrigin
}

func (p *Program) Origin() uint {
	return p.origin
}

func (p *Program) Strings(pinCountSideset uint, sidesetIsOptional bool) ([]string, error) {
	var s []string

	for i, instruction := range p.instructions {
		ins, err := instr.String(instruction, pinCountSideset, sidesetIsOptional)
		if err != nil {
			return nil, fmt.Errorf("failed to decode instruction: %w", err)
		}
		s = append(s, fmt.Sprintf("% 2d: %s", i, ins))
	}

	return s, nil
}
