package pio

type InputOverride int
type OutputEnableOverride int
type OutputOverride int
type GPIOFunction int

const (
	InputOverrideNormal InputOverride = iota
	InputOverrideInvert
	InputOverrideLow
	InputOverrideHigh

	OutputEnableOverrideNormal OutputEnableOverride = iota
	OutputEnableOverrideInvert
	OutputEnableOverrideDisable
	OutputEnableOverrideEnable

	OutputOverrideNormal OutputOverride = iota
	OutputOverrideInvert
	OutputOverrideLow
	OutputOverrideHigh

	GPIOFunctionNull GPIOFunction = iota
	GPIOFunctionPIO0
	GPIOFunctionPIO1
)

type GPIOStatus interface {
	InputToPeripheral() bool
	InputFromPad() bool
	OutputEnableToPad() bool
	OutputEnableFromPeripheral() bool
	OutputToPad() bool
	OutputFromPeripheral() bool
}

type GPIOControl interface {
	SetInputOverride(InputOverride)
	GetInputOverride() InputOverride
	SetOutputEnableOverride(OutputEnableOverride)
	GetOutputEnableOverride() OutputEnableOverride
	SetOutputOverride(OutputOverride)
	GetOutputOverride() OutputOverride
	SetFunction(GPIOFunction)
	GetFunction() GPIOFunction
}

type GPIO struct {
	padGPIO  PadGPIO
	muxerPin MuxerPinGPIO

	inputOverride        InputOverride
	outputEnableOverride OutputEnableOverride
	outputOverride       OutputOverride
	function             GPIOFunction
}

func NewGPIO(padGPIO PadGPIO, muxerPinGPIO MuxerPinGPIO) *GPIO {
	return &GPIO{
		padGPIO:  padGPIO,
		muxerPin: muxerPinGPIO,

		inputOverride:        InputOverrideNormal,
		outputEnableOverride: OutputEnableOverrideNormal,
		outputOverride:       OutputOverrideNormal,
		function:             GPIOFunctionNull,
	}
}

func (g *GPIO) InputToPeripheral() bool {
	switch g.inputOverride {
	case InputOverrideNormal:
		return g.InputFromPad()
	case InputOverrideInvert:
		return !g.InputFromPad()
	case InputOverrideLow:
		return false
	case InputOverrideHigh:
		return true
	}
	return false
}
func (g *GPIO) InputFromPad() bool {
	return g.padGPIO.GetInput()
}
func (g *GPIO) OutputEnableToPad() bool {
	switch g.outputEnableOverride {
	case OutputEnableOverrideNormal:
		return g.OutputEnableFromPeripheral()
	case OutputEnableOverrideInvert:
		return !g.OutputEnableFromPeripheral()
	case OutputEnableOverrideDisable:
		return false
	case OutputEnableOverrideEnable:
		return true
	}
	return false
}
func (g *GPIO) OutputEnableFromPeripheral() bool {
	return g.muxerPin.GetOutputEnable(g.function)
}
func (g *GPIO) OutputToPad() bool {
	switch g.outputOverride {
	case OutputOverrideNormal:
		return g.OutputFromPeripheral()
	case OutputOverrideInvert:
		return !g.OutputFromPeripheral()
	case OutputOverrideLow:
		return false
	case OutputOverrideHigh:
		return true
	}
	return false
}
func (g *GPIO) OutputFromPeripheral() bool {
	return g.muxerPin.GetOutput(g.function)
}

func (g *GPIO) SetInputOverride(inputOverride InputOverride) {
	g.inputOverride = inputOverride
}
func (g *GPIO) GetInputOverride() InputOverride {
	return g.inputOverride
}
func (g *GPIO) SetOutputEnableOverride(outputEnableOverride OutputEnableOverride) {
	g.outputEnableOverride = outputEnableOverride
}
func (g *GPIO) GetOutputEnableOverride() OutputEnableOverride {
	return g.outputEnableOverride
}
func (g *GPIO) SetOutputOverride(outputOverride OutputOverride) {
	g.outputOverride = outputOverride
}
func (g *GPIO) GetOutputOverride() OutputOverride {
	return g.outputOverride
}
func (g *GPIO) SetFunction(function GPIOFunction) {
	g.function = function
}
func (g *GPIO) GetFunction() GPIOFunction {
	return g.function
}

func (g *GPIO) Update() {
	g.padGPIO.SetOutputEnable(g.OutputEnableToPad())
	g.padGPIO.SetOutput(g.OutputToPad())
	g.muxerPin.SetInput(g.function, g.InputToPeripheral())
}
