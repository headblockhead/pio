package pio

type Pin interface {
	GetOutput() (bool, error)
	GetOutputEnable() (bool, error)
	SetInput(input bool) error
}

type PinManipulator interface {
	SetOutput(output bool) error
	SetOutputEnable(outputEnable bool) error
	GetInput() (bool, error)
}

type Pin struct {
	output       bool
	outputEnable bool
	input        bool
}

func NewPin() *Pin {
	return &Pin{}
}

func (p *Pin) SetOutput(output bool) error {
	p.output = output
	return nil
}
func (p *Pin) GetOutput() (bool, error) {
	return p.output, nil
}
func (p *Pin) SetOutputEnable(outputEnable bool) error {
	p.outputEnable = outputEnable
	return nil
}
func (p *Pin) GetOutputEnable() (bool, error) {
	return p.outputEnable, nil
}
func (p *Pin) SetInput(input bool) error {
	p.input = input
	return nil
}
func (p *Pin) GetInput() (bool, error) {
	return p.input, nil
}
