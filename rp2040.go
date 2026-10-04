package pio

import (
	"fmt"

	"github.com/headblockhead/pio/internal/gpio"
	"github.com/headblockhead/pio/internal/pad"
	"github.com/headblockhead/pio/internal/pio"
)

type RP2040 struct {
	pios  [2]*pio.PIO
	gpios [30]*gpio.GPIO
	pads  [30]*pad.Pad
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
