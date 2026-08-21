package pio

type pin struct {
	output       bool
	outputEnable bool
	input        bool
}

func newPin() *pin {
	return &pin{}
}
