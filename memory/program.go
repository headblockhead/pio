package memory

type Program struct {
	instructions []uint16
	origin       *uint
}

func NewProgram(instructions []uint16) *Program {
	return &Program{
		instructions: instructions,
		origin:       nil,
	}
}

func NewProgramWithOrigin(instructions []uint16, origin uint) *Program {
	return &Program{
		instructions: instructions,
		origin:       &origin,
	}
}

func (p *Program) Instructions() []uint16 {
	return p.instructions
}

func (p *Program) Origin() *uint {
	return p.origin
}
