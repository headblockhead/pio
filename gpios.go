package pio

import "errors"

type InputOverride uint
type OutputEnableOverride uint
type OutputOverride uint
type GPIOFunction uint

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

const gpioCount = 30

type GPIOs struct {
	pinMuxer map[GPIOFunction]PinsGPIOs

	outputEnableOverrides [gpioCount]OutputEnableOverride
	outputOverrides       [gpioCount]OutputOverride
	inputOverrides        [gpioCount]InputOverride
	functions             [gpioCount]GPIOFunction

	pads PadsGPIOs
}

func NewGPIOs(pio0Pins PinsGPIOs, pio1Pins PinsGPIOs, pads PadsGPIOs) *GPIOs {
	return &GPIOs{
		pinMuxer: map[GPIOFunction]PinsGPIOs{
			GPIOFunctionPIO0: pio0Pins,
			GPIOFunctionPIO1: pio1Pins,
		},
		pads: pads,
	}
}

var ErrGPIOOutOfBounds = errors.New("out of bounds")

func (g *GPIOs) SetOutputEnableOverride(gpio uint, outputEnableOverride OutputEnableOverride) error {
	if gpio >= gpioCount {
		return ErrGPIOOutOfBounds
	}
	g.outputEnableOverrides[gpio] = outputEnableOverride
	return nil
}

func (g *GPIOs) SetOutputOverride(gpio uint, outputOverride OutputOverride) error {
	if gpio >= gpioCount {
		return ErrGPIOOutOfBounds
	}
	g.outputOverrides[gpio] = outputOverride
	return nil
}

func (g *GPIOs) SetInputOverride(gpio uint, inputOverride InputOverride) error {
	if gpio >= gpioCount {
		return ErrGPIOOutOfBounds
	}
	g.inputOverrides[gpio] = inputOverride
	return nil
}

func (g *GPIOs) SetFunction(gpio uint, function GPIOFunction) error {
	if gpio >= gpioCount {
		return ErrGPIOOutOfBounds
	}
	g.functions[gpio] = function
	return nil
}

var ErrGPIOsUpdateInvalidFunction = errors.New("gpio is set to a function which does not exist in the muxer")

func (g *GPIOs) Update() error {
	var i uint
	for i = 0; i < gpioCount; i++ {
		pins, ok := g.pinMuxer[g.functions[i]]
		if !ok {
			return ErrGPIOsUpdateInvalidFunction
		}
		outputEnableFromPeripheral, err := pins.GetOutputEnable(i)
		if err != nil {
			return err
		}
		var outputEnableToPad PadOutputEnable
		switch g.outputEnableOverrides[i] {
		case OutputEnableOverrideNormal:
			switch outputEnableFromPeripheral {
			case PinOutputEnabled:
				outputEnableToPad = PadOutputEnabled
			case PinOutputNotEnabled:
				outputEnableToPad = PadOutputNotEnabled
			}
		case OutputEnableOverrideInvert:
			switch outputEnableFromPeripheral {
			case PinOutputEnabled:
				outputEnableToPad = PadOutputNotEnabled
			case PinOutputNotEnabled:
				outputEnableToPad = PadOutputEnabled
			}
		case OutputEnableOverrideDisable:
			outputEnableToPad = PadOutputNotEnabled
		case OutputEnableOverrideEnable:
			outputEnableToPad = PadOutputEnabled
		}
		err = g.pads.SetOutputEnable(i, outputEnableToPad)
		if err != nil {
			return err
		}
		outputFromPeripheral, err := pins.GetOutput(i)
		if err != nil {
			return err
		}
		var outputToPad PadOutput
		switch g.outputOverrides[i] {
		case OutputOverrideNormal:
			switch outputFromPeripheral {
			case PinOutputHigh:
				outputToPad = PadOutputHigh
			case PinOutputLow:
				outputToPad = PadOutputLow
			}
		case OutputOverrideInvert:
			switch outputFromPeripheral {
			case PinOutputHigh:
				outputToPad = PadOutputLow
			case PinOutputLow:
				outputToPad = PadOutputHigh
			}
		case OutputOverrideLow:
			outputToPad = PadOutputLow
		case OutputOverrideHigh:
			outputToPad = PadOutputHigh
		}
		err = g.pads.SetOutput(i, outputToPad)
		if err != nil {
			return err
		}
		inputFromPad, err := g.pads.GetInput(i)
		if err != nil {
			return err
		}
		var inputToPeripheral PinInput
		switch g.inputOverrides[i] {
		case InputOverrideNormal:
			switch inputFromPad {
			case PadInputHigh:
				inputToPeripheral = PinInputHigh
			case PadInputLow:
				inputToPeripheral = PinInputLow
			}
		case InputOverrideInvert:
			switch inputFromPad {
			case PadInputHigh:
				inputToPeripheral = PinInputLow
			case PadInputLow:
				inputToPeripheral = PinInputHigh
			}
		case InputOverrideLow:
			inputToPeripheral = PinInputLow
		case InputOverrideHigh:
			inputToPeripheral = PinInputHigh
		}
		err = pins.SetInput(i, inputToPeripheral)
		if err != nil {
			return err
		}
	}
	return nil
}
