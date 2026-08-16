package pio

type StatusSelection int

const (
	StatusSelectionTXLevel StatusSelection = iota
	StatusSelectionRXLevel
)

type SM struct {
	id uint

	im   InstructionMemoryReader
	rx   FIFOWriter
	tx   FIFOReader
	irq  IRQ
	pins [32]PinPeripheralIO

	stalled     bool

	sidesetOptional      bool
	sidesetControlsPinDirection      bool
	jumpPin         uint
	inlineOutEnableBit    uint
	inlineOutEnable bool
	sticky       bool
	wrapTop         uint
	wrapBottom      uint
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
	stickyInstr uint16

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
	if sm.addr == sm.wrapTop {
		sm.addr = sm.wrapBottom
	} else {
		sm.addr = (sm.addr + 1) & 0b11111
	}
	return nil
}

type SMInstruction uint

const (
	SMInstructionJump     SMInstruction = 0b000
	SMInstructionWait                   = 0b001
	SMInstructionIn                     = 0b010
	SMInstructionOut                    = 0b011
	SMInstructionPushPull               = 0b100
	SMInstructionMove                   = 0b101
	SMInstructionIRQ                    = 0b110
	SMInstructionSet                    = 0b111
)

func (sm *SM) Execute() error {
	var instructionType SMInstruction = (SMInstruction)(sm.instr>>13) & 0b111
	switch instructionType {
	case SMInstructionJump:
		sm.ExecuteJump((JumpCondition)(sm.instr>>5)&0b111, (uint)(sm.instr)&0b11111)
	case SMInstructionWait:
		sm.ExecuteWait(((sm.instr>>7)&0b1) == 1, (WaitSource)(sm.instr>>5)&0b11, (uint)(sm.instr)&0b11111)
	case SMInstructionIn:
		sm.ExecuteIn((InSource)(sm.instr>>5)&0b111, (uint)(sm.instr&0b11111))
	case SMInstructionOut:
		sm.ExecuteOut()
	case SMInstructionPushPull:
		sm.ExecutePushOrPull()
	case SMInstructionMove:
		sm.ExecuteMove()
	case SMInstructionIRQ:
		sm.ExecuteIRQ()
	case SMInstructionSet:
		sm.ExecuteSet()
	}

}

type JumpCondition uint

const (
	JumpAlways                JumpCondition = 0b000
	JumpXZero                               = 0b001
	JumpXNonZeroThenDecrement               = 0b010
	JumpYZero                               = 0b011
	JumpYNonZeroThenDecrement               = 0b100
	JumpXNotEqualY                          = 0b101
	JumpPin                                 = 0b110
	JumpOSRENotEmpty                        = 0b111
)

func (sm *SM) ExecuteJump(condition JumpCondition, address uint) {
	var shouldJump bool
	switch condition {
	case JumpAlways:
		shouldJump = true
	case JumpXZero:
		shouldJump = (sm.x == 0)
	case JumpXNonZeroThenDecrement:
		shouldJump = (sm.x != 0)
		sm.x--
	case JumpYZero:
		shouldJump = (sm.y == 0)
	case JumpYNonZeroThenDecrement:
		shouldJump = (sm.y != 0)
		sm.y--
	case JumpXNotEqualY:
		shouldJump = (sm.x != sm.y)
	case JumpPin:
		shouldJump = (sm.pins[sm.jumpPin].GetInput())
	case JumpOSRENotEmpty:
		shouldJump = sm.osrShiftCounter != sm.pullThreshold
	}
	if shouldJump {
		sm.addr = address
	}
}

type WaitSource uint

const (
	WaitSourceGPIO WaitSource = 0b00
	WaitSourcePin             = 0b01
	WaitSourceIRQ             = 0b10
)

func (sm *SM) ExecuteWait(polarity bool, source WaitSource, index uint) error {
	var continueWaiting bool
	switch source {
	case WaitSourceGPIO:
		continueWaiting = (sm.pins[index].GetInput() == polarity)
	case WaitSourcePin:
		continueWaiting = (sm.pins[(sm.inBase+index)%32].GetInput() == polarity)
	case WaitSourceIRQ:
		var relative bool = ((index >> 4) & 0b1) == 1
		var irqIndex uint = index & 0b111
		if relative {
			var lowerBits uint = index & 0b11
			var upperBit uint = index & 0b100
			lowerBits = (lowerBits + sm.id) & 0b11
			irqIndex = upperBit | lowerBits
		}
		irqSet, err := sm.irq.Read(irqIndex)
		if err != nil {
			return err
		}
		continueWaiting = irqSet == polarity
		if polarity == true && !continueWaiting {
			sm.irq.Clear(irqIndex)
		}
	}
	sm.execStalled = continueWaiting
	return nil
}

type InSource uint

const (
	InSourcePins InSource = 0b000
	InSourceX             = 0b001
	InSourceY             = 0b010
	InSourceNull          = 0b011
	InSourceISR           = 0b110
	InSourceOSR           = 0b111
)

func (sm *SM) ExecuteIn(source InSource, bits uint) error {
	switch source {
	case InSourcePins:
		if sm.inShiftdir == 
		sm.isr
	}
}
