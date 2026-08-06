package pio

type SM struct {
	readPort InstructionMemoryRead
	rx       FIFOWrite
	tx       FIFORead
	irq      IRQ

	execStalled     bool
	sideEnable      bool
	sidePindir      bool
	jumpPin         uint
	OutEnableBit    uint
	InlineOutEnable bool
	OutSticky       bool
	WrapTop         uint
	WrapBottom      uint // Target

	joinRX        bool
	joinTX        bool
	pullThreshold uint
	pushThreshold uint
	outShiftdir   bool
	inShiftdir    bool
	autopull      bool
	autopush      bool

	instr           uint16
	osr             uint32
	osrShiftCounter uint
	isr             uint32
	isrShiftCounter uint
	x               uint32
	y               uint32
	pc              uint
}
