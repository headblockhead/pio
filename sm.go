package pio

import (
	"errors"
	"fmt"
)

type StatusSelection uint
type ShiftDirection uint

const (
	// If TX FIFO level is < statusComparisonLevel, returns all '1's.
	StatusSelectionTXLevel StatusSelection = iota
	// If RX FIFO level is < statusComparisonLevel, returns all '1's.
	StatusSelectionRXLevel

	ShiftDirectionRight ShiftDirection = iota
	ShiftDirectionLeft
)

type SM struct {
	id uint

	instructionMemory InstructionMemoryReader
	fifoRX            *FIFO
	fifoTX            *FIFO
	pins              PinsSMs
	irqs              IRQSMs

	// Use the most significant bit of the delay/side-set field in each instruction as a 'side set enable' bit.
	// This allows instructions to perform side-set optionally, rather than on every instruction.
	sidesetIsOptional bool
	// Use sideset to control the direction of pins, rather than the output value of pins.
	sidesetControlsPinDirection bool
	// Use a bit from OUT data to enable/disable writing pin data.
	// If stickyOutSetAssertion is enabled, a '0' bit deasserts the latest write.
	inlineOutEnableIsUsed bool
	// Which bit of OUT data is used for inlineOutEnable.
	inlineOutEnableBit uint
	// Perform the most recent OUT/SET pin data write repeatedly every tick.
	stickyOutSetAssertion bool
	// When the programCounter reaches wrapFromAddress, it will be set to wrapToAddress.
	// However, if the instruction is a jump, and the condition is true, the jump takes priority.
	wrapFromAddress uint
	// The address the programCounter is set to after reaching wrapFromAddress.
	wrapToAddress uint
	// Select which FIFO is used for the 'MOV x, STATUS' instruction.
	statusSelection StatusSelection
	// Choose a level which the selected FIFO's level will be compared to.
	statusComparisonLevel uint
	// Number of bits shifted out of the outputShiftRegister before an autopull (or pull ifempty) will happen.
	// Can be a value of 1-32.
	pullThreshold uint
	// Number of bits shifted into the inputShiftRegister before an autopush (or push iffull) will happen.
	// Can be a value of 1-32.
	pushThreshold uint
	// Direction the outputShiftRegister is shifted in as data is taken out.
	outShiftdir ShiftDirection
	// Direction the inputShiftRegister is shifted in as data is put in.
	inShiftdir ShiftDirection
	// Perform a pull automatically when the outputShiftRegister is emptied (eg on/following an OUT which causes the outputShiftRegisterCounter to be >= pullThreshold)
	autopull bool
	// Perform a push automatically when the inputShiftRegister is filled (eg on an IN which causes the inputShiftRegisterCounter to be >= pushThreshold)
	autopush bool
	// The lowest pin affected by a sideset operation, which takes the value of the least significant bit of the sideset portion of the delay/side-set field.
	sidesetBase uint
	// The number of bits (starting from the most significant bit) used for side-set values.
	// This includes the bit used for sidesetIsOptional, if that is enabled.
	sidesetCount uint
	// The lowest pin affected by a SET PINS/PINDIRS instruction, which takes the value of the least significant bit.
	setBase uint
	// The number of pins asserted by a SET instruction.
	// Can be a value of 0-5.
	setCount uint
	// The pin read into the least significant bit of an IN instruction's input.
	// Consecutively higher pins are mapped to consecutively more significant bits.
	inBase uint
	// The lowest pin affected by an OUT PINS/PINDIRS or MOV PINS instruction, which takes the value of the least significant bit.
	outBase uint
	// The number of pins asserted by an OUT PINS/PINDIRS or MOV PINS instruction.
	// Can be a value of 0-32.
	outCount uint
	// The pin used for the JMP PIN instruction.
	jumpPin uint

	// current instruction address
	programCounter uint
	// currently executing instruction
	currentInstruction uint16

	// state flags

	stalled     bool
	execStalled bool

	sticky      bool
	stickyInstr uint16
	exec        bool
	execInstr   uint16

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	x                          uint32
	y                          uint32
}

func NewSM(id uint, instructionMemory InstructionMemoryReader, pins PinsSMs, irqs IRQSMs) *SM {
	return &SM{
		id: id,

		instructionMemory: instructionMemory,
		fifoRX:            NewFIFO(4),
		fifoTX:            NewFIFO(4),
		pins:              pins,
		irqs:              irqs,

		wrapFromAddress: 31,
		setCount:        5,

		osrShiftCounter: 32,
	}
}

type FIFOJoinMode uint

const (
	FIFOJoinNone FIFOJoinMode = iota
	FIFOJoinRX
	FIFOJoinTX
)

var ErrSMInvalidFIFOJoinMode = errors.New("invalid fifo join mode")

func (sm *SM) SetFIFOJoinMode(joinMode FIFOJoinMode) error {
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
	default:
		return ErrSMInvalidFIFOJoinMode
	}
	return nil
}

func (sm *SM) Fetch() error {
	var err error
	sm.instr, err = sm.instructionMemory.Read(sm.pc)
	if err != nil {
		return fmt.Errorf("error fetching instruction from address %d: %w", sm.pc, err)
	}
	return nil
}

type SMInstructionType uint

const (
	SMInstructionJump     SMInstructionType = 0b000
	SMInstructionWait     SMInstructionType = 0b001
	SMInstructionIn       SMInstructionType = 0b010
	SMInstructionOut      SMInstructionType = 0b011
	SMInstructionPushPull SMInstructionType = 0b100
	SMInstructionMove     SMInstructionType = 0b101
	SMInstructionIRQ      SMInstructionType = 0b110
	SMInstructionSet      SMInstructionType = 0b111
)

var ErrSMInvalidInstructionType = errors.New("invalid instruction type")

func (sm *SM) ForceInstruction(instruction uint16) error {
	sm.execInstr = instruction
	sm.exec = true
	return nil
}

func (sm *SM) Execute() error {
	var instructionType SMInstructionType = (SMInstructionType)(sm.instr>>13) & 0b111
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
	default:
		return ErrSMInvalidInstructionType
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

var ErrSMJumpInvalidCondition = errors.New("invalid condition")

func (sm *SM) ExecuteJump(condition JumpCondition, address uint) error {
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
		input, err := sm.pins.GetInput(sm.jumpPin)
		if err != nil {
			return fmt.Errorf("error getting input of pin %d: %w", sm.jumpPin, err)
		}
		shouldJump = (input == PinInputHigh)
	case JumpOSRENotEmpty:
		shouldJump = (sm.osrShiftCounter != sm.pullThreshold)
	default:
		return ErrSMJumpInvalidCondition
	}
	if shouldJump {
		sm.pc = address
		sm.justJumped = true
	}
	return nil
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

func (sm *SM) SystemTick() error {

	if sm.exec == true {
		sm.Execute()

		// only if stalls
		sm.instr = sm.execInstr
	}
}

func (sm *SM) Tick() error {

	//  do instructions

	if !sm.justJumped {
		if sm.pc == sm.wrapTop {
			sm.pc = sm.wrapBottom
		} else {
			sm.pc = (sm.pc + 1) & 0b11111
		}
	}

	if sm.sticky == true {
	}

}
