package pio

import (
	"errors"
	"fmt"
	"math"
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

type StickyWriteRecord struct {
	pin          uint
	output       *PinOutput
	outputEnable *PinOutputEnable
}

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

	// Clock divider integer component
	clockDividerInteger uint16
	// Clock divider fractional component
	clockDividerFractional uint8
	// Internal clock accumulator
	clockAccumulator int32

	// Current instruction address.
	programCounter uint
	// Currently executing instruction.
	currentInstruction uint16
	// If the current instruction is stalled.
	stalled bool
	// If the instruction that just executed was a successful jump.
	jumped bool
	// If there is a forced instruction that is awaiting execution.
	newForcedInstruction bool
	// The forced instruction to be executed.
	forcedInstruction uint16
	// If there is an EXEC'd instruction awaiting execution.
	newEXECdInstruction bool
	// The EXEC'd instruction to be executed.
	execdInstruction uint16

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	x                          uint32
	y                          uint32

	stickyWriteRecords []StickyWriteRecord
}

func NewSM(id uint, instructionMemory InstructionMemoryReader, pins PinsSMs, irqs IRQSMs) *SM {
	return &SM{
		id: id,

		instructionMemory: instructionMemory,
		fifoRX:            NewFIFO(4),
		fifoTX:            NewFIFO(4),
		pins:              pins,
		irqs:              irqs,

		wrapFromAddress:            31,
		setCount:                   5,
		clockAccumulator:           clockDivisorAsQ8(0, 0),
		outputShiftRegisterCounter: 32,
	}
}

var ErrSMClockDivisorOutOfRange = errors.New("clock divisor out of range")

func (sm *SM) SetClockDivisor(divisor float64) error {
	if divisor < 1.0 || divisor > 65536.0 {
		return ErrSMClockDivisorOutOfRange
	}
	q8 := math.Round(divisor * 256)
	integer := uint32(q8) >> 8
	fractional := uint8(uint32(q8) & 0xFF)
	if integer == 65536 {
		integer = 0
		fractional = 0
	}
	sm.clockDividerInteger = uint16(integer)
	sm.clockDividerFractional = fractional
	return nil
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
	sm.currentInstruction, err = sm.instructionMemory.Read(sm.programCounter)
	if err != nil {
		return fmt.Errorf("error reading instruction from address %d: %w", sm.programCounter, err)
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

func (sm *SM) ScheduleForcedInstruction(instruction uint16) error {
	sm.forcedInstruction = instruction
	sm.newForcedInstruction = true
	return nil
}

var ErrSMInvalidInstructionType = errors.New("invalid instruction type")

func (sm *SM) Execute() error {
	instr := sm.currentInstruction
	var instructionType SMInstructionType = (SMInstructionType)(instr>>13) & 0b111
	switch instructionType {
	case SMInstructionJump:
		err := sm.ExecuteJump((JumpCondition)(instr>>5)&0b111, (uint)(instr)&0b11111)
		if err != nil {
			return fmt.Errorf("error excecuting jump: %w", err)
		}
	case SMInstructionWait:
		err := sm.ExecuteWait(((instr>>7)&0b1) == 1, (WaitSource)(instr>>5)&0b11, (uint)(instr)&0b11111)
		if err != nil {
			return fmt.Errorf("error excecuting wait: %w", err)
		}
	case SMInstructionIn:
		err := sm.ExecuteIn((InSource)(instr>>5)&0b111, (uint)(instr&0b11111))
		if err != nil {
			return fmt.Errorf("error excecuting in: %w", err)
		}
	case SMInstructionOut:
		err := sm.ExecuteOut()
		if err != nil {
			return fmt.Errorf("error excecuting out: %w", err)
		}
	case SMInstructionPushPull:
		err := sm.ExecutePushOrPull()
		if err != nil {
			return fmt.Errorf("error excecuting push/pull: %w", err)
		}
	case SMInstructionMove:
		err := sm.ExecuteMove()
		if err != nil {
			return fmt.Errorf("error excecuting move: %w", err)
		}
	case SMInstructionIRQ:
		err := sm.ExecuteIRQ()
		if err != nil {
			return fmt.Errorf("error excecuting irq: %w", err)
		}
	case SMInstructionSet:
		err := sm.ExecuteSet()
		if err != nil {
			return fmt.Errorf("error excecuting set: %w", err)
		}
	default:
		return ErrSMInvalidInstructionType
	}
	return nil
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

var ErrSMJumpInvalidCondition = errors.New("invalid jump condition")

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
		shouldJump = (sm.outputShiftRegisterCounter < sm.pullThreshold)
	default:
		return ErrSMJumpInvalidCondition
	}
	if shouldJump {
		sm.programCounter = address
		sm.jumped = true
	}
	return nil
}

type WaitSource uint

const (
	WaitSourceGPIO WaitSource = 0b00
	WaitSourcePin  WaitSource = 0b01
	WaitSourceIRQ  WaitSource = 0b10
)

var ErrSMWaitInvalidSource = errors.New("inavlid wait source")

func (sm *SM) ExecuteWait(polarity bool, source WaitSource, index uint) error {
	switch source {
	case WaitSourceGPIO:
		gpioInput, err := sm.pins.GetInput(index)
		if err != nil {
			return fmt.Errorf("error getting input of pin %d: %w", index, err)
		}
		if polarity == true {
			sm.stalled = !(gpioInput == PinInputHigh)
		} else if polarity == false {
			sm.stalled = !(gpioInput == PinInputLow)
		}
	case WaitSourcePin:
		pinIndex := (sm.inBase + index) % 32
		pinInput, err := sm.pins.GetInput(pinIndex)
		if err != nil {
			return fmt.Errorf("error getting input of pin %d: %w", pinIndex, err)
		}
		if polarity == true {
			sm.stalled = !(pinInput == PinInputHigh)
		} else if polarity == false {
			sm.stalled = !(pinInput == PinInputLow)
		}
	case WaitSourceIRQ:
		var relative bool = ((index >> 4) & 0b1) == 1
		var irqIndex uint = index & 0b111
		if relative {
			var upperBit uint = irqIndex & 0b100
			var lowerBits uint = irqIndex & 0b11
			lowerBits = (lowerBits + sm.id) & 0b11
			irqIndex = upperBit | lowerBits
		}
		irqState, err := sm.irqs.Read(irqIndex)
		if err != nil {
			return fmt.Errorf("error reading state of irq %d: %w", irqIndex, err)
		}
		if polarity == true {
			sm.stalled = !(irqState == IRQSet)
			if !sm.stalled {
				err := sm.irqs.Clear(irqIndex)
				if err != nil {
					return fmt.Errorf("error clearing irq %d: %w", irqIndex, err)
				}
			}
		} else if polarity == false {
			sm.stalled = !(irqState == IRQCleared)
		}
	default:
		return ErrSMWaitInvalidSource
	}
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

var ErrSMInInvalidSource = errors.New("invalid in source")

func (sm *SM) ExecuteIn(source InSource, bits uint) error {
	switch source {
	case InSourcePins:
	case InSourceX:
		// etc... TODO
	default:
		return ErrSMInInvalidSource
	}
	return nil
}

func clockDivisorAsQ8(integer uint16, fractional uint8) int32 {
	intPart := int32(integer)
	if intPart == 0 {
		intPart = 65536
	}
	return (intPart << 8) | int32(fractional)
}

func (sm *SM) SystemTick() error {
	sm.clockAccumulator -= 1 << 8
	if sm.stickyOutSetAssertion {
		for i, record := range sm.stickyWriteRecords {
			if record.output != nil {
				err := sm.pins.SetOutput(record.pin, *record.output)
				if err != nil {
					return fmt.Errorf("error applying sticky pin write %d for pin %d's output: %w", i, record.pin, err)
				}
			}
			if record.outputEnable != nil {
				err := sm.pins.SetOutputEnable(record.pin, *record.outputEnable)
				if err != nil {
					return fmt.Errorf("error applying sticky pin write %d for pin %d's output enable: %w", i, record.pin, err)
				}
			}
		}
	}
	if sm.newForcedInstruction {
		sm.currentInstruction = sm.forcedInstruction
		err := sm.Execute()
		if err != nil {
			return fmt.Errorf("error executing forced instruction: %w", err)
		}
		if !sm.stalled {
			sm.newForcedInstruction = false
		}
	} else if sm.clockAccumulator <= 0 {
		sm.clockAccumulator += clockDivisorAsQ8(sm.clockDividerInteger, sm.clockDividerFractional)
		err := sm.Tick()
		if err != nil {
			return fmt.Errorf("error performing tick: %w", err)
		}
	}
	return nil
}

func (sm *SM) Tick() error {
	if sm.newEXECdInstruction {
		sm.currentInstruction = sm.execdInstruction
		sm.newEXECdInstruction = false
	} else if !sm.stalled {
		err := sm.Fetch()
		if err != nil {
			return fmt.Errorf("error fetching: %w", err)
		}
	}
	err := sm.Execute()
	if err != nil {
		return fmt.Errorf("error executing an instruction: %w", err)
	}
	if !sm.stalled && !sm.newEXECdInstruction && !sm.jumped {
		if sm.programCounter == sm.wrapFromAddress {
			sm.programCounter = sm.wrapToAddress
		} else {
			sm.programCounter = (sm.programCounter + 1) % 32
		}
	}
	if sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.IsEmpty() {
		osr, err := sm.fifoTX.Read()
		if err != nil {
			return fmt.Errorf("error reading TX FIFO for autopull: %w", err)
		}
		sm.outputShiftRegister = osr
		sm.outputShiftRegisterCounter = 0
	}
	return nil
}
