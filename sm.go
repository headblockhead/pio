package pio

type StatusSelection uint

const (
	StatusSelectionTXLevel StatusSelection = iota
	StatusSelectionRXLevel
)

type SM struct {
	id uint

	instructionMemory InstructionMemoryReader
	fifoRX            *FIFO
	fifoTX            *FIFO
	pins              PinsSMs
	irqs              IRQSMs

	stalled bool

	sidesetOptional             bool
	sidesetControlsPinDirection bool
	jumpPin                     uint
	inlineOutEnableBit          uint
	inlineOutEnable             bool
	stickyOut                   bool
	wrapTop                     uint
	wrapBottom                  uint
	statusSelection             StatusSelection
	statusComparisonLevel       uint

	pullThreshold uint
	pushThreshold uint
	outShiftdir   bool
	inShiftdir    bool
	autopull      bool
	autopush      bool

	addr        uint
	instr       uint16
	stickyInstr uint16
	systemInstr uint16

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

func NewSM(id uint, instructionMemory InstructionMemoryReader, pins PinsSMs, irqs IRQSMs) *SM {
	return &SM{
		id:                id,
		instructionMemory: instructionMemory,
		fifoRX:            NewFIFO(4),
		fifoTX:            NewFIFO(4),
		pins:              pins,
		irqs:              irqs,
	}
}

type FIFOJoinMode uint

const (
	FIFOJoinNone FIFOJoinMode = iota
	FIFOJoinRX
	FIFOJoinTX
)

func (sm *SM) SetFIFOJoinMode(joinMode FIFOJoinMode) {
	switch joinMode {
	case FIFOJoinNone:
		sm.fifoRX = NewFIFO(4)
		sm.fifoTX = NewFIFO(4)
	case FIFOJoinRX:
		sm.fifoRX = NewFIFO(8)
		sm.fifoTX = NewFIFO(0)
	case FIFOJoinTX:
		sm.fifoRX = NewFIFO(0)
		sm.fifoTX = NewFIFO(8)
	}
}

func (sm *SM) Fetch() error {
	var err error
	sm.instr, err = sm.instructionMemory.Read(sm.addr)
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
	SMInstructionWait     SMInstruction = 0b001
	SMInstructionIn       SMInstruction = 0b010
	SMInstructionOut      SMInstruction = 0b011
	SMInstructionPushPull SMInstruction = 0b100
	SMInstructionMove     SMInstruction = 0b101
	SMInstructionIRQ      SMInstruction = 0b110
	SMInstructionSet      SMInstruction = 0b111
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
	JumpXZero                 JumpCondition = 0b001
	JumpXNonZeroThenDecrement JumpCondition = 0b010
	JumpYZero                 JumpCondition = 0b011
	JumpYNonZeroThenDecrement JumpCondition = 0b100
	JumpXNotEqualY            JumpCondition = 0b101
	JumpPin                   JumpCondition = 0b110
	JumpOSRENotEmpty          JumpCondition = 0b111
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
	WaitSourcePin  WaitSource = 0b01
	WaitSourceIRQ  WaitSource = 0b10
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
		irqState, err := sm.irqs.Read(irqIndex)
		if err != nil {
			return err
		}

		// TODO

		if polarity == true && !continueWaiting {
			sm.irqs.Clear(irqIndex)
		}
	}
	sm.execStalled = continueWaiting
	return nil
}

type InSource uint

const (
	InSourcePins InSource = 0b000
	InSourceX    InSource = 0b001
	InSourceY    InSource = 0b010
	InSourceNull InSource = 0b011
	InSourceISR  InSource = 0b110
	InSourceOSR  InSource = 0b111
)

func (sm *SM) ExecuteIn(source InSource, bits uint) error {
	switch source {
	}
	return nil
}
