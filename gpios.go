package pio

import "errors"
import "fmt"

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
	muxer map[GPIOFunction]PinsGPIOs

	outputEnableOverrides [gpioCount]OutputEnableOverride
	outputOverrides       [gpioCount]OutputOverride
	inputOverrides        [gpioCount]InputOverride
	functions             [gpioCount]GPIOFunction

	pads PadsGPIOs
}

func NewGPIOs(pio0Pins PinsGPIOs, pio1Pins PinsGPIOs, pads PadsGPIOs) *GPIOs {
	return &GPIOs{
		muxer: map[GPIOFunction]PinsGPIOs{
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

var ErrGPIOsUpdateInvalidOverride = errors.New("invalid override")
var ErrGPIOsUpdateInvalidPinState = errors.New("invalid pin state")

func (g *GPIOs) UpdateOutputEnable(pin uint, pins PinsGPIOs) error {
	outputEnableFromPeripheral, err := pins.GetOutputEnable(pin)
	if err != nil {
		return fmt.Errorf("error getting pin %d: %w", pin, err)
	}
	var outputEnableToPad PadOutputEnable
	switch g.outputEnableOverrides[pin] {
	case OutputEnableOverrideNormal:
		switch outputEnableFromPeripheral {
		case PinOutputEnabled:
			outputEnableToPad = PadOutputEnabled
		case PinOutputNotEnabled:
			outputEnableToPad = PadOutputNotEnabled
		default:
			return ErrGPIOsUpdateInvalidPinState
		}
	case OutputEnableOverrideInvert:
		switch outputEnableFromPeripheral {
		case PinOutputEnabled:
			outputEnableToPad = PadOutputNotEnabled
		case PinOutputNotEnabled:
			outputEnableToPad = PadOutputEnabled
		default:
			return ErrGPIOsUpdateInvalidPinState
		}
	case OutputEnableOverrideDisable:
		outputEnableToPad = PadOutputNotEnabled
	case OutputEnableOverrideEnable:
		outputEnableToPad = PadOutputEnabled
	default:
		return ErrGPIOsUpdateInvalidOverride
	}
	err = g.pads.SetOutputEnable(pin, outputEnableToPad)
	if err != nil {
		return fmt.Errorf("error setting pad %d: %w", pin, err)
	}
	return nil
}

func (g *GPIOs) UpdateOutput(pin uint, pins PinsGPIOs) error {
	outputFromPeripheral, err := pins.GetOutput(pin)
	if err != nil {
		return fmt.Errorf("error getting pin %d: %w", pin, err)
	}
	var outputToPad PadOutput
	switch g.outputOverrides[pin] {
	case OutputOverrideNormal:
		switch outputFromPeripheral {
		case PinOutputHigh:
			outputToPad = PadOutputHigh
		case PinOutputLow:
			outputToPad = PadOutputLow
		default:
			return ErrGPIOsUpdateInvalidPinState
		}
	case OutputOverrideInvert:
		switch outputFromPeripheral {
		case PinOutputHigh:
			outputToPad = PadOutputLow
		case PinOutputLow:
			outputToPad = PadOutputHigh
		default:
			return ErrGPIOsUpdateInvalidPinState
		}
	case OutputOverrideLow:
		outputToPad = PadOutputLow
	case OutputOverrideHigh:
		outputToPad = PadOutputHigh
	default:
		return ErrGPIOsUpdateInvalidOverride
	}
	err = g.pads.SetOutput(pin, outputToPad)
	if err != nil {
		return fmt.Errorf("error setting pad %d: %w", pin, err)
	}
	return nil
}

var ErrGPIOsUpdateInvalidPadState = errors.New("invalid pad state")

func (g *GPIOs) UpdateInput(pin uint, pins PinsGPIOs) error {
	inputFromPad, err := g.pads.GetInput(pin)
	if err != nil {
		return fmt.Errorf("error getting pad %d: %w", pin, err)
	}
	var inputToPeripheral PinInput
	switch g.inputOverrides[pin] {
	case InputOverrideNormal:
		switch inputFromPad {
		case PadInputHigh:
			inputToPeripheral = PinInputHigh
		case PadInputLow:
			inputToPeripheral = PinInputLow
		default:
			return ErrGPIOsUpdateInvalidPadState
		}
	case InputOverrideInvert:
		switch inputFromPad {
		case PadInputHigh:
			inputToPeripheral = PinInputLow
		case PadInputLow:
			inputToPeripheral = PinInputHigh
		default:
			return ErrGPIOsUpdateInvalidPadState
		}
	case InputOverrideLow:
		inputToPeripheral = PinInputLow
	case InputOverrideHigh:
		inputToPeripheral = PinInputHigh
	default:
		return ErrGPIOsUpdateInvalidOverride
	}
	err = pins.SetInput(pin, inputToPeripheral)
	if err != nil {
		return fmt.Errorf("error setting pin %d: %w", pin, err)
	}
	return nil
}

var ErrGPIOsUpdateInvalidFunction = errors.New("gpio is set to a function which does not exist in the muxer")

func (g *GPIOs) Update() error {
	var i uint
	for i = 0; i < gpioCount; i++ {
		pins, ok := g.muxer[g.functions[i]]
		if !ok {
			return ErrGPIOsUpdateInvalidFunction
		}
		err := g.UpdateOutputEnable(i, pins)
		if err != nil {
			return fmt.Errorf("error updating outputEnable: %w", err)
		}
		err = g.UpdateOutput(i, pins)
		if err != nil {
			return fmt.Errorf("error updating output: %w", err)
		}
		err = g.UpdateInput(i, pins)
		if err != nil {
			return fmt.Errorf("error updating input: %w", err)
		}
	}
	return nil
}
