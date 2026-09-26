package sm

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/headblockhead/pio/internal/fifo"
	"github.com/headblockhead/pio/internal/memory"
)

type Observer interface {
	Index() uint

	FIFORXObserver() fifo.Observer
	FIFOTXObserver() fifo.Observer

	Enabled() bool

	PinOutputEnables() uint32
	PinOutputEnablesMask() uint32
	PinOutputs() uint32
	PinOutputsMask() uint32
	PinSidesets() uint32
	PinSidesetsMask() uint32
	IRQWrites() uint8
	IRQWritesMask() uint8

	SidesetIsOptional() bool
	SidesetControlsPinDirection() bool
	OutWriteEnableUsed() bool
	OutWriteEnableBitIndex() uint
	StickyOutSetAssertionEnabled() bool

	WrapFromAddress() uint
	WrapToAddress() uint

	StatusValueUsesRXFIFO() bool
	StatusValueComparisonLevel() uint
	PullThreshold() uint
	PushThreshold() uint
	OutShiftMovesRight() bool
	InShiftMovesRight() bool
	AutopullEnabled() bool
	AutopushEnabled() bool

	BaseSidesetPin() uint
	BitCountSideset() uint
	BaseSetPin() uint
	PinCountSet() uint
	BaseInPin() uint
	BaseOutPin() uint
	PinCountOut() uint
	JumpPin() uint

	ProgramCounter() uint
	Stalled() bool
	StalledIRQ() bool
	Jumped() bool
	DelaysRemaining() uint
	NewForcedInstruction() bool
	ForcedInstruction() uint16
	ForcedInstructionStalled() bool
	NewEXECdInstruction() bool
	EXECdInstructionStalled() bool
	LatchedInstruction() uint16

	OutputShiftRegister() uint32
	OutputShiftRegisterCounter() uint
	InputShiftRegister() uint32
	InputShiftRegisterCounter() uint
	XRegister() uint32
	XRegisterInitialized() bool
	YRegister() uint32
	YRegisterInitialized() bool

	ClockDivisor() float32
	ClockDivisorInteger() uint
	ClockDivisorFractional() uint8

	ClockDividerTicksRemaining() uint
	ClockDividerFractionAccumulator() uint8
}

type Configurator interface {
	Observer

	Restart()
	SetEnabled(bool)
	SetSidesetIsOptional(bool)
	SetSidesetControlsPinDirection(bool)
	SetOutWriteEnableUsed(bool)
	SetOutWriteEnableBitIndex(uint) error
	SetStickyOutSetAssertionEnabled(bool)

	SetWrapFromAddress(uint) error
	SetWrapToAddress(uint) error

	SetStatusValueUsesRXFIFO(bool)
	SetStatusValueComparisonLevel(uint) error
	SetPullThreshold(uint) error
	SetPushThreshold(uint) error
	SetOutShiftMovesRight(bool)
	SetInShiftMovesRight(bool)
	SetAutopullEnabled(bool)
	SetAutopushEnabled(bool)

	SetBaseSidesetPin(uint) error
	SetBitCountSideset(uint) error
	SetBaseSetPin(uint) error
	SetPinCountSet(uint) error
	SetBaseInPin(uint) error
	SetBaseOutPin(uint) error
	SetPinCountOut(uint) error
	SetJumpPin(uint) error

	SetFIFOJoin(FIFOJoin) error

	ForceInstruction(uint16)

	RestartClockDivider()
	SetClockDivisor(float32) error
	SetClockDivisorInteger(uint) error
	SetClockDivisorFractional(uint8)
}

type Controller interface {
	SetPinInputs(uint32)
	SetIRQInputs(uint8)
	PinOutputEnables() uint32
	PinOutputEnablesMask() uint32
	PinOutputs() uint32
	PinOutputsMask() uint32
	PinSidesets() uint32
	PinSidesetsMask() uint32
	IRQWrites() uint8
	IRQWritesMask() uint8

	Tick() error
}

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
	pinOutputsMask       uint32
	pinSidesets          uint32
	pinSidesetsMask      uint32
	irqWrites            uint8
	irqWritesMask        uint8

	sidesetIsOptional            bool
	sidesetControlsPinDirection  bool
	outWriteEnableUsed           bool
	outWriteEnableBitIndex       uint
	stickyOutSetAssertionEnabled bool

	wrapFromAddress uint
	wrapToAddress   uint

	statusValueUsesRXFIFO      bool
	statusValueComparisonLevel uint
	pullThreshold              uint
	pushThreshold              uint
	outShiftMovesRight         bool
	inShiftMovesRight          bool
	autopullEnabled            bool
	autopushEnabled            bool

	baseSidesetPin  uint
	bitCountSideset uint
	baseSetPin      uint
	pinCountSet     uint
	baseInPin       uint
	baseOutPin      uint
	pinCountOut     uint
	jumpPin         uint

	programCounter           uint
	stalled                  bool
	stalledIRQ               bool
	jumped                   bool
	delaysRemaining          uint
	newForcedInstruction     bool
	forcedInstruction        uint16
	forcedInstructionStalled bool
	newEXECdInstruction      bool
	execdInstructionStalled  bool
	latchedInstruction       uint16

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint
	xRegister                  uint32
	xRegisterInitialized       bool
	yRegister                  uint32
	yRegisterInitialized       bool

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
		pullThreshold:              32,
		pushThreshold:              32,
		pinCountSet:                5,
		outputShiftRegisterCounter: 32,
		outShiftMovesRight:         true,
		inShiftMovesRight:          true,

		clockDivisorInteger: 1,
	}
}

func (sm *SM) Observer() Observer {
	return sm
}

func (sm *SM) Index() uint { return sm.index }

func (sm *SM) FIFORXObserver() fifo.Observer { return sm.fifoRX.Observer() }
func (sm *SM) FIFOTXObserver() fifo.Observer { return sm.fifoTX.Observer() }

func (sm *SM) Enabled() bool { return sm.enabled }

func (sm *SM) PinOutputEnables() uint32     { return sm.pinOutputEnables }
func (sm *SM) PinOutputEnablesMask() uint32 { return sm.pinOutputEnablesMask }
func (sm *SM) PinOutputs() uint32           { return sm.pinOutputs }
func (sm *SM) PinOutputsMask() uint32       { return sm.pinOutputsMask }
func (sm *SM) PinSidesets() uint32          { return sm.pinSidesets }
func (sm *SM) PinSidesetsMask() uint32      { return sm.pinSidesetsMask }
func (sm *SM) IRQWrites() uint8             { return sm.irqWrites }
func (sm *SM) IRQWritesMask() uint8         { return sm.irqWritesMask }

func (sm *SM) SidesetIsOptional() bool            { return sm.sidesetIsOptional }
func (sm *SM) SidesetControlsPinDirection() bool  { return sm.sidesetControlsPinDirection }
func (sm *SM) OutWriteEnableUsed() bool           { return sm.outWriteEnableUsed }
func (sm *SM) OutWriteEnableBitIndex() uint       { return sm.outWriteEnableBitIndex }
func (sm *SM) StickyOutSetAssertionEnabled() bool { return sm.stickyOutSetAssertionEnabled }

func (sm *SM) WrapFromAddress() uint { return sm.wrapFromAddress }
func (sm *SM) WrapToAddress() uint   { return sm.wrapToAddress }

func (sm *SM) StatusValueUsesRXFIFO() bool      { return sm.statusValueUsesRXFIFO }
func (sm *SM) StatusValueComparisonLevel() uint { return sm.statusValueComparisonLevel }
func (sm *SM) PullThreshold() uint              { return sm.pullThreshold }
func (sm *SM) PushThreshold() uint              { return sm.pushThreshold }
func (sm *SM) OutShiftMovesRight() bool         { return sm.outShiftMovesRight }
func (sm *SM) InShiftMovesRight() bool          { return sm.inShiftMovesRight }
func (sm *SM) AutopullEnabled() bool            { return sm.autopullEnabled }
func (sm *SM) AutopushEnabled() bool            { return sm.autopushEnabled }

func (sm *SM) BaseSidesetPin() uint  { return sm.baseSidesetPin }
func (sm *SM) BitCountSideset() uint { return sm.bitCountSideset }
func (sm *SM) BaseSetPin() uint      { return sm.baseSetPin }
func (sm *SM) PinCountSet() uint     { return sm.pinCountSet }
func (sm *SM) BaseInPin() uint       { return sm.baseInPin }
func (sm *SM) BaseOutPin() uint      { return sm.baseOutPin }
func (sm *SM) PinCountOut() uint     { return sm.pinCountOut }
func (sm *SM) JumpPin() uint         { return sm.jumpPin }

func (sm *SM) ProgramCounter() uint           { return sm.programCounter }
func (sm *SM) Stalled() bool                  { return sm.stalled }
func (sm *SM) StalledIRQ() bool               { return sm.stalledIRQ }
func (sm *SM) Jumped() bool                   { return sm.jumped }
func (sm *SM) DelaysRemaining() uint          { return sm.delaysRemaining }
func (sm *SM) NewForcedInstruction() bool     { return sm.newForcedInstruction }
func (sm *SM) ForcedInstruction() uint16      { return sm.forcedInstruction }
func (sm *SM) ForcedInstructionStalled() bool { return sm.forcedInstructionStalled }
func (sm *SM) NewEXECdInstruction() bool      { return sm.newEXECdInstruction }
func (sm *SM) EXECdInstructionStalled() bool  { return sm.execdInstructionStalled }
func (sm *SM) LatchedInstruction() uint16     { return sm.latchedInstruction }

func (sm *SM) OutputShiftRegister() uint32      { return sm.outputShiftRegister }
func (sm *SM) OutputShiftRegisterCounter() uint { return sm.outputShiftRegisterCounter }
func (sm *SM) InputShiftRegister() uint32       { return sm.inputShiftRegister }
func (sm *SM) InputShiftRegisterCounter() uint  { return sm.inputShiftRegisterCounter }
func (sm *SM) XRegister() uint32                { return sm.xRegister }
func (sm *SM) XRegisterInitialized() bool       { return sm.xRegisterInitialized }
func (sm *SM) YRegister() uint32                { return sm.yRegister }
func (sm *SM) YRegisterInitialized() bool       { return sm.yRegisterInitialized }

func (sm *SM) ClockDivisor() float32 {
	return clockDivisorToFloat32(sm.clockDivisorInteger, sm.clockDivisorFractional)
}
func (sm *SM) ClockDivisorInteger() uint {
	if sm.clockDivisorInteger == 0 {
		return 65536
	} else {
		return uint(sm.clockDivisorInteger)
	}
}
func (sm *SM) ClockDivisorFractional() uint8          { return sm.clockDivisorFractional }
func (sm *SM) ClockDividerTicksRemaining() uint       { return sm.clockDividerTicksRemaining }
func (sm *SM) ClockDividerFractionAccumulator() uint8 { return sm.clockDividerFractionAccumulator }

func (sm *SM) Configurator() Configurator {
	return sm
}

func (sm *SM) Restart() {
	sm.inputShiftRegisterCounter = 0
	sm.outputShiftRegisterCounter = 32
	sm.inputShiftRegister = 0
	sm.delaysRemaining = 0
	sm.stalledIRQ = false
	sm.newForcedInstruction = false
	sm.forcedInstructionStalled = false
	sm.newEXECdInstruction = false
	sm.execdInstructionStalled = false
	sm.latchedInstruction = 0
	sm.pinOutputEnables = 0
	sm.pinOutputEnablesMask = 0
	sm.pinOutputs = 0
	sm.pinOutputsMask = 0
}
func (sm *SM) SetEnabled(enabled bool) {
	sm.enabled = enabled
}
func (sm *SM) SetSidesetIsOptional(sidesetIsOptional bool) {
	sm.sidesetIsOptional = sidesetIsOptional
}
func (sm *SM) SetSidesetControlsPinDirection(sidesetControlsPinDirection bool) {
	sm.sidesetControlsPinDirection = sidesetControlsPinDirection
}
func (sm *SM) SetOutWriteEnableUsed(outWriteEnableUsed bool) {
	sm.outWriteEnableUsed = outWriteEnableUsed
}

var ErrSMSetOutWriteEnableBitIndexOutOfRange = errors.New("out write enable bit index must be less than 32")

func (sm *SM) SetOutWriteEnableBitIndex(outWriteEnableBitIndex uint) error {
	if outWriteEnableBitIndex > 31 {
		return ErrSMSetOutWriteEnableBitIndexOutOfRange
	}
	sm.outWriteEnableBitIndex = outWriteEnableBitIndex
	return nil
}
func (sm *SM) SetStickyOutSetAssertionEnabled(stickyOutSetAssertionEnabled bool) {
	sm.stickyOutSetAssertionEnabled = stickyOutSetAssertionEnabled
}

var ErrSMSetWrapFromAddressOutOfRange = errors.New("wrap from address must be less than 32")

func (sm *SM) SetWrapFromAddress(wrapFromAddress uint) error {
	if wrapFromAddress > 31 {
		return ErrSMSetWrapFromAddressOutOfRange
	}
	sm.wrapFromAddress = wrapFromAddress
	return nil
}

var ErrSMSetWrapToAddressOutOfRange = errors.New("wrap to address must be less than 32")

func (sm *SM) SetWrapToAddress(wrapToAddress uint) error {
	if wrapToAddress > 31 {
		return ErrSMSetWrapToAddressOutOfRange
	}
	sm.wrapToAddress = wrapToAddress
	return nil
}

func (sm *SM) SetStatusValueUsesRXFIFO(statusValueUsesRXFIFO bool) {
	sm.statusValueUsesRXFIFO = statusValueUsesRXFIFO
}

var ErrSMSetStatusValueComparisonLevelOutOfRange = errors.New("status value comparison level must be less than 16")

func (sm *SM) SetStatusValueComparisonLevel(statusValueComparisonLevel uint) error {
	if statusValueComparisonLevel > 15 {
		return ErrSMSetStatusValueComparisonLevelOutOfRange
	}
	sm.statusValueComparisonLevel = statusValueComparisonLevel
	return nil
}

var ErrSMSetPullThresholdOutOfRange = errors.New("pull threshold must be in the range 1 to 32")

func (sm *SM) SetPullThreshold(pullThreshold uint) error {
	if pullThreshold > 32 || pullThreshold < 1 {
		return ErrSMSetPullThresholdOutOfRange
	}
	sm.pullThreshold = pullThreshold
	return nil
}

var ErrSMSetPushThresholdOutOfRange = errors.New("push threshold must be in the range 1 to 32")

func (sm *SM) SetPushThreshold(pushThreshold uint) error {
	if pushThreshold > 32 || pushThreshold < 1 {
		return ErrSMSetPushThresholdOutOfRange
	}
	sm.pushThreshold = pushThreshold
	return nil
}
func (sm *SM) SetOutShiftMovesRight(outShiftMovesRight bool) {
	sm.outShiftMovesRight = outShiftMovesRight
}
func (sm *SM) SetInShiftMovesRight(inShiftMovesRight bool) {
	sm.inShiftMovesRight = inShiftMovesRight
}
func (sm *SM) SetAutopullEnabled(autopullEnabled bool) {
	sm.autopullEnabled = autopullEnabled
}
func (sm *SM) SetAutopushEnabled(autopushEnabled bool) {
	sm.autopushEnabled = autopushEnabled
}

var ErrSMSetBaseSidesetPinOutOfRange = errors.New("base sideset pin must be less than 32")

func (sm *SM) SetBaseSidesetPin(baseSidesetPin uint) error {
	if baseSidesetPin > 31 {
		return ErrSMSetBaseSidesetPinOutOfRange
	}
	sm.baseSidesetPin = baseSidesetPin
	return nil
}

var ErrSMSetBitCountSidesetOutOfRange = errors.New("sideset bit count must be less than 6")

func (sm *SM) SetBitCountSideset(bitCountSideset uint) error {
	if bitCountSideset > 5 {
		return ErrSMSetBitCountSidesetOutOfRange
	}
	sm.bitCountSideset = bitCountSideset
	return nil
}

var ErrSMSetBaseSetPinOutOfRange = errors.New("base set pin must be less than 32")

func (sm *SM) SetBaseSetPin(baseSetPin uint) error {
	if baseSetPin > 31 {
		return ErrSMSetBaseSetPinOutOfRange
	}
	sm.baseSetPin = baseSetPin
	return nil
}

var ErrSMSetPinCountSetOutOfRange = errors.New("set pin count must be less than 6")

func (sm *SM) SetPinCountSet(pinCountSet uint) error {
	if pinCountSet > 5 {
		return ErrSMSetPinCountSetOutOfRange
	}
	sm.pinCountSet = pinCountSet
	return nil
}

var ErrSMSetBaseInPinOutOfRange = errors.New("base in pin must be less than 32")

func (sm *SM) SetBaseInPin(baseInPin uint) error {
	if baseInPin > 31 {
		return ErrSMSetBaseInPinOutOfRange
	}
	sm.baseInPin = baseInPin
	return nil
}

var ErrSMSetBaseOutPinOutOfRange = errors.New("base out pin must be less than 32")

func (sm *SM) SetBaseOutPin(baseOutPin uint) error {
	if baseOutPin > 31 {
		return ErrSMSetBaseOutPinOutOfRange
	}
	sm.baseOutPin = baseOutPin
	return nil
}

var ErrSMSetPinCountOutOutOfRange = errors.New("out pin count must be less than 33")

func (sm *SM) SetPinCountOut(pinCountOut uint) error {
	if pinCountOut > 32 {
		return ErrSMSetPinCountOutOutOfRange
	}
	sm.pinCountOut = pinCountOut
	return nil
}

var ErrSMSetJumpPinOutOfRange = errors.New("jump pin must be less than 32")

func (sm *SM) SetJumpPin(jumpPin uint) error {
	if jumpPin > 31 {
		return ErrSMSetJumpPinOutOfRange
	}
	sm.jumpPin = jumpPin
	return nil
}

type FIFOJoin uint

const (
	FIFOJoinNone FIFOJoin = iota
	FIFOJoinRX
	FIFOJoinTX
	FIFOJoinDisabled
)

var ErrSMSetFIFOJoinInvalidJoin = errors.New("invalid FIFO join")

func (sm *SM) SetFIFOJoin(join FIFOJoin) error {
	switch join {
	case FIFOJoinNone:
		sm.fifoRX.Resize(4)
		sm.fifoTX.Resize(4)
	case FIFOJoinRX:
		sm.fifoRX.Resize(8)
		sm.fifoTX.Resize(0)
	case FIFOJoinTX:
		sm.fifoRX.Resize(0)
		sm.fifoTX.Resize(8)
	case FIFOJoinDisabled:
		sm.fifoRX.Resize(0)
		sm.fifoTX.Resize(0)
	default:
		return ErrSMSetFIFOJoinInvalidJoin
	}
	return nil
}

func (sm *SM) ForceInstruction(forcedInstruction uint16) {
	sm.forcedInstruction = forcedInstruction
	sm.newForcedInstruction = true
}

func (sm *SM) RestartClockDivider() {
	sm.clockDividerTicksRemaining = 0
	sm.clockDividerFractionAccumulator = 0
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

var ErrSMSetClockDivisorIntegerOutOfRange = errors.New("clock divisor integer must be in the range 1 to 65536")

func (sm *SM) SetClockDivisorInteger(divInt uint) error {
	if divInt < 1 || divInt > 65536 {
		return ErrSMSetClockDivisorIntegerOutOfRange
	}
	if divInt == 65536 {
		sm.clockDivisorInteger = 0
	} else {
		sm.clockDivisorInteger = uint16(divInt)
	}
	return nil
}
func (sm *SM) SetClockDivisorFractional(divFrac uint8) {
	sm.clockDivisorFractional = divFrac
}

func (sm *SM) Controller() Controller {
	return sm
}

func (sm *SM) SetPinInputs(pinInputs uint32) {
	sm.pinInputs = pinInputs
}
func (sm *SM) SetIRQInputs(irqInputs uint8) {
	sm.irqInputs = irqInputs
}

var ErrSMForcedInstructionStalledDuringEXECdInstruction = errors.New("a forced instruction stalled while an EXEC'd instruction was pending, which would've overwritten the EXEC'd instruction")

func (sm *SM) Tick() error {
	if !sm.stickyOutSetAssertionEnabled {
		sm.pinOutputEnables = 0
		sm.pinOutputEnablesMask = 0
		sm.pinOutputs = 0
		sm.pinOutputsMask = 0
	}
	sm.pinSidesets = 0
	sm.pinSidesetsMask = 0
	sm.irqWrites = 0
	sm.irqWritesMask = 0

	if sm.newForcedInstruction {
		sm.newForcedInstruction = false
		jumped, stalled, stalledIRQ, err := sm.execute(sm.forcedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing forced instruction: %w", err)
		}
		sm.jumped = jumped
		sm.forcedInstructionStalled = stalled || stalledIRQ
		if sm.forcedInstructionStalled {
			if sm.newEXECdInstruction || sm.execdInstructionStalled {
				return ErrSMForcedInstructionStalledDuringEXECdInstruction
			}
			sm.latchedInstruction = sm.forcedInstruction
		}
	} else if sm.forcedInstructionStalled {
		jumped, stalled, stalledIRQ, err := sm.execute(sm.latchedInstruction, true)
		if err != nil {
			return fmt.Errorf("error executing stalled forced instruction: %w", err)
		}
		sm.jumped = jumped
		sm.forcedInstructionStalled = stalled || stalledIRQ
	} else if sm.clockDividerTicksRemaining == 0 && sm.enabled {
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
	if sm.newEXECdInstruction {
		sm.newEXECdInstruction = false
		jumped, stalled, stalledIRQ, err := sm.execute(sm.latchedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing EXEC'd instruction: %w", err)
		}
		sm.jumped = jumped
		sm.execdInstructionStalled = stalled || stalledIRQ
	} else if sm.execdInstructionStalled {
		jumped, stalled, stalledIRQ, err := sm.execute(sm.latchedInstruction, true)
		if err != nil {
			return fmt.Errorf("error executing stalled EXEC'd instruction: %w", err)
		}
		sm.jumped = jumped
		sm.execdInstructionStalled = stalled || stalledIRQ
	} else if sm.stalled || sm.stalledIRQ {
		instr, err := sm.memoryReader.Read(sm.programCounter)
		if err != nil {
			return fmt.Errorf("error reading instruction memory: %w", err)
		}
		jumped, stalled, stalledIRQ, err := sm.execute(instr, true)
		if err != nil {
			return fmt.Errorf("error executing stalled instruction: %w", err)
		}
		if !jumped && !(stalled || stalledIRQ) {
			sm.incrementProgramCounter()
		}
		sm.jumped = jumped
		sm.stalled = stalled
		sm.stalledIRQ = stalledIRQ
	} else if sm.delaysRemaining > 0 {
		sm.delaysRemaining--
	} else {
		instr, err := sm.memoryReader.Read(sm.programCounter)
		if err != nil {
			return fmt.Errorf("error reading instruction memory: %w", err)
		}
		jumped, stalled, stalledIRQ, err := sm.execute(instr, false)
		if err != nil {
			return fmt.Errorf("error executing instruction: %w", err)
		}
		if !jumped && !(stalled || stalledIRQ) {
			sm.incrementProgramCounter()
		}
		sm.jumped = jumped
		sm.stalled = stalled
		sm.stalledIRQ = stalledIRQ
	}
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

func (sm *SM) execute(instr uint16, currentlyStalled bool) (jumped bool, stalled bool, stalledIRQ bool, err error) {
	instructionType := instruction((instr >> 13) & 0b111)

	switch instructionType {
	case instructionJump:
		condition := jumpCondition((instr >> 5) & 0b111)
		address := uint(instr & 0b11111)
		jumped, err = sm.executeJump(condition, address)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing jump: %w", err)
		}
	case instructionWait:
		polarity := (instr>>7)&0b1 == 1
		source := waitSource((instr >> 5) & 0b11)
		index := uint(instr & 0b11111)
		stalled, err = sm.executeWait(polarity, source, index)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing wait: %w", err)
		}
	case instructionIn:
		source := inSource((instr >> 5) & 0b111)
		numberOfBits := uint(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		stalled, err = sm.executeIn(source, numberOfBits, currentlyStalled)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing in: %w", err)
		}
	case instructionOut:
		destination := outDestination((instr >> 5) & 0b111)
		numberOfBits := uint(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		jumped, stalled, err = sm.executeOut(destination, numberOfBits)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing out: %w", err)
		}
	case instructionPushPull:
		isPull := (instr>>7)&0b1 == 1
		ifThreshold := (instr>>6)&0b1 == 1
		block := (instr>>5)&0b1 == 1
		stalled, err = sm.executePushOrPull(isPull, ifThreshold, block)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing push/pull: %w", err)
		}
	case instructionMove:
		destination := moveDestination((instr >> 5) & 0b111)
		operation := moveOperation((instr >> 3) & 0b11)
		source := moveSource(instr & 0b111)
		jumped, err = sm.executeMove(destination, operation, source)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing move: %w", err)
		}
	case instructionIRQ:
		clearIRQ := (instr>>6)&0b1 == 1
		waitIRQ := (instr>>5)&0b1 == 1
		index := uint(instr & 0b11111)
		stalledIRQ, err = sm.executeIRQ(clearIRQ, waitIRQ, index, currentlyStalled)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing irq: %w", err)
		}
	case instructionSet:
		destination := setDestination((instr >> 5) & 0b111)
		data := uint(instr & 0b11111)
		err := sm.executeSet(destination, data)
		if err != nil {
			return false, false, false, fmt.Errorf("error executing set: %w", err)
		}
	default:
		return false, false, false, ErrSMInvalidInstructionType
	}

	if sm.autopullEnabled && sm.outputShiftRegisterCounter >= sm.pullThreshold && !sm.fifoTX.IsEmpty() {
		osr, err := sm.fifoTX.Read()
		if err != nil {
			return false, false, false, fmt.Errorf("error reading TX FIFO for autopull: %w", err)
		}
		sm.outputShiftRegister = osr
		sm.outputShiftRegisterCounter = 0
	}

	delaySidesetData := (instr >> 8) & 0b11111

	var delayMask uint16 = (0b1 << (5 - sm.bitCountSideset)) - 1
	var sidesetMask = ^delayMask
	if sm.sidesetIsOptional {
		sidesetMask &= 0b01111
	}

	doSideset := (sm.bitCountSideset > 0) && (!sm.sidesetIsOptional || ((delaySidesetData>>4)&0b1 == 1))
	if doSideset {
		sidesetData := (delaySidesetData & sidesetMask) >> (5 - sm.bitCountSideset)
		pinCount := sm.bitCountSideset
		if sm.sidesetIsOptional {
			pinCount -= 1
		}
		var pinData uint32 = bits.RotateLeft32(uint32(sidesetData), int(sm.baseSidesetPin))
		var pinMask uint32 = bits.RotateLeft32((0b1<<pinCount)-1, int(sm.baseSidesetPin))
		sm.pinSidesets = pinData
		sm.pinSidesetsMask = pinMask
	}

	doDelay := (sm.bitCountSideset < 5) && !sm.newForcedInstruction
	if doDelay {
		delayData := (delaySidesetData & delayMask)
		sm.delaysRemaining = uint(delayData)
	} else {
		sm.delaysRemaining = 0
	}

	return jumped, stalled, stalledIRQ, nil
}

func (sm *SM) incrementProgramCounter() {
	if sm.programCounter == sm.wrapFromAddress {
		sm.programCounter = sm.wrapToAddress
	} else {
		sm.programCounter = (sm.programCounter + 1) % 32
	}
}

var ErrSMXRegisterNotInitialized = errors.New("X register not initialized")
var ErrSMYRegisterNotInitialized = errors.New("Y register not initialized")

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

func (sm *SM) executeJump(condition jumpCondition, address uint) (jumped bool, err error) {
	var shouldJump bool
	switch condition {
	case jumpAlways:
		shouldJump = true
	case jumpXZero:
		if !sm.xRegisterInitialized {
			return false, ErrSMXRegisterNotInitialized
		}
		shouldJump = (sm.xRegister == 0)
	case jumpXNonZeroThenDecrement:
		if !sm.xRegisterInitialized {
			return false, ErrSMXRegisterNotInitialized
		}
		shouldJump = (sm.xRegister != 0)
		sm.xRegister--
	case jumpYZero:
		if !sm.yRegisterInitialized {
			return false, ErrSMYRegisterNotInitialized
		}
		shouldJump = (sm.yRegister == 0)
	case jumpYNonZeroThenDecrement:
		if !sm.yRegisterInitialized {
			return false, ErrSMYRegisterNotInitialized
		}
		shouldJump = (sm.yRegister != 0)
		sm.yRegister--
	case jumpXNotEqualY:
		if !sm.xRegisterInitialized {
			return false, ErrSMXRegisterNotInitialized
		}
		if !sm.yRegisterInitialized {
			return false, ErrSMYRegisterNotInitialized
		}
		shouldJump = (sm.xRegister != sm.yRegister)
	case jumpPin:
		shouldJump = (sm.pinInputs>>sm.jumpPin)&0b1 == 1
	case jumpOSRENotEmpty:
		shouldJump = (sm.outputShiftRegisterCounter < sm.pullThreshold)
	default:
		return false, ErrSMJumpInvalidCondition
	}
	if shouldJump {
		sm.programCounter = address
		jumped = true
	}
	return jumped, nil
}

type waitSource uint

const (
	waitSourceGPIO waitSource = 0b00
	waitSourcePin  waitSource = 0b01
	waitSourceIRQ  waitSource = 0b10
)

var ErrSMWaitInvalidSource = errors.New("invalid source")

func (sm *SM) executeWait(polarity bool, source waitSource, index uint) (stalled bool, err error) {
	switch source {
	case waitSourceGPIO:
		stalled = ((sm.pinInputs>>index)&0b1 == 1) != polarity
	case waitSourcePin:
		pin := (sm.baseInPin + index) % 32
		stalled = ((sm.pinInputs>>pin)&0b1 == 1) != polarity
	case waitSourceIRQ:
		relative := (index>>4)&0b1 == 1
		irq := index & 0b111
		if relative {
			upperBit := irq & 0b100
			lowerBits := (irq + sm.index) & 0b011
			irq = upperBit | lowerBits
		}
		stalled = ((sm.irqInputs>>irq)&0b1 == 1) != polarity
		if !stalled && polarity {
			var clearIRQMask uint8 = (0b1 << irq)
			sm.irqWrites &= ^clearIRQMask
			sm.irqWritesMask |= clearIRQMask
		}
	default:
		return false, ErrSMWaitInvalidSource
	}
	return stalled, nil
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

func (sm *SM) executeIn(source inSource, count uint, currentlyStalled bool) (stalled bool, err error) {
	if !currentlyStalled {
		var data uint32
		var mask uint32 = (0b1 << count) - 1
		switch source {
		case inSourcePins:
			data = bits.RotateLeft32(sm.pinInputs, -int(sm.baseInPin)) & mask
		case inSourceX:
			if !sm.xRegisterInitialized {
				return false, ErrSMXRegisterNotInitialized
			}
			data = (sm.xRegister & mask)
		case inSourceY:
			if !sm.yRegisterInitialized {
				return false, ErrSMYRegisterNotInitialized
			}
			data = (sm.yRegister & mask)
		case inSourceNull:
			data = 0
		case inSourceISR:
			data = (sm.inputShiftRegister & mask)
		case inSourceOSR:
			data = (sm.outputShiftRegister & mask)
		default:
			return false, ErrSMInInvalidSource
		}
		if sm.inShiftMovesRight {
			sm.inputShiftRegister >>= count
			sm.inputShiftRegister |= (data << (32 - count))
		} else {
			sm.inputShiftRegister <<= count
			sm.inputShiftRegister |= data
		}
		sm.inputShiftRegisterCounter += count
		if sm.inputShiftRegisterCounter > 32 {
			sm.inputShiftRegisterCounter = 32
		}
	}
	shouldPush := sm.autopushEnabled && (sm.inputShiftRegisterCounter >= sm.pushThreshold)
	stalled = shouldPush && sm.fifoRX.IsFull()
	if shouldPush && !sm.fifoRX.IsFull() {
		err := sm.fifoRX.Write(sm.inputShiftRegister)
		if err != nil {
			return false, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
		}
		sm.inputShiftRegister = 0
		sm.inputShiftRegisterCounter = 0
	}
	return stalled, nil
}

type outDestination uint

const (
	outDestinationPins               outDestination = 0b000
	outDestinationX                  outDestination = 0b001
	outDestinationY                  outDestination = 0b010
	outDestinationNull               outDestination = 0b011
	outDestinationPinDirections      outDestination = 0b100
	outDestinationProgramCounter     outDestination = 0b101
	outDestinationInputShiftRegister outDestination = 0b110
	outDestinationEXEC               outDestination = 0b111
)

var ErrSMOutInvalidDestination = errors.New("invalid out destination")

func (sm *SM) executeOut(destination outDestination, count uint) (jumped bool, stalled bool, err error) {
	shouldPull := sm.autopullEnabled && (sm.outputShiftRegisterCounter >= sm.pullThreshold)
	if shouldPull {
		if !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return false, false, fmt.Errorf("error reading TX fifo for autopull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		}
		return false, true, nil
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
	if sm.outputShiftRegisterCounter > 32 {
		sm.outputShiftRegisterCounter = 32
	}

	isPinOperation := destination == outDestinationPins || destination == outDestinationPinDirections

	outWriteEnable := ((data >> sm.outWriteEnableBitIndex) & 0b1) == 1
	writePins := true
	if isPinOperation && sm.outWriteEnableUsed && !outWriteEnable {
		writePins = false
		if sm.stickyOutSetAssertionEnabled {
			sm.pinOutputEnables = 0
			sm.pinOutputEnablesMask = 0
			sm.pinOutputs = 0
			sm.pinOutputsMask = 0
		}
	}

	var pinData uint32 = bits.RotateLeft32(data, int(sm.baseOutPin))
	var pinMask uint32 = bits.RotateLeft32((0b1<<sm.pinCountOut)-1, int(sm.baseOutPin))

	switch destination {
	case outDestinationPins:
		if writePins {
			sm.pinOutputs = pinData
			sm.pinOutputsMask = pinMask
		}
	case outDestinationX:
		sm.xRegister = data
		sm.xRegisterInitialized = true
	case outDestinationY:
		sm.yRegister = data
		sm.yRegisterInitialized = true
	case outDestinationNull:
		// discards data
	case outDestinationPinDirections:
		if writePins {
			sm.pinOutputEnables = pinData
			sm.pinOutputEnablesMask = pinMask
		}
	case outDestinationProgramCounter:
		sm.programCounter = uint(data % 32)
		jumped = true
	case outDestinationInputShiftRegister:
		sm.inputShiftRegister = data
		sm.inputShiftRegisterCounter = count
	case outDestinationEXEC:
		sm.latchedInstruction = uint16(data)
		sm.newEXECdInstruction = true
	default:
		return false, false, ErrSMOutInvalidDestination
	}

	return jumped, false, nil
}

func (sm *SM) executePushOrPull(isPull bool, ifThreshold bool, block bool) (stalled bool, err error) {
	if isPull {
		shouldPull := (!ifThreshold && !sm.autopullEnabled) || (sm.outputShiftRegisterCounter >= sm.pullThreshold)
		stalled = block && shouldPull && sm.fifoTX.IsEmpty()
		if shouldPull && !sm.fifoTX.IsEmpty() {
			osr, err := sm.fifoTX.Read()
			if err != nil {
				return false, fmt.Errorf("error reading TX fifo for pull: %w", err)
			}
			sm.outputShiftRegister = osr
			sm.outputShiftRegisterCounter = 0
		} else if shouldPull && !block {
			if !sm.xRegisterInitialized {
				return false, ErrSMXRegisterNotInitialized
			}
			sm.outputShiftRegister = sm.xRegister
			sm.outputShiftRegisterCounter = 0
		}
	} else {
		shouldPush := !ifThreshold || (sm.inputShiftRegisterCounter >= sm.pushThreshold)
		stalled = block && shouldPush && sm.fifoRX.IsFull()
		if shouldPush && !sm.fifoRX.IsFull() {
			err := sm.fifoRX.Write(sm.inputShiftRegister)
			if err != nil {
				return false, fmt.Errorf("error writing RX FIFO for autopush: %w", err)
			}
			sm.inputShiftRegister = 0
			sm.inputShiftRegisterCounter = 0
		}
	}
	return stalled, nil
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

func (sm *SM) executeMove(destination moveDestination, operation moveOperation, source moveSource) (jumped bool, err error) {
	var sourceData uint32
	switch source {
	case moveSourcePins:
		sourceData = bits.RotateLeft32(sm.pinInputs, -int(sm.baseInPin))
	case moveSourceX:
		if !sm.xRegisterInitialized {
			return false, ErrSMXRegisterNotInitialized
		}
		sourceData = sm.xRegister
	case moveSourceY:
		if !sm.yRegisterInitialized {
			return false, ErrSMYRegisterNotInitialized
		}
		sourceData = sm.yRegister
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
		if sm.autopullEnabled {
			return false, ErrSMMoveSourceOSRDisallowedWhenAutopullEnabled
		} else {
			sourceData = sm.outputShiftRegister
		}
	default:
		return false, ErrSMInvalidMoveSource
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
		return false, ErrSMInvalidMoveOperation
	}

	switch destination {
	case moveDestinationPins:
		outWriteEnable := ((modifiedData >> sm.outWriteEnableBitIndex) & 0b1) == 1
		writePins := true
		if sm.outWriteEnableUsed && !outWriteEnable {
			writePins = false
			if sm.stickyOutSetAssertionEnabled {
				sm.pinOutputEnables = 0
				sm.pinOutputEnablesMask = 0
				sm.pinOutputs = 0
				sm.pinOutputsMask = 0
			}
		}
		if writePins {
			sm.pinOutputsMask = bits.RotateLeft32((0b1<<sm.pinCountOut)-1, int(sm.baseOutPin))
			sm.pinOutputs = bits.RotateLeft32(modifiedData, int(sm.baseOutPin))
		}
	case moveDestinationX:
		sm.xRegister = modifiedData
		sm.xRegisterInitialized = true
	case moveDestinationY:
		sm.yRegister = modifiedData
		sm.yRegisterInitialized = true
	case moveDestinationEXEC:
		sm.latchedInstruction = uint16(modifiedData)
		sm.newEXECdInstruction = true
	case moveDestinationPC:
		sm.programCounter = uint(modifiedData % 32)
		jumped = true
	case moveDestinationISR:
		sm.inputShiftRegister = modifiedData
		sm.inputShiftRegisterCounter = 0
	case moveDestinationOSR:
		sm.outputShiftRegister = modifiedData
		sm.outputShiftRegisterCounter = 0
	default:
		return false, ErrSMInvalidMoveDestination
	}

	return jumped, nil
}

var ErrSMIRQClearWaitDisallowed = errors.New("malformed instruction, cannot clear and wait at the same time")

func (sm *SM) executeIRQ(clearIRQ bool, waitIRQ bool, index uint, currentlyStalled bool) (stalled bool, err error) {
	relative := (index>>4)&0b1 == 1
	irq := index & 0b111
	if relative {
		upperBit := irq & 0b100
		lowerBits := (irq + sm.index) & 0b011
		irq = upperBit | lowerBits
	}

	if currentlyStalled {
		return ((sm.irqInputs >> irq) & 0b1) == 1, nil
	}

	var irqMask uint8 = (0b1 << irq)
	if clearIRQ {
		if waitIRQ {
			return false, ErrSMIRQClearWaitDisallowed
		}
		sm.irqWrites &= ^irqMask
		sm.irqWritesMask |= irqMask
	} else {
		sm.irqWrites |= irqMask
		sm.irqWritesMask |= irqMask
		if waitIRQ {
			return true, nil
		}
	}

	return false, nil
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
	var pinData uint32 = bits.RotateLeft32(uint32(data), int(sm.baseSetPin))
	var pinMask uint32 = bits.RotateLeft32((0b1<<sm.pinCountSet)-1, int(sm.baseSetPin))
	switch destination {
	case setDestinationPins:
		sm.pinOutputs = pinData
		sm.pinOutputsMask = pinMask
	case setDestinationX:
		sm.xRegister = uint32(data)
		sm.xRegisterInitialized = true
	case setDestinationY:
		sm.yRegister = uint32(data)
		sm.yRegisterInitialized = true
	case setDestinationPinDirections:
		sm.pinOutputEnables = pinData
		sm.pinOutputEnablesMask = pinMask
	default:
		return ErrSMSetInvalidDestination
	}
	return nil
}
