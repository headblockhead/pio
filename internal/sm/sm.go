package sm

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/headblockhead/pio/internal/fifo"
	"github.com/headblockhead/pio/internal/memory"
)

type SM struct {
	index uint

	memoryReader memory.MemoryReader
	fifoRX       *fifo.FIFO
	fifoTX       *fifo.FIFO

	pinInputs uint32
	irqInputs uint8

	enabled bool

	pinOutputEnables     uint32
	pinOutputEnablesMask uint32
	pinOutputs           uint32
	pinOutputMask        uint32
	pinSidesets          uint32
	pinSidesetsMask      uint32
	irqWrites            uint8
	irqWritesMask        uint8

	sidesetIsOptional            bool
	sidesetControlsPinDirection  bool
	inlineOutWriteEnableIsUsed   bool
	inlineOutWriteEnableBitIndex uint
	stickyOutSetAssertion        bool

	wrapFromAddress uint
	wrapToAddress   uint

	statusValueUsesRXFIFO      bool
	statusValueComparisonLevel uint
	pullThreshold              uint
	pushThreshold              uint
	outShiftMovesRight         bool
	inShiftMovesRight          bool
	autoPull                   bool
	autoPush                   bool

	sidesetBasePin  uint
	sidesetBitCount uint
	setBasePin      uint
	setPinCount     uint
	inBasePin       uint
	outBasePin      uint
	outPinCount     uint
	jumpPin         uint

	programCounter          uint
	currentInstruction      uint16
	stalled                 bool
	jumped                  bool
	forcedInstructionActive bool
	forcedInstruction       uint16
	execdInstructionActive  bool
	execdInstruction        uint16
	delaysRemaining         uint

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	x                          uint32
	y                          uint32

	clockDivisorInteger    uint16
	clockDivisorFractional uint8

	clockDividerTicksRemaining      uint
	clockDividerFractionAccumulator uint8
}

func NewSM(index uint, memoryReader memory.MemoryReader) *SM {
	return &SM{
		index: index,

		memoryReader: memoryReader,
		fifoRX:       fifo.NewFIFO(4),
		fifoTX:       fifo.NewFIFO(4),

		wrapFromAddress:            31,
		setPinCount:                5,
		outputShiftRegisterCounter: 32,
		outShiftMovesRight:         true,
		inShiftMovesRight:          true,

		clockDivisorInteger: 1,
	}
}

func (sm *SM) SetClockDivisor(divider float32) error {
	divInt, divFrac, err := clockDivisorFromFloat32(divider)
	if err != nil {
		return err
	}
	sm.clockDivisorInteger = divInt
	sm.clockDivisorFractional = divFrac
	return nil
}

func (sm *SM) Tick() error {
	if !sm.stickyOutSetAssertion {
		sm.pinOutputEnables = 0
		sm.pinOutputEnablesMask = 0
		sm.pinOutputs = 0
		sm.pinOutputMask = 0
	}
	sm.pinSidesets = 0
	sm.pinSidesetsMask = 0
	sm.irqWrites = 0
	sm.irqWritesMask = 0

	if sm.forcedInstructionActive {
		sm.currentInstruction = sm.forcedInstruction
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing forced instruction: %w", err)
		}
		sm.forcedInstructionActive = sm.stalled
	} else if sm.clockDividerTicksRemaining == 0 {
		err := sm.dividedTick()
		if err != nil {
			return fmt.Errorf("error performing divided tick: %w", err)
		}
	}

	if sm.clockDividerTicksRemaining == 0 {
		sm.clockDividerTicksRemaining, sm.clockDividerFractionAccumulator = clockDividerUpdateTicksRemaining(sm.clockDividerFractionAccumulator, sm.clockDivisorInteger, sm.clockDivisorFractional)
	}

	sm.clockDividerTicksRemaining--

	return nil
}

func (sm *SM) dividedTick() error {
	if sm.execdInstructionActive {
		sm.currentInstruction = sm.execdInstruction
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing EXEC'd instruction: %w", err)
		}
		sm.execdInstructionActive = sm.stalled
	} else if sm.stalled {
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing stalled instruction: %w", err)
		}
		if !sm.jumped && !sm.stalled {
			sm.incrementProgramCounter()
		}
	} else if sm.delaysRemaining > 0 {
		sm.delaysRemaining--
	} else {
		err := sm.fetch()
		if err != nil {
			return fmt.Errorf("error fetching instruction: %w", err)
		}
		err = sm.execute()
		if err != nil {
			return fmt.Errorf("error executing instruction: %w", err)
		}
		if !sm.jumped && !sm.stalled {
			sm.incrementProgramCounter()
		}
	}
	return nil
}

func (sm *SM) fetch() error {
	instruction, err := sm.memoryReader.Read(sm.programCounter)
	if err != nil {
		return fmt.Errorf("error reading instruction memory: %w", err)
	}
	sm.currentInstruction = instruction
	return nil
}

type instruction uint

const (
	instructionJump     instruction = 0b000
	instructionWait     instruction = 0b001
	instructionIn       instruction = 0b010
	instructionOut      instruction = 0b011
	instructionPushPull instruction = 0b100
	instructionMove     instruction = 0b101
	instructionIRQ      instruction = 0b110
	instructionSet      instruction = 0b111
)

var ErrSMInvalidInstructionType = errors.New("invalid instruction type")

func (sm *SM) execute() error {
	instructionType := instruction((sm.currentInstruction >> 13) & 0b111)
	sm.jumped = false

	switch instructionType {
	case instructionJump:
		condition := jumpCondition((sm.currentInstruction >> 5) & 0b111)
		address := uint(sm.currentInstruction & 0b11111)
		err := sm.executeJump(condition, address)
		if err != nil {
			return fmt.Errorf("error excecuting jump: %w", err)
		}
	case instructionWait:
		polarity := (sm.currentInstruction>>7)&0b1 == 1
		source := waitSource((sm.currentInstruction >> 5) & 0b11)
		index := uint(sm.currentInstruction & 0b11111)
		err := sm.executeWait(polarity, source, index)
		if err != nil {
			return fmt.Errorf("error excecuting wait: %w", err)
		}
	case instructionIn:
		source := inSource((sm.currentInstruction >> 5) & 0b111)
		numberOfBits := uint(sm.currentInstruction & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		err := sm.executeIn(source, numberOfBits)
		if err != nil {
			return fmt.Errorf("error excecuting in: %w", err)
		}
	case instructionOut:
		destination := outDestination((sm.currentInstruction >> 5) & 0b111)
		numberOfBits := uint(sm.currentInstruction & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		err := sm.executeOut(destination, numberOfBits)
		if err != nil {
			return fmt.Errorf("error excecuting out: %w", err)
		}
	case instructionPushPull:
		isPull := (sm.currentInstruction>>7)&0b1 == 1
		ifThreshold := (sm.currentInstruction>>6)&0b1 == 1
		block := (sm.currentInstruction>>5)&0b1 == 1
		err := sm.executePushOrPull(isPull, ifThreshold, block)
		if err != nil {
			return fmt.Errorf("error excecuting push/pull: %w", err)
		}
	case instructionMove:
		destination := moveDestination((sm.currentInstruction >> 5) & 0b111)
		operation := moveOperation((sm.currentInstruction >> 3) & 0b11)
		source := moveSource(sm.currentInstruction & 0b111)
		err := sm.executeMove(destination, operation, source)
		if err != nil {
			return fmt.Errorf("error excecuting move: %w", err)
		}
	case instructionIRQ:
		clearIRQ := (sm.currentInstruction>>6)&0b1 == 1
		waitIRQ := (sm.currentInstruction>>5)&0b1 == 1
		index := uint(sm.currentInstruction & 0b11111)
		err := sm.executeIRQ(clearIRQ, waitIRQ, index)
		if err != nil {
			return fmt.Errorf("error excecuting irq: %w", err)
		}
	case instructionSet:
		destination := setDestination((sm.currentInstruction >> 5) & 0b111)
		data := uint(sm.currentInstruction & 0b11111)
		err := sm.executeSet(destination, data)
		if err != nil {
			return fmt.Errorf("error excecuting set: %w", err)
		}
	default:
		return ErrSMInvalidInstructionType
	}

	if sm.autoPull && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.IsEmpty() {
		osr, err := sm.fifoTX.Read()
		if err != nil {
			return fmt.Errorf("error reading TX FIFO for autopull: %w", err)
		}
		sm.outputShiftRegister = osr
		sm.outputShiftRegisterCounter = 0
	}

	delaySidesetData := (sm.currentInstruction >> 8) & 0b11111
	sm.pinSidesets, sm.pinSidesetsMask, sm.delaysRemaining = delaySidesetUpdate(sm.sidesetIsOptional, sm.sidesetBasePin, sm.sidesetBitCount, delaySidesetData)

	return nil
}

func (sm *SM) incrementProgramCounter() {
	if sm.programCounter == sm.wrapFromAddress {
		sm.programCounter = sm.wrapToAddress
	} else {
		sm.programCounter = (sm.programCounter + 1) % 32
	}
}

type jumpCondition uint

const (
	jumpAlways                jumpCondition = 0b000
	jumpXZero                 jumpCondition = 0b001
	jumpXNonZeroThenDecrement jumpCondition = 0b010
	jumpYZero                 jumpCondition = 0b011
	jumpYNonZeroThenDecrement jumpCondition = 0b100
	jumpXNotEqualY            jumpCondition = 0b101
	jumpPin                   jumpCondition = 0b110
	jumpOSRENotEmpty          jumpCondition = 0b111
)

var ErrSMJumpInvalidCondition = errors.New("invalid condition")

func (sm *SM) executeJump(condition jumpCondition, address uint) error {
	var shouldJump bool
	switch condition {
	case jumpAlways:
		shouldJump = true
	case jumpXZero:
		shouldJump = (sm.x == 0)
	case jumpXNonZeroThenDecrement:
		shouldJump = (sm.x != 0)
		sm.x--
	case jumpYZero:
		shouldJump = (sm.y == 0)
	case jumpYNonZeroThenDecrement:
		shouldJump = (sm.y != 0)
		sm.y--
	case jumpXNotEqualY:
		shouldJump = (sm.x != sm.y)
	case jumpPin:
		shouldJump = (sm.pinInputs>>sm.jumpPin)&0b1 == 1
	case jumpOSRENotEmpty:
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

type waitSource uint

const (
	waitSourceGPIO waitSource = 0b00
	waitSourcePin  waitSource = 0b01
	waitSourceIRQ  waitSource = 0b10
)

var ErrSMWaitInvalidSource = errors.New("invalid source")

func (sm *SM) executeWait(polarity bool, source waitSource, index uint) error {
	switch source {
	case waitSourceGPIO:
		sm.stalled = ((sm.pinInputs>>index)&0b1 == 1) == polarity
	case waitSourcePin:
		pin := (sm.inBasePin + index) % 32
		sm.stalled = ((sm.pinInputs>>pin)&0b1 == 1) == polarity
	case waitSourceIRQ:
		relative := (index>>4)&0b1 == 1
		irq := index
		if relative {
			upperBit := irq & 0b100
			lowerBits := (irq + sm.index) & 0b011
			irq = upperBit | lowerBits
		}
		sm.stalled = ((sm.irqInputs>>irq)&0b1 == 1) == polarity
		if polarity == true && !sm.stalled {
			var clearIRQMask uint8 = (0b1 << irq)
			sm.irqWrites &= ^clearIRQMask
			sm.irqWritesMask |= clearIRQMask
		}
	default:
		return ErrSMWaitInvalidSource
	}
	return nil
}

type inSource uint

const (
	inSourcePins inSource = 0b000
	inSourceX    inSource = 0b001
	inSourceY    inSource = 0b010
	inSourceNull inSource = 0b011
	inSourceISR  inSource = 0b110
	inSourceOSR  inSource = 0b111
)

var ErrSMInInvalidSource = errors.New("invalid source")

func (sm *SM) executeIn(source inSource, count uint) error {
	if !sm.stalled {
		var data uint32
		var mask uint32 = (0b1 << count) - 1
		switch source {
		case inSourcePins:
			pinMask := bits.RotateLeft32(mask, -int(sm.inBasePin))
			data = bits.RotateLeft32(sm.pinInputs&pinMask, -int(sm.inBasePin))
		case inSourceX:
			data = (sm.x & mask)
		case inSourceY:
			data = (sm.y & mask)
		case inSourceNull:
			data = 0
		case inSourceISR:
			data = (sm.inputShiftRegister & mask)
		case inSourceOSR:
			data = (sm.outputShiftRegister & mask)
		default:
			return ErrSMInInvalidSource
		}
		if sm.inShiftMovesRight {
			sm.inputShiftRegister >>= count
			sm.inputShiftRegister |= (data << (32 - count))
		} else {
			sm.inputShiftRegister <<= count
			sm.inputShiftRegister |= data
		}
		sm.inputShiftRegisterCounter += count
	}
	sm.stalled = false
	if sm.autoPush && sm.inputShiftRegisterCounter >= sm.pushThreshold {
		sm.stalled = sm.fifoRX.IsFull()
		if !sm.stalled {
			err := sm.fifoRX.Write(sm.inputShiftRegister)
			if err != nil {
				return fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			sm.inputShiftRegister = 0
			sm.inputShiftRegisterCounter = 0
		}
	}
	return nil
}

type outDestination uint

const (
	outDestinationPins               = 0b000
	outDestinationX                  = 0b001
	outDestinationY                  = 0b010
	outDestinationNull               = 0b011
	outDestinationPinDirections      = 0b100
	outDestinationProgramCounter     = 0b101
	outDestinationInputShiftRegister = 0b110
	outDestinationEXEC               = 0b111
)

var ErrSMOutInvalidDestination = errors.New("invalid out destination")

func (sm *SM) executeOut(destination outDestination, count uint) error {
	alreadyStalled := sm.stalled
	sm.stalled = false
	if sm.autoPull && sm.outputShiftRegisterCounter >= sm.pullThreshold {
		if !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return fmt.Errorf("error reading TX fifo for autopull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		}
		if !alreadyStalled {
			sm.stalled = true
		} else if !sm.fifoTX.IsEmpty() {
			sm.stalled = false
		}
		return nil
	}

	var data uint32
	var mask uint32 = (0b1 << count) - 1
	if sm.outShiftMovesRight {
		data = sm.outputShiftRegister & mask
		sm.outputShiftRegister >>= count
	} else {
		data = (sm.outputShiftRegister >> (32 - count))
		sm.outputShiftRegister <<= count
	}
	sm.outputShiftRegisterCounter += count

	writePins := true
	if sm.inlineOutWriteEnableIsUsed {
		writePins = ((data >> sm.inlineOutWriteEnableBitIndex) & 0b1) == 1
	}

	var pinData uint32 = bits.RotateLeft32(data, -int(sm.outBasePin))
	var pinMask uint32 = bits.RotateLeft32((0b1<<sm.outPinCount)-1, -int(sm.outBasePin))

	switch destination {
	case outDestinationPins:
		if writePins {
			sm.pinOutputs = pinData
			sm.pinOutputMask = pinMask
		}
	case outDestinationX:
		sm.x = data
	case outDestinationY:
		sm.y = data
	case outDestinationNull:
		// discards data
	case outDestinationPinDirections:
		if writePins {
			sm.pinOutputEnables = pinData
			sm.pinOutputEnablesMask = pinMask
		}
	case outDestinationProgramCounter:
		sm.programCounter = uint(data % 32)
		sm.jumped = true
	case outDestinationInputShiftRegister:
		sm.inputShiftRegister = data
		sm.inputShiftRegisterCounter = count
	case outDestinationEXEC:
		sm.execdInstruction = uint16(data)
		sm.execdInstructionActive = true
	default:
		return ErrSMOutInvalidDestination
	}

	return nil
}

func (sm *SM) executePushOrPull(isPull bool, ifThreshold bool, block bool) error {
	if isPull {
		shouldPull := (!ifThreshold && !sm.autoPull) || (sm.outputShiftRegisterCounter >= sm.pullThreshold)
		sm.stalled = block && shouldPull && sm.fifoTX.IsEmpty()
		if shouldPull && !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return fmt.Errorf("error reading TX fifo for pull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
			sm.stalled = false
		} else if shouldPull && !block {
			sm.outputShiftRegister = sm.x
			sm.outputShiftRegisterCounter = 0
		}
	} else {
		shouldPush := !ifThreshold || (sm.inputShiftRegisterCounter >= sm.pushThreshold)
		sm.stalled = block && shouldPush && sm.fifoRX.IsFull()
		if shouldPush && !sm.fifoRX.IsFull() {
			err := sm.fifoRX.Write(sm.inputShiftRegister)
			if err != nil {
				return fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			sm.inputShiftRegister = 0
			sm.inputShiftRegisterCounter = 0
		}
	}
	return nil
}

type moveDestination uint
type moveOperation uint
type moveSource uint

const (
	moveDestinationPins moveDestination = 0b000
	moveDestinationX    moveDestination = 0b001
	moveDestinationY    moveDestination = 0b010
	moveDestinationEXEC moveDestination = 0b100
	moveDestinationPC   moveDestination = 0b101
	moveDestinationISR  moveDestination = 0b110
	moveDestinationOSR  moveDestination = 0b111

	moveOperationNone       moveOperation = 0b00
	moveOperationInvert     moveOperation = 0b01
	moveOperationBitReverse moveOperation = 0b10

	moveSourcePins   moveSource = 0b000
	moveSourceX      moveSource = 0b001
	moveSourceY      moveSource = 0b010
	moveSourceNull   moveSource = 0b011
	moveSourceStatus moveSource = 0b101
	moveSourceISR    moveSource = 0b110
	moveSourceOSR    moveSource = 0b111
)

var ErrSMInvalidMoveSource = errors.New("invalid move source")
var ErrSMMoveSourceOSRDisallowedWhenAutopullEnabled = errors.New("cannot use OSR as source whilst autopull is enabled")
var ErrSMInvalidMoveOperation = errors.New("invalid move operation")
var ErrSMInvalidMoveDestination = errors.New("invalid move destination")

func (sm *SM) executeMove(destination moveDestination, operation moveOperation, source moveSource) error {
	var sourceData uint32
	switch source {
	case moveSourcePins:
		sourceData = bits.RotateLeft32(sm.pinInputs, -int(sm.inBasePin))
	case moveSourceX:
		sourceData = sm.x
	case moveSourceY:
		sourceData = sm.y
	case moveSourceNull:
		sourceData = 0
	case moveSourceStatus:
		var fifoLevel uint
		if sm.statusValueUsesRXFIFO {
			fifoLevel = sm.fifoRX.Level()
		} else {
			fifoLevel = sm.fifoTX.Level()
		}
		if fifoLevel < sm.statusValueComparisonLevel {
			sourceData = 0xFFFFFFFF
		} else {
			sourceData = 0x00000000
		}
	case moveSourceISR:
		sourceData = sm.inputShiftRegister
	case moveSourceOSR:
		if sm.autoPull {
			return ErrSMMoveSourceOSRDisallowedWhenAutopullEnabled
		} else {
			sourceData = sm.outputShiftRegister
		}
	default:
		return ErrSMInvalidMoveSource
	}

	var modifiedData uint32
	switch operation {
	case moveOperationNone:
		modifiedData = sourceData
	case moveOperationInvert:
		modifiedData = sourceData ^ 0xFFFFFFFF
	case moveOperationBitReverse:
		modifiedData = bits.Reverse32(sourceData)
	default:
		return ErrSMInvalidMoveOperation
	}

	writePins := true
	if sm.inlineOutWriteEnableIsUsed {
		writePins = ((modifiedData >> sm.inlineOutWriteEnableBitIndex) & 0b1) == 1
	}

	switch destination {
	case moveDestinationPins:
		if writePins {
			sm.pinOutputs = bits.RotateLeft32(modifiedData, -int(sm.outBasePin))
			sm.pinOutputMask = 0xFFFFFFFF
		}
	case moveDestinationX:
		sm.x = modifiedData
	case moveDestinationY:
		sm.y = modifiedData
	case moveDestinationEXEC:
		sm.execdInstruction = uint16(modifiedData)
		sm.execdInstructionActive = true
	case moveDestinationPC:
		sm.programCounter = uint(modifiedData % 32)
	case moveDestinationISR:
		sm.inputShiftRegister = modifiedData
		sm.inputShiftRegisterCounter = 0
	case moveDestinationOSR:
		sm.outputShiftRegister = modifiedData
		sm.outputShiftRegisterCounter = 0
	default:
		return ErrSMInvalidMoveDestination
	}

	return nil
}

var ErrSMIRQClearWaitDisallowed = errors.New("cannot clear and wait at the same time")

func (sm *SM) executeIRQ(clearIRQ bool, waitIRQ bool, index uint) error {
	relative := (index>>4)&0b1 == 1
	irq := index
	if relative {
		upperBit := irq & 0b100
		lowerBits := (irq + sm.index) & 0b011
		irq = upperBit | lowerBits
	}

	if sm.stalled {
		sm.stalled = ((sm.irqInputs >> irq) & 0b1) == 1
		return nil
	}

	var irqMask uint8 = (0b1 << irq)
	if clearIRQ {
		if waitIRQ {
			return ErrSMIRQClearWaitDisallowed
		}
		sm.irqWrites &= ^irqMask
		sm.irqWritesMask |= irqMask
	} else {
		sm.irqWrites |= irqMask
		sm.irqWritesMask |= irqMask
		if waitIRQ {
			sm.stalled = true
		}
	}

	return nil
}

type setDestination uint

const (
	setDestinationPins          setDestination = 0b000
	setDestinationX             setDestination = 0b001
	setDestinationY             setDestination = 0b010
	setDestinationPinDirections setDestination = 0b100
)

var ErrSMSetInvalidDestination = errors.New("invalid destination")

func (sm *SM) executeSet(destination setDestination, data uint) error {
	var pinData uint32 = bits.RotateLeft32(uint32(data), -int(sm.setBasePin))
	var pinMask uint32 = bits.RotateLeft32((0b1<<sm.setPinCount)-1, -int(sm.setBasePin))
	switch destination {
	case setDestinationPins:
		sm.pinOutputs = pinData
		sm.pinOutputMask = pinMask
	case setDestinationX:
		sm.x = uint32(data)
	case setDestinationY:
		sm.y = uint32(data)
	case setDestinationPinDirections:
		sm.pinOutputEnables = pinData
		sm.pinOutputEnablesMask = pinMask
	default:
		return ErrSMSetInvalidDestination
	}
	return nil
}
