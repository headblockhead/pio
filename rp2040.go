package pio

import (
	"errors"
	"fmt"

	"github.com/headblockhead/pio/conn"
	"github.com/headblockhead/pio/internal/gpio"
	"github.com/headblockhead/pio/internal/pad"
	"github.com/headblockhead/pio/internal/pio"
)

const rp2040PadCount = 30

type RP2040 struct {
	pios  [2]*pio.PIO
	gpios [rp2040PadCount]*gpio.GPIO
	pads  [rp2040PadCount]*pad.Pad
}

func NewRP2040(id string) *RP2040 {
	r := &RP2040{}
	for i := range 2 {
		r.pios[i] = pio.NewPIO(32, 4)
	}
	for i := range 30 {
		r.gpios[i] = gpio.NewGPIO()
		r.pads[i] = pad.NewPad(fmt.Sprintf("%s_pad%d", id, i))
	}
	return r
}

func (r *RP2040) Connection(i uint) (c conn.Connection, ok bool) {
	if i >= rp2040PadCount {
		return nil, false
	}
	return r.pads[i].Connection(), true
}

func (r *RP2040) ConnectionBlock() conn.ConnectionBlock {
	return r
}

func (r *RP2040) PIOObserver(i uint) (o pio.Observer, ok bool) {
	if i >= 2 {
		return nil, false
	}
	return r.pios[i].Observer(), true
}

func (r *RP2040) PIOConfigurator(i uint) (c pio.Configurator, ok bool) {
	if i >= 2 {
		return nil, false
	}
	return r.pios[i].Configurator(), true
}

func (r *RP2040) GPIOObserver(i uint) (o gpio.Observer, ok bool) {
	if i >= rp2040PadCount {
		return nil, false
	}
	return r.gpios[i].Observer(), true
}

func (r *RP2040) GPIOConfigurator(i uint) (c gpio.Configurator, ok bool) {
	if i >= rp2040PadCount {
		return nil, false
	}
	return r.gpios[i].Configurator(), true
}

func (r *RP2040) PadObserver(i uint) (o pad.Observer, ok bool) {
	if i >= rp2040PadCount {
		return nil, false
	}
	return r.pads[i].Observer(), true
}

func (r *RP2040) PadConfigurator(i uint) (c pad.Configurator, ok bool) {
	if i >= rp2040PadCount {
		return nil, false
	}
	return r.pads[i].Configurator(), true
}

var ErrGPIOFunctionInvalid = errors.New("gpio function invalid")

func (r *RP2040) Tick() error {
	var pinInputs uint32
	for i, p := range r.pads {
		cp := p.Controller()

		inputHigh := cp.GetInput()
		if inputHigh {
			pinInputs |= (0b1 << i)
		}

		err := cp.Tick()
		if err != nil {
			return fmt.Errorf("pad %s: %w", cp.ID(), err)
		}
	}

	for i, p := range r.pios {
		c := p.Controller()
		c.SetPinInputs(pinInputs)
		err := c.Tick()
		if err != nil {
			return fmt.Errorf("pio %d: %w", i, err)
		}
	}
	pio0C := r.pios[0].Controller()
	pio0PinOutputs := pio0C.PinOutputs()
	pio0PinOutputEnables := pio0C.PinOutputEnables()
	pio1C := r.pios[1].Controller()
	pio1PinOutputs := pio1C.PinOutputs()
	pio1PinOutputEnables := pio1C.PinOutputEnables()

	for i, p := range r.pads {
		cp := p.Controller()
		cg := r.gpios[i].Controller()
		f := cg.GetFunction()
		switch f {
		case gpio.FunctionNone:
			// no action
		case gpio.FunctionPIO0:
			outputHigh := (pio0PinOutputs>>i)&0b1 == 1
			outputEnabled := (pio0PinOutputEnables>>i)&0b1 == 1
			cp.SetOutput(outputHigh)
			cp.SetOutputEnabled(outputEnabled)
		case gpio.FunctionPIO1:
			outputHigh := (pio1PinOutputs>>i)&0b1 == 1
			outputEnabled := (pio1PinOutputEnables>>i)&0b1 == 1
			cp.SetOutput(outputHigh)
			cp.SetOutputEnabled(outputEnabled)
		default:
			return ErrGPIOFunctionInvalid
		}
	}

	return nil
}
