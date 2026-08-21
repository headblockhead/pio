package pio

type irq struct {
	set bool
}

func newIRQ() *irq {
	return &irq{}
}
