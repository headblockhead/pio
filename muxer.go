package pio

type PinPeripheralIO interface {
	SetOutput(bool)
	SetOutputEnable(bool)
	GetInput() bool
}

type PinGPIO interface {
	GetOutput() bool
	GetOutputEnable() bool
	SetInput(bool)
}

type Pin struct {
	output       bool
	outputEnable bool
	input        bool
}

func (p *Pin) SetOutput(output bool) {
	p.output = output
}
func (p *Pin) GetOutput() bool {
	return p.output
}
func (p *Pin) SetOutputEnable(outputEnable bool) {
	p.outputEnable = outputEnable
}
func (p *Pin) GetOutputEnable() bool {
	return p.outputEnable
}
func (p *Pin) GetInput() bool {
	return p.input
}
func (p *Pin) SetInput(input bool) {
	p.input = input
}

type MuxerPinPeripheralIO interface {
	SetOutput(GPIOFunction, bool)
	SetOutputEnable(GPIOFunction, bool)
	GetInput(GPIOFunction) bool
}

type MuxerPinGPIO interface {
	GetOutput(GPIOFunction) bool
	GetOutputEnable(GPIOFunction) bool
	SetInput(GPIOFunction, bool)
}

type MuxerPin struct {
	muxer map[GPIOFunction]*Pin
}

func NewMuxerPin(pio0Pin *Pin, pio1Pin *Pin) *MuxerPin {
	return &MuxerPin{
		muxer: map[GPIOFunction]*Pin{
			GPIOFunctionNull: &Pin{},
			GPIOFunctionPIO0: pio0Pin,
			GPIOFunctionPIO1: pio1Pin,
		},
	}
}

func (mp *MuxerPin) SetOutput(function GPIOFunction, output bool) {
	mp.muxer[function].SetOutput(output)
}
func (mp *MuxerPin) GetOutput(function GPIOFunction) bool {
	return mp.muxer[function].GetOutput()
}
func (mp *MuxerPin) SetOutputEnable(function GPIOFunction, outputEnable bool) {
	mp.muxer[function].SetOutputEnable(outputEnable)
}
func (mp *MuxerPin) GetOutputEnable(function GPIOFunction) bool {
	return mp.muxer[function].GetOutputEnable()
}
func (mp *MuxerPin) SetInput(function GPIOFunction, input bool) {
	mp.muxer[function].SetInput(input)
}
func (mp *MuxerPin) GetInput(function GPIOFunction) bool {
	return mp.muxer[function].GetInput()
}
