package pio

type Pin struct {
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
