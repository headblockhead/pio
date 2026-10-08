package pio

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/headblockhead/pio/gpio"
	"github.com/headblockhead/pio/pad"
	"github.com/headblockhead/pio/pioblock"
	"github.com/headblockhead/pio/simulation"
)

type RP2040Observer interface {
	PIOObserver(uint) (pioblock.Observer, error)
	GPIOObserver(uint) (gpio.Observer, error)
	PadObserver(uint) (pad.Observer, error)
}
type RP2040Configurator interface {
	SetLabel(string)
	PIOConfigurator(uint) (pioblock.Configurator, error)
	GPIOConfigurator(uint) (gpio.Configurator, error)
	PadConfigurator(uint) (pad.Configurator, error)
	Connection(uint) (simulation.Connection, error)
}
type RP2040Operator interface {
	simulation.Ticker
}

const rp2040PIOCount = 2
const rp2040MemorySize = 32
const rp2040StateMachineCount = 4
const rp2040PadCount = 30

type RP2040 struct {
	id    simulation.ComponentIdentifier
	label string

	pioBlocks [rp2040PIOCount]*pioblock.PIOBlock
	gpios     [rp2040PadCount]*gpio.GPIO
	pads      [rp2040PadCount]*pad.Pad
}

func NewRP2040(label string) *RP2040 {
	r := &RP2040{
		id:        simulation.NewComponentIdentifier(),
		label:     label,
		pioBlocks: [rp2040PIOCount]*pioblock.PIOBlock{},
		gpios:     [rp2040PadCount]*gpio.GPIO{},
		pads:      [rp2040PadCount]*pad.Pad{},
	}
	for i := range 2 {
		r.pioBlocks[i] = pioblock.New(strconv.Itoa(int(i)), rp2040MemorySize, rp2040StateMachineCount)
	}
	for i := range 30 {
		r.gpios[i] = gpio.New()
		r.pads[i] = pad.New(strconv.Itoa(int(i)))
	}
	return r
}

var ErrOutOfRange = errors.New("out of range")

func (r *RP2040) PIO(i uint) (*pioblock.PIOBlock, error) {
	if i >= rp2040PIOCount {
		return nil, fmt.Errorf("%w: %d, should be < %d", ErrOutOfRange, i, rp2040PIOCount)
	}
	return r.pioBlocks[i], nil
}

func (r *RP2040) GPIO(i uint) (*gpio.GPIO, error) {
	if i >= rp2040PadCount {
		return nil, fmt.Errorf("%w: %d, should be < %d", ErrOutOfRange, i, rp2040PadCount)
	}
	return r.gpios[i], nil
}

func (r *RP2040) Pad(i uint) (*pad.Pad, error) {
	if i >= rp2040PadCount {
		return nil, fmt.Errorf("%w: %d, should be < %d", ErrOutOfRange, i, rp2040PadCount)
	}
	return r.pads[i], nil
}

func (r *RP2040) Observer() RP2040Observer {
	return r
}

func (r *RP2040) PIOObserver(i uint) (pioblock.Observer, error) {
	return r.PIO(i)
}
func (r *RP2040) GPIOObserver(i uint) (gpio.Observer, error) {
	return r.GPIO(i)
}
func (r *RP2040) PadObserver(i uint) (pad.Observer, error) {
	return r.Pad(i)
}

func (r *RP2040) Configurator() RP2040Configurator {
	return r
}

func (r *RP2040) SetLabel(label string) { r.label = label }

func (r *RP2040) PIOConfigurator(i uint) (pioblock.Configurator, error) {
	return r.PIO(i)
}
func (r *RP2040) GPIOConfigurator(i uint) (gpio.Configurator, error) {
	return r.GPIO(i)
}
func (r *RP2040) PadConfigurator(i uint) (pad.Configurator, error) {
	return r.Pad(i)
}
func (r *RP2040) Connection(i uint) (simulation.Connection, error) {
	return r.Pad(i)
}

func (r *RP2040) Operator() RP2040Operator {
	return r
}

func (r *RP2040) ID() simulation.ComponentIdentifier { return r.id }
func (r *RP2040) Label() string                      { return r.label }

var ErrGPIOFunctionInvalid = errors.New("gpio function invalid")

func (r *RP2040) Tick() error {
	var pinInputs uint32

	for i, p := range r.pads {
		padOperator := p.Operator()

		inputHigh := padOperator.GetInput()
		if inputHigh {
			pinInputs |= (0b1 << i)
		}

		err := padOperator.Tick()
		if err != nil {
			return fmt.Errorf("pad [%v]: %w", padOperator.Label(), err)
		}
	}

	for _, p := range r.pioBlocks {
		pioBlockOperator := p.Operator()

		pioBlockOperator.SetPinInputs(pinInputs)

		err := pioBlockOperator.Tick()
		if err != nil {
			return fmt.Errorf("pio [%v]: %w", pioBlockOperator.Label(), err)
		}
	}

	pio0Operator := r.pioBlocks[0].Operator()
	pio0PinOutputs := pio0Operator.PinOutputs()
	pio0PinOutputEnables := pio0Operator.PinOutputEnables()
	pio1Operator := r.pioBlocks[1].Operator()
	pio1PinOutputs := pio1Operator.PinOutputs()
	pio1PinOutputEnables := pio1Operator.PinOutputEnables()

	for i, p := range r.pads {
		padOperator := p.Operator()
		gpioOperator := r.gpios[i].Operator()
		f := gpioOperator.GetFunction()
		switch f {
		case gpio.FunctionNone:
			padOperator.SetOutput(false)
			padOperator.SetOutputEnabled(false)
		case gpio.FunctionPIO0:
			padOperator.SetOutput((pio0PinOutputs>>i)&0b1 == 1)
			padOperator.SetOutputEnabled((pio0PinOutputEnables>>i)&0b1 == 1)
		case gpio.FunctionPIO1:
			padOperator.SetOutput((pio1PinOutputs>>i)&0b1 == 1)
			padOperator.SetOutputEnabled((pio1PinOutputEnables>>i)&0b1 == 1)
		default:
			return fmt.Errorf("%w: %d", ErrGPIOFunctionInvalid, f)
		}
	}

	return nil
}
