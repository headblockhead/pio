package pio

import (
	"errors"
	"fmt"
)

const pioCount = 2
const padCount = 30

type rp2040 struct {
	id string

	pios  [pioCount]*pio
	gpios [padCount]*gpio
	pads  [padCount]*pad
}

func newRP2040(id string) *rp2040 {
	r := &rp2040{
		id: id,
	}
	for i := range pioCount {
		r.pios[i] = newPIO()
	}
	for i := range padCount {
		r.gpios[i] = newGPIO()
		r.pads[i] = newPad(fmt.Sprintf("%s:pad_%d", id, i))
	}
	return r
}

func (r *rp2040) ID() string {
	return r.id
}

func (r *rp2040) Tick() error {
	// Put input states of pads into the PIOs.
	for i := range padCount {
		input := r.pads[i].getInput()
		inputAfterGPIO, err := r.gpios[i].inputOverride.applyTo(input)
		if err != nil {
			return fmt.Errorf("error applying gpio %d's input override: %w", i, err)
		}
		pio := r.gpios[i].pioAssignment
		r.pios[pio].pins[i].input = inputAfterGPIO
	}

	// Update the PIOs.
	for i, pio := range r.pios {
		err := pio.tick()
		if err != nil {
			return fmt.Errorf("error ticking pio %d: %w", i, err)
		}
	}

	// Put output states of the PIOs into the pads.
	for i := range padCount {
		pio := r.gpios[i].pioAssignment

		output := r.pios[pio].pins[i].output
		outputAfterGPIO, err := r.gpios[i].outputOverride.applyTo(output)
		if err != nil {
			return fmt.Errorf("error applying gpio %d's output override: %w", i, err)
		}
		r.pads[i].output = outputAfterGPIO

		outputEnable := r.pios[pio].pins[i].outputEnable
		outputEnableAfterGPIO, err := r.gpios[i].outputEnableOverride.applyTo(outputEnable)
		if err != nil {
			return fmt.Errorf("error applying gpio %d's output enable override: %w", i, err)
		}
		r.pads[i].outputEnabled = outputEnableAfterGPIO
	}

	// Update the pads.
	for _, pad := range r.pads {
		pad.tick()
	}

	return nil
}

var ErrPadOutOfRange = errors.New("pad out of range")

func (r *rp2040) PadConnection(pad uint) (Connection, error) {
	if pad >= padCount {
		return nil, ErrPadOutOfRange
	}
	return r.pads[pad], nil
}

func (r *rp2040) SetPadPullUp(pad uint, pu bool) error {
	if pad >= padCount {
		return ErrPadOutOfRange
	}
	r.pads[pad].pullUp = pu
	return nil
}

func (r *rp2040) SetPadPullDown(pad uint, pd bool) error {
	if pad >= padCount {
		return ErrPadOutOfRange
	}
	r.pads[pad].pullDown = pd
	return nil
}

func (r *rp2040) SetPadOutputDisabled(pad uint, od bool) error {
	if pad >= padCount {
		return ErrPadOutOfRange
	}
	r.pads[pad].outputDisabled = od
	return nil
}

func (r *rp2040) SetPadInputEnabled(pad uint, ie bool) error {
	if pad >= padCount {
		return ErrPadOutOfRange
	}
	r.pads[pad].inputEnabled = ie
	return nil
}

func (r *rp2040) SetPadInputSynchroniserEnabled(pad uint, ise bool) error {
	if pad >= padCount {
		return ErrPadOutOfRange
	}
	r.pads[pad].inputSynchroniserEnabled = ise
	return nil
}

func (r *rp2040) SetGPIOOutputEnableOverride(gpio uint, o Override) error {
	if gpio >= padCount {
		return ErrPadOutOfRange
	}
	r.gpios[gpio].outputEnableOverride = o
	return nil
}

func (r *rp2040) SetGPIOOutputOverride(gpio uint, o Override) error {
	if gpio >= padCount {
		return ErrPadOutOfRange
	}
	r.gpios[gpio].outputOverride = o
	return nil
}

func (r *rp2040) SetGPIOInputOverride(gpio uint, o Override) error {
	if gpio >= padCount {
		return ErrPadOutOfRange
	}
	r.gpios[gpio].inputOverride = o
	return nil
}

func (r *rp2040) SetGPIOAssignment(gpio uint, pio uint) error {
	if gpio >= padCount {
		return ErrPadOutOfRange
	}
	if pio >= pioCount {
		return fmt.Errorf("pio out of range")
	}
	r.gpios[gpio].pioAssignment = pio
	return nil
}
