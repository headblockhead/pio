package pio

const smCount = 4
const irqCount = 8
const pinCount = 32

type pio struct {
	im   *im
	sms  [smCount]*sm
	irqs [irqCount]*irq
	pins [pinCount]*pin
}

func newPIO() *pio {
	p := &pio{}
	p.im = newIM()
	for i := range smCount {
		p.sms[i] = newSM((uint)(i), p.im)
	}
	for i := range irqCount {
		p.irqs[i] = newIRQ()
	}
	for i := range pinCount {
		p.pins[i] = newPin()
	}
	return p
}

func (p *pio) tick() error {

	// TODO: take the output, output mask, output enable, and output enable mask values from each state machine

	return nil
}
