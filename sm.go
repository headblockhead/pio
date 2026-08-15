package pio

type StatusSelection int

const (
	StatusSelectionTXLevel StatusSelection = iota
	StatusSelectionRXLevel
)

type SM struct {
	im   InstructionMemoryReader
	rx   FIFOWriter
	tx   FIFOReader
	irq  IRQ
	pins [30]PinPeripheralIO

	execStalled     bool
	sideEnable      bool
	sidePindir      bool
	jumpPin         uint
	outEnableBit    uint
	inlineOutEnable bool
	outSticky       bool
	wrapTop         uint
	wrapBottom      uint // Target
	statusSelection StatusSelection
	statusN         uint

	joinRX        bool
	joinTX        bool
	pullThreshold uint
	pushThreshold uint
	outShiftdir   bool
	inShiftdir    bool
	autopull      bool
	autopush      bool

	addr  uint
	instr uint16

	sidesetCount uint
	setCount     uint
	outCount     uint
	inBase       uint
	sideSetBase  uint
	setBase      uint
	outBase      uint

	osr             uint32
	osrShiftCounter uint
	isr             uint32
	isrShiftCounter uint
	x               uint32
	y               uint32
}

func (sm *SM) Fetch() error {
	var err error
	sm.instr, err = sm.im.Read(sm.addr)
	if err != nil {
		return err
	}
	sm.addr++
	return nil
}

func (sm *SM) Execute() error {
	var instructionType uint = (uint)(sm.instr>>13) & 0b111
	switch instructionType {
	case 0b000:
		sm.ExecuteJump((uint)(sm.instr>>5)&0b111, (uint)(sm.instr)&0b11111)
	case 0b001:
		sm.ExecuteWait()
	case 0b010:
		sm.ExecuteIn()
	case 0b011:
		sm.ExecuteOut()
	case 0b100:
		sm.ExecutePushOrPull()
	case 0b101:
		sm.ExecuteMove()
	case 0b110:
		sm.ExecuteIRQ()
	case 0b111:
		sm.ExecuteSet()
	}
}

func (sm *SM) decrementX() {
	if sm.x == 0 {
		sm.x = 31
	} else {
		sm.x--
	}
}
func (sm *SM) decrementY() {
	if sm.y == 0 {
		sm.y = 31
	} else {
		sm.y--
	}
}
func (sm *SM) ExecuteJump(condition uint, address uint) {
	shouldJump := false
	switch condition {
	case 0b000:
		shouldJump = true
	case 0b001:
		shouldJump = (sm.x == 0)
	case 0b010:
		shouldJump = (sm.x != 0)
		sm.decrementX()
	case 0b011:
		shouldJump = (sm.y == 0)
	case 0b100:
		shouldJump = (sm.y != 0)
		sm.decrementY()
	case 0b101:
		shouldJump = (sm.x != sm.y)
	case 0b110:
		shouldJump = (sm.pins[sm.jumpPin].GetInput())
	case 0b111:
		shouldJump = sm.osrShiftCounter != sm.pullThreshold
	}
	if shouldJump {
		sm.addr = address
	}
}
