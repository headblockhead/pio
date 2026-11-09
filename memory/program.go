package memory

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
