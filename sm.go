package pio

import (
	"errors"
	"fmt"
	"math/bits"
)

type sm struct {
	id uint

	im        *im
	fifoRX    *fifo
	fifoTX    *fifo
	pinInputs uint32
	irqInputs uint8

	enabled bool

	pinOutputs           uint32
	pinOutputMask        uint32
	pinOutputEnables     uint32
	pinOutputEnablesMask uint32
	pinSidesets          uint32
	pinSidesetsMask      uint32
	irqWrites            uint8
	irqWritesMask        uint8

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
	// If the level of the selected FIFO is < statusComparisonLevel, 'MOV x, STATUS' returns all '1's.
	statusSelectionUsesRX bool
	// Choose a level which the selected FIFO's level will be compared to.
	statusComparisonLevel uint
	// Number of bits shifted out of the outputShiftRegister before an autopull (or pull ifempty) will happen.
	// Can be a value of 1-32.
	pullThreshold uint
	// Number of bits shifted into the inputShiftRegister before an autopush (or push iffull) will happen.
	// Can be a value of 1-32.
	pushThreshold uint
	// Direction the outputShiftRegister is shifted in as data is taken out.
	outShiftDirectionRight bool
	// Direction the inputShiftRegister is shifted in as data is put in.
	inShiftDirectionRight bool
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

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	x                          uint32
	y                          uint32

	// The integer component of the clock divisor.
	clockDivisorInteger uint16
	// The fractional component of the clock divisor, as a number of 256ths to be added to the integer component.
	clockDivisorFractional uint8

	// A computed number of system clock ticks remaining until the next divided tick.
	clockDividerTicksRemaining uint
	// An accumulated count of 256ths used to approximate a fractional divide of the system clock.
	// Every time this accumulator wraps, the next divided tick is delayed by +1 clock cycle.
	clockDividerFractionAccumulator uint8
}

func newSM(id uint, im *im) *sm {
	return &sm{
		id: id,

		im:     im,
		fifoRX: newFIFO(4),
		fifoTX: newFIFO(4),

		wrapFromAddress:            31,
		setCount:                   5,
		outputShiftRegisterCounter: 32,
		outShiftDirectionRight:     true,
		inShiftDirectionRight:      true,

		clockDivisorInteger: 1,
	}
}

var ErrInvalidClockDivider = errors.New("invalid clock divider")

func (sm *sm) setClockDivider(divider float32) error {
	if divider < 1 || divider > 65536 {
		return ErrInvalidClockDivider
	}
	if divider == 65536 {
		sm.clockDivisorInteger = 0
		sm.clockDivisorFractional = 0
	} else {
		sm.clockDivisorInteger = uint16(divider)
		sm.clockDivisorFractional = uint8((divider - float32(sm.clockDivisorInteger)) * 256)
	}
	return nil
}

func (sm *sm) tick() error {
	sm.pinSidesets = 0
	sm.pinSidesetsMask = 0
	if !sm.stickyOutSetAssertion {
		sm.pinOutputs = 0
		sm.pinOutputMask = 0
		sm.pinOutputEnables = 0
		sm.pinOutputEnablesMask = 0
	}
	sm.irqWrites = 0
	sm.irqWritesMask = 0

	if sm.newForcedInstruction {
		sm.currentInstruction = sm.forcedInstruction
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing forced instruction: %w", err)
		}
		sm.newForcedInstruction = sm.stalled
	} else if sm.clockDividerTicksRemaining == 0 {
		sm.dividedTick()
	}

	if sm.clockDividerTicksRemaining == 0 {
		actualIntegerDivisor := uint(sm.clockDivisorInteger)
		// clockDivisorInteger as 0 actually represents a divisor of 65536, as dividing by 0 is not a useful operation.
		if actualIntegerDivisor == 0 {
			actualIntegerDivisor = 65536
		}

		// Setup the amount of ticks to count until the next divided tick based on the integer divisor.
		sm.clockDividerTicksRemaining = actualIntegerDivisor

		// Now, use the fractionAccumulator to determine whether to delay the next divided clock tick by an additional clock cycle.
		// Importantly, clockDividerFractionAccumulator is a uint8, so automatically wraps around every 256.
		// As the clockDivisorFractional represents a number of 256ths, this makes calculations very easy.
		previousAccumulatorValue := sm.clockDividerFractionAccumulator
		sm.clockDividerFractionAccumulator += sm.clockDivisorFractional
		// If adding the clockDivisorFractional has caused the fraction accumulator to wrap,
		// delay the next divided clock tick by an additional clock cycle.
		if sm.clockDividerFractionAccumulator < previousAccumulatorValue {
			sm.clockDividerTicksRemaining++
		}
	}

	// Count down a cycle.
	sm.clockDividerTicksRemaining--

	return nil
}

func (sm *sm) dividedTick() error {
	if sm.newEXECdInstruction {
		sm.currentInstruction = sm.execdInstruction
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing EXEC'd instruction: %w", err)
		}
		sm.newEXECdInstruction = sm.stalled
	} else if sm.stalled {
		err := sm.execute()
		if err != nil {
			return fmt.Errorf("error executing stalled instruction: %w", err)
		}
		if !sm.jumped && !sm.stalled {
			sm.incrementProgramCounter()
		}
	} else if sm.delays > 0 {
		sm.delays--
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

func (sm *sm) fetch() error {
	instruction, err := sm.im.read(sm.programCounter)
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

func (sm *sm) execute() error {
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

		// remember, MOV PINS also affected by inlineOutWriteEnable
		// the modified value of the MOV is used to determine inlineOutWriteEnable

		/* err := sm.ExecuteMove()*/
		/*if err != nil {*/
		/*return fmt.Errorf("error excecuting move: %w", err)*/
		/*}*/
	case instructionIRQ:
	/*   err := sm.ExecuteIRQ()*/
	/*if err != nil {*/
	/*return fmt.Errorf("error excecuting irq: %w", err)*/
	/*}*/
	case instructionSet:
	/*   err := sm.ExecuteSet()*/
	/*if err != nil {*/
	/*return fmt.Errorf("error excecuting set: %w", err)*/
	/*}*/
	default:
		return ErrSMInvalidInstructionType
	}

	if sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.isEmpty() {
		osr, err := sm.fifoTX.read()
		if err != nil {
			return fmt.Errorf("error reading TX FIFO for autopull: %w", err)
		}
		sm.outputShiftRegister = osr
		sm.outputShiftRegisterCounter = 0
	}

	delaySidesetData := (sm.currentInstruction >> 8) & 0b11111
	doSideset := (sm.sidesetBitCount > 0) && (!sm.sidesetIsOptional || ((sm.currentInstruction>>12)&0b1 == 1))
	doDelay := (sm.sidesetBitCount < 5)
	var delayMask uint16 = (0b1 << (5 - sm.sidesetBitCount)) - 1
	var sidesetMask = ^delayMask
	if sm.sidesetIsOptional {
		sidesetMask &= 0b01111
	}
	if doSideset {
		sidesetData := (delaySidesetData & sidesetMask) >> (5 - sm.sidesetBitCount)
		pinCount := sm.sidesetBitCount
		if sm.sidesetIsOptional {
			pinCount -= 1
		}
		var pinData uint32 = bits.RotateLeft32(uint32(sidesetData), int(sm.sidesetBase))
		var pinMask uint32 = bits.RotateLeft32((0b1<<pinCount)-1, int(sm.sidesetBase))
		sm.pinSidesets = pinData
		sm.pinSidesetsMask = pinMask
	}
	if doDelay {
		delayData := (delaySidesetData & delayMask)
		sm.delays = uint(delayData)
	}

	return nil
}

func (sm *sm) incrementProgramCounter() {
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

func (sm *sm) executeJump(condition jumpCondition, address uint) error {
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

func (sm *sm) executeWait(polarity bool, source waitSource, index uint) error {
	switch source {
	case waitSourceGPIO:
		sm.stalled = ((sm.pinInputs>>index)&0b1 == 1) == polarity
	case waitSourcePin:
		pin := (sm.inBase + index) % 32
		sm.stalled = ((sm.pinInputs>>pin)&0b1 == 1) == polarity
	case waitSourceIRQ:
		relative := (index>>4)&0b1 == 1
		irq := index
		if relative {
			upperBit := irq & 0b100
			lowerBits := (irq + sm.id) & 0b011
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

func (sm *sm) executeIn(source inSource, count uint) error {
	if !sm.stalled {
		var data uint32
		var mask uint32 = (0b1 << count) - 1
		switch source {
		case inSourcePins:
			pinMask := bits.RotateLeft32(mask, int(sm.inBase))
			data = bits.RotateLeft32(sm.pinInputs&pinMask, -int(sm.inBase))
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
		if sm.inShiftDirectionRight {
			sm.inputShiftRegister >>= count
			sm.inputShiftRegister |= (data << (32 - count))
		} else {
			sm.inputShiftRegister <<= count
			sm.inputShiftRegister |= data
		}
		sm.inputShiftRegisterCounter += count
	}
	sm.stalled = false
	if sm.autopush && sm.inputShiftRegisterCounter >= sm.pushThreshold {
		sm.stalled = sm.fifoRX.isFull()
		if !sm.stalled {
			err := sm.fifoRX.write(sm.inputShiftRegister)
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

func (sm *sm) executeOut(destination outDestination, count uint) error {
	alreadyStalled := sm.stalled
	sm.stalled = false
	if sm.autopull && sm.outputShiftRegisterCounter >= sm.pullThreshold {
		if !sm.fifoTX.isEmpty() {
			osr, err := sm.fifoTX.read()
			if err != nil {
				return fmt.Errorf("error reading TX fifo for autopull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		}
		if !alreadyStalled {
			sm.stalled = true
		} else if !sm.fifoTX.isEmpty() {
			sm.stalled = false
		}
		return nil
	}

	var data uint32
	var mask uint32 = (0b1 << count) - 1
	if sm.outShiftDirectionRight {
		data = sm.outputShiftRegister & mask
		sm.outputShiftRegister >>= count
	} else {
		data = (sm.outputShiftRegister >> (32 - count))
		sm.outputShiftRegister <<= count
	}
	sm.outputShiftRegisterCounter += count

	writePins := true
	if sm.inlineOutWriteEnableIsUsed {
		writePins = ((data >> sm.inlineOutWriteEnableBit) & 0b1) == 1
	}

	var pinData uint32 = bits.RotateLeft32(data, int(sm.outBase))
	var pinMask uint32 = bits.RotateLeft32((0b1<<sm.outCount)-1, int(sm.outBase))

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
		sm.newEXECdInstruction = true
	default:
		return ErrSMOutInvalidDestination
	}

	return nil
}

func (sm *sm) executePushOrPull(isPull bool, ifThreshold bool, block bool) error {
	if isPull {
		shouldPull := (!ifThreshold && !sm.autopull) || (sm.outputShiftRegisterCounter >= sm.pullThreshold)
		sm.stalled = block && shouldPull && sm.fifoTX.isEmpty()
		if shouldPull && !sm.fifoTX.isEmpty() {
			osr, err := sm.fifoTX.read()
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
		sm.stalled = block && shouldPush && sm.fifoRX.isFull()
		if shouldPush && !sm.fifoRX.isFull() {
			err := sm.fifoRX.write(sm.inputShiftRegister)
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

func (sm *sm) executeMove(destination moveDestination, operation moveOperation, source moveSource) error {
	return nil
}
