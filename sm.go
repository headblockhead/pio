package pio

import (
	"errors"
	"fmt"
	"math"
)

type StatusSelection uint
type ShiftDirection uint
type SMInstructionType uint

const (
	// If TX FIFO level is < statusComparisonLevel, returns all '1's.
	StatusSelectionTXLevel StatusSelection = iota
	// If RX FIFO level is < statusComparisonLevel, returns all '1's.
	StatusSelectionRXLevel

	ShiftDirectionRight ShiftDirection = iota
	ShiftDirectionLeft

	SMInstructionJump     SMInstructionType = 0b000
	SMInstructionWait     SMInstructionType = 0b001
	SMInstructionIn       SMInstructionType = 0b010
	SMInstructionOut      SMInstructionType = 0b011
	SMInstructionPushPull SMInstructionType = 0b100
	SMInstructionMove     SMInstructionType = 0b101
	SMInstructionIRQ      SMInstructionType = 0b110
	SMInstructionSet      SMInstructionType = 0b111
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
	// If stickyOutSetAssertion is enabled, a '0' bit also prevents the write from being 'stickied'.
	inlineOutWriteEnableIsUsed bool
	// Which bit of OUT data is used for inlineOutWriteEnable.
	inlineOutWriteEnableBit uint
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
	sidesetBitCount uint
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
	// The amount of cycles of delay left to be completed.
	delays uint

	// The type of the instruction that is stored in currentInstruction.
	currentInstructionType SMInstructionType

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

func (sm *SM) ScheduleForcedInstruction(instruction uint16) error {
	sm.forcedInstruction = instruction
	sm.newForcedInstruction = true
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

func (sm *SM) Decode() error {
	sm.currentInstructionType = (SMInstructionType)(sm.currentInstruction>>13) & 0b111
	return nil
}

var ErrSMInvalidInstructionType = errors.New("invalid instruction type")

func (sm *SM) Execute() error {
	instr := sm.currentInstruction
	switch sm.currentInstructionType {
	case SMInstructionJump:
		condition := (JumpCondition)(instr>>5) & 0b111
		address := (uint)(instr) & 0b11111
		err := sm.ExecuteJump(condition, address)
		if err != nil {
			return fmt.Errorf("error excecuting jump: %w", err)
		}
	case SMInstructionWait:
		polarity := ((instr >> 7) & 0b1) == 1
		source := (WaitSource)(instr>>5) & 0b11
		index := (uint)(instr) & 0b11111
		err := sm.ExecuteWait(polarity, source, index)
		if err != nil {
			return fmt.Errorf("error excecuting wait: %w", err)
		}
	case SMInstructionIn:
		source := (InSource)(instr>>5) & 0b111
		numberOfBits := (uint)(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		err := sm.ExecuteIn(source, numberOfBits)
		if err != nil {
			return fmt.Errorf("error excecuting in: %w", err)
		}
	case SMInstructionOut:
		destination := (OutDestination)(instr>>5) & 0b111
		numberOfBits := (uint)(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		err := sm.ExecuteOut(destination, numberOfBits)
		if err != nil {
			return fmt.Errorf("error excecuting out: %w", err)
		}
	case SMInstructionPushPull:
		err := sm.ExecutePushOrPull()
		if err != nil {
			return fmt.Errorf("error excecuting push/pull: %w", err)
		}
	case SMInstructionMove:

		// remember, MOV PINS also affected by inlineOutWriteEnable

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

var ErrSMInvalidShiftDirection = errors.New("invalid shift direction")

var ErrSMInInvalidSource = errors.New("invalid in source")
var ErrSMInInvalidPinState = errors.New("invalid pin state")

func (sm *SM) ExecuteIn(source InSource, bits uint) error {
	if !sm.stalled {
		var inData uint32
		var i uint
		switch source {
		case InSourcePins:
			for i = 0; i < bits; i++ {
				pinNumber := (sm.inBase + i) % 32
				pinState, err := sm.pins.GetInput(pinNumber)
				if err != nil {
					return fmt.Errorf("error getting input to pin %d: %w", pinNumber, err)
				}
				switch pinState {
				case PinInputHigh:
					inData = (inData << 1) | 0b1
				case PinInputLow:
					inData = (inData << 1) | 0b0
				default:
					return ErrSMInInvalidPinState
				}
			}
		case InSourceX:
			for i = 0; i < bits; i++ {
				bit := ((sm.x >> i) & 0b1)
				inData = (inData << 1) | bit
			}
		case InSourceY:
			for i = 0; i < bits; i++ {
				bit := ((sm.y >> i) & 0b1)
				inData = (inData << 1) | bit
			}
		case InSourceNull:
			// Data is all zeros, no action needed.
		case InSourceISR:
			for i = 0; i < bits; i++ {
				bit := ((sm.inputShiftRegister >> i) & 0b1)
				inData = (inData << 1) | bit
			}
		case InSourceOSR:
			for i = 0; i < bits; i++ {
				bit := ((sm.outputShiftRegister >> i) & 0b1)
				inData = (inData << 1) | bit
			}
		default:
			return ErrSMInInvalidSource
		}
		switch sm.inShiftdir {
		case ShiftDirectionLeft:
			for i = 0; i < bits; i++ {
				bit := (inData >> i) & 0b1
				sm.inputShiftRegister = (sm.inputShiftRegister << 1) | bit
			}
		case ShiftDirectionRight:
			for i = 0; i < bits; i++ {
				bit := (inData >> (bits - (i + 1))) & 0b1
				rightAlignedBit := (bit << 31)
				sm.inputShiftRegister = (sm.inputShiftRegister >> 1) | rightAlignedBit
			}
		default:
			return ErrSMInvalidShiftDirection
		}
		sm.inputShiftRegisterCounter += bits
	}
	sm.stalled = false
	if sm.autopush && sm.inputShiftRegisterCounter >= sm.pushThreshold {
		if !sm.fifoRX.IsFull() {
			err := sm.fifoRX.Write(sm.inputShiftRegister)
			if err != nil {
				return fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			sm.inputShiftRegister = 0
			sm.inputShiftRegisterCounter = 0
		} else {
			sm.stalled = true
		}
	}
	return nil
}

type OutDestination uint

const (
	OutDestinationPins               = 0b000
	OutDestinationX                  = 0b001
	OutDestinationY                  = 0b010
	OutDestinationNull               = 0b011
	OutDestinationPinDirections      = 0b100
	OutDestinationProgramCounter     = 0b101
	OutDestinationInputShiftRegister = 0b110
	OutDestinationEXEC               = 0b111
)

var ErrSMOutInvalidDestination = errors.New("invalid out destination")

func (sm *SM) ExecuteOut(destination OutDestination, bits uint) error {
	sm.stalled = false
	if sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold {
		if !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return fmt.Errorf("error reading TX FIFO for pre-out autopull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		} else {
			sm.stalled = true
		}
	} else {
		var outData uint32
		var i uint
		switch sm.outShiftdir {
		case ShiftDirectionLeft:
			for i = 0; i < bits; i++ {
				bit := ((sm.outputShiftRegister << i) & 0b10000000000000000000000000000000) >> 31
				outData = (outData << 1) | bit
			}
			sm.outputShiftRegister = (sm.outputShiftRegister << bits)
		case ShiftDirectionRight:
			for i = 0; i < bits; i++ {
				bit := (sm.outputShiftRegister >> (bits - (i + 1))) & 0b1
				outData = (outData << 1) | bit
			}
			sm.outputShiftRegister = (sm.outputShiftRegister >> bits)
		default:
			return ErrSMInvalidShiftDirection
		}
		sm.outputShiftRegisterCounter += bits

		var doWrite bool = true
		if sm.inlineOutWriteEnableIsUsed {
			doWrite = ((outData >> sm.inlineOutWriteEnableBit) & 0b1) == 1
		}

		sm.stickyWriteRecords = []StickyWriteRecord{}

		switch destination {
		case OutDestinationPins:
			if doWrite {
				for i = 0; i < sm.outCount; i++ {
					pinNumber := (sm.outBase + i) % 32
					bit := (((outData >> i) & 0b1) == 1)
					var err error
					var output PinOutput
					if bit == true {
						output = PinOutputHigh
					} else if bit == false {
						output = PinOutputLow
					}
					err = sm.pins.SetOutput(pinNumber, output)
					if err != nil {
						return fmt.Errorf("error setting pin %d's output: %w", pinNumber, err)
					}
					sm.stickyWriteRecords = append(sm.stickyWriteRecords, StickyWriteRecord{
						pin:    pinNumber,
						output: &output,
					})
				}
			}
		case OutDestinationX:
			sm.x = outData
		case OutDestinationY:
			sm.y = outData
		case OutDestinationNull:
			// No action required, discards data.
		case OutDestinationPinDirections:
			if doWrite {
				for i = 0; i < sm.outCount; i++ {
					pinNumber := (sm.outBase + i) % 32
					bit := (((outData >> i) & 0b1) == 1)
					var err error
					var outputEnable PinOutputEnable
					if bit == true {
						outputEnable = PinOutputEnabled
					} else if bit == false {
						outputEnable = PinOutputNotEnabled
					}
					err = sm.pins.SetOutputEnable(pinNumber, outputEnable)
					if err != nil {
						return fmt.Errorf("error setting pin %d's output enable: %w", pinNumber, err)
					}
					sm.stickyWriteRecords = append(sm.stickyWriteRecords, StickyWriteRecord{
						pin:          pinNumber,
						outputEnable: &outputEnable,
					})
				}
			}
		case OutDestinationProgramCounter:
			sm.programCounter = (uint)(outData % 32)
			sm.jumped = true
		case OutDestinationInputShiftRegister:
			sm.inputShiftRegister = outData
			sm.inputShiftRegisterCounter = bits
		case OutDestinationEXEC:
			sm.execdInstruction = (uint16)(outData)
			sm.newEXECdInstruction = true
		default:
			return ErrSMOutInvalidDestination
		}
		if sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return fmt.Errorf("error reading TX FIFO for post-out autopull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		}
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
	if !sm.newEXECdInstruction && sm.delays > 0 {
		sm.delays--
		return nil
	}

	if sm.newEXECdInstruction {
		sm.currentInstruction = sm.execdInstruction
		sm.newEXECdInstruction = false
	} else if !sm.stalled {
		err := sm.Fetch()
		if err != nil {
			return fmt.Errorf("error fetching: %w", err)
		}
	}

	err := sm.Decode()
	if err != nil {
		return fmt.Errorf("error decoding: %w", err)
	}

	if sm.currentInstructionType != SMInstructionOut && sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.IsEmpty() {
		osr, err := sm.fifoTX.Read()
		if err != nil {
			return fmt.Errorf("error reading TX FIFO for autopull: %w", err)
		}
		sm.outputShiftRegister = osr
		sm.outputShiftRegisterCounter = 0
	}

	sm.jumped = false
	err = sm.Execute()
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

	var doSideSet bool = sm.sidesetBitCount > 0
	if doSideSet && sm.sidesetIsOptional {
		doSideSet = ((sm.currentInstruction >> 12) & 0b1) == 1
	}
	if doSideSet {
		bits := sm.sidesetBitCount
		if sm.sidesetIsOptional {
			bits -= 1
		}
		var i uint
		for i = 0; i < bits; i++ {
			bitIndex := (13 - sm.sidesetBitCount) + i
			bit := ((sm.currentInstruction >> bitIndex) & 0b1) == 1
			if sm.sidesetControlsPinDirection {
				var outputEnable PinOutputEnable
				if bit == true {
					outputEnable = PinOutputEnabled
				} else if bit == false {
					outputEnable = PinOutputNotEnabled
				}
				pinNumber := sm.sidesetBase + i
				err := sm.pins.SetOutputEnable(pinNumber, outputEnable)
				if err != nil {
					return fmt.Errorf("error sideset-ing pin %d's output enable: %w", pinNumber, err)
				}
			} else {
				var output PinOutput
				if bit == true {
					output = PinOutputHigh
				} else if bit == false {
					output = PinOutputLow
				}
				pinNumber := sm.sidesetBase + i
				err := sm.pins.SetOutput(pinNumber, output)
				if err != nil {
					return fmt.Errorf("error sideset-ing pin %d's output: %w", pinNumber, err)
				}
			}
		}
	}
	var doDelays bool = sm.sidesetBitCount < 5
	if doDelays {
		var delayMask uint = (1 << (5 - sm.sidesetBitCount)) - 1
		sm.delays = (uint)(sm.currentInstruction>>8) & delayMask
	}

	return nil
}
