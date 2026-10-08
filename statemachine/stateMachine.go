package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/headblockhead/pio/fifo"
	"github.com/headblockhead/pio/memory"
	"github.com/headblockhead/pio/simulation"
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
	SetLabel(string)

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

type Operator interface {
	simulation.Ticker

	SetPinInputs(uint32)
	SetIRQInputs(uint8)

	PinOutputEnables() uint32
	PinOutputEnablesMask() uint32
	PinOutputs() uint32
	PinOutputsMask() uint32
	SidesetControlsPinDirection() bool
	PinSidesets() uint32
	PinSidesetsMask() uint32
	IRQWrites() uint8
	IRQWritesMask() uint8
}

type StateMachine struct {
	id    simulation.ComponentIdentifier
	label string

	index uint

	memoryReader memory.Reader
	fifoRX       fifo.Writer
	fifoTX       fifo.Reader

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

func New(index uint, label string, memoryReader memory.Reader) *StateMachine {
	return &StateMachine{
		id:    simulation.NewComponentIdentifier(),
		label: label,

		index: index,

		memoryReader: memoryReader,
		fifoRX:       fifo.New(4),
		fifoTX:       fifo.New(4),

		pinInputs: 0,
		irqInputs: 0,

		enabled: false,

		pinOutputEnables:     0,
		pinOutputEnablesMask: 0,
		pinOutputs:           0,
		pinOutputsMask:       0,
		pinSidesets:          0,
		pinSidesetsMask:      0,
		irqWrites:            0,
		irqWritesMask:        0,

		sidesetIsOptional:            false,
		sidesetControlsPinDirection:  false,
		outWriteEnableUsed:           false,
		outWriteEnableBitIndex:       0,
		stickyOutSetAssertionEnabled: false,

		wrapFromAddress: 31,
		wrapToAddress:   0,

		statusValueUsesRXFIFO:      false,
		statusValueComparisonLevel: 0,
		pullThreshold:              32,
		pushThreshold:              32,
		outShiftMovesRight:         true,
		inShiftMovesRight:          true,
		autopullEnabled:            false,
		autopushEnabled:            false,

		baseSidesetPin:  0,
		bitCountSideset: 0,
		baseSetPin:      0,
		pinCountSet:     5,
		baseInPin:       0,
		baseOutPin:      0,
		pinCountOut:     0,
		jumpPin:         0,

		programCounter:           0,
		stalled:                  false,
		stalledIRQ:               false,
		jumped:                   false,
		delaysRemaining:          0,
		newForcedInstruction:     false,
		forcedInstruction:        0,
		forcedInstructionStalled: false,
		newEXECdInstruction:      false,
		execdInstructionStalled:  false,
		latchedInstruction:       0,

		outputShiftRegister:        0,
		outputShiftRegisterCounter: 32,
		inputShiftRegister:         0,
		inputShiftRegisterCounter:  0,
		xRegister:                  0,
		xRegisterInitialized:       false,
		yRegister:                  0,
		yRegisterInitialized:       false,

		clockDivisorInteger:    1,
		clockDivisorFractional: 0,
	}
}

func (sm *StateMachine) Observer() Observer {
	return sm
}

func (sm *StateMachine) Index() uint { return sm.index }

func (sm *StateMachine) FIFORXObserver() fifo.Observer { return sm.fifoRX.Observer() }
func (sm *StateMachine) FIFOTXObserver() fifo.Observer { return sm.fifoTX.Observer() }

func (sm *StateMachine) Enabled() bool { return sm.enabled }

func (sm *StateMachine) PinOutputEnables() uint32     { return sm.pinOutputEnables }
func (sm *StateMachine) PinOutputEnablesMask() uint32 { return sm.pinOutputEnablesMask }
func (sm *StateMachine) PinOutputs() uint32           { return sm.pinOutputs }
func (sm *StateMachine) PinOutputsMask() uint32       { return sm.pinOutputsMask }
func (sm *StateMachine) PinSidesets() uint32          { return sm.pinSidesets }
func (sm *StateMachine) PinSidesetsMask() uint32      { return sm.pinSidesetsMask }
func (sm *StateMachine) IRQWrites() uint8             { return sm.irqWrites }
func (sm *StateMachine) IRQWritesMask() uint8         { return sm.irqWritesMask }

func (sm *StateMachine) SidesetIsOptional() bool            { return sm.sidesetIsOptional }
func (sm *StateMachine) SidesetControlsPinDirection() bool  { return sm.sidesetControlsPinDirection }
func (sm *StateMachine) OutWriteEnableUsed() bool           { return sm.outWriteEnableUsed }
func (sm *StateMachine) OutWriteEnableBitIndex() uint       { return sm.outWriteEnableBitIndex }
func (sm *StateMachine) StickyOutSetAssertionEnabled() bool { return sm.stickyOutSetAssertionEnabled }

func (sm *StateMachine) WrapFromAddress() uint { return sm.wrapFromAddress }
func (sm *StateMachine) WrapToAddress() uint   { return sm.wrapToAddress }

func (sm *StateMachine) StatusValueUsesRXFIFO() bool      { return sm.statusValueUsesRXFIFO }
func (sm *StateMachine) StatusValueComparisonLevel() uint { return sm.statusValueComparisonLevel }
func (sm *StateMachine) PullThreshold() uint              { return sm.pullThreshold }
func (sm *StateMachine) PushThreshold() uint              { return sm.pushThreshold }
func (sm *StateMachine) OutShiftMovesRight() bool         { return sm.outShiftMovesRight }
func (sm *StateMachine) InShiftMovesRight() bool          { return sm.inShiftMovesRight }
func (sm *StateMachine) AutopullEnabled() bool            { return sm.autopullEnabled }
func (sm *StateMachine) AutopushEnabled() bool            { return sm.autopushEnabled }

func (sm *StateMachine) BaseSidesetPin() uint  { return sm.baseSidesetPin }
func (sm *StateMachine) BitCountSideset() uint { return sm.bitCountSideset }
func (sm *StateMachine) BaseSetPin() uint      { return sm.baseSetPin }
func (sm *StateMachine) PinCountSet() uint     { return sm.pinCountSet }
func (sm *StateMachine) BaseInPin() uint       { return sm.baseInPin }
func (sm *StateMachine) BaseOutPin() uint      { return sm.baseOutPin }
func (sm *StateMachine) PinCountOut() uint     { return sm.pinCountOut }
func (sm *StateMachine) JumpPin() uint         { return sm.jumpPin }

func (sm *StateMachine) ProgramCounter() uint           { return sm.programCounter }
func (sm *StateMachine) Stalled() bool                  { return sm.stalled }
func (sm *StateMachine) StalledIRQ() bool               { return sm.stalledIRQ }
func (sm *StateMachine) Jumped() bool                   { return sm.jumped }
func (sm *StateMachine) DelaysRemaining() uint          { return sm.delaysRemaining }
func (sm *StateMachine) NewForcedInstruction() bool     { return sm.newForcedInstruction }
func (sm *StateMachine) ForcedInstruction() uint16      { return sm.forcedInstruction }
func (sm *StateMachine) ForcedInstructionStalled() bool { return sm.forcedInstructionStalled }
func (sm *StateMachine) NewEXECdInstruction() bool      { return sm.newEXECdInstruction }
func (sm *StateMachine) EXECdInstructionStalled() bool  { return sm.execdInstructionStalled }
func (sm *StateMachine) LatchedInstruction() uint16     { return sm.latchedInstruction }

func (sm *StateMachine) OutputShiftRegister() uint32      { return sm.outputShiftRegister }
func (sm *StateMachine) OutputShiftRegisterCounter() uint { return sm.outputShiftRegisterCounter }
func (sm *StateMachine) InputShiftRegister() uint32       { return sm.inputShiftRegister }
func (sm *StateMachine) InputShiftRegisterCounter() uint  { return sm.inputShiftRegisterCounter }
func (sm *StateMachine) XRegister() uint32                { return sm.xRegister }
func (sm *StateMachine) XRegisterInitialized() bool       { return sm.xRegisterInitialized }
func (sm *StateMachine) YRegister() uint32                { return sm.yRegister }
func (sm *StateMachine) YRegisterInitialized() bool       { return sm.yRegisterInitialized }

func (sm *StateMachine) ClockDivisor() float32 {
	return clockDivisorToFloat32(sm.clockDivisorInteger, sm.clockDivisorFractional)
}
func (sm *StateMachine) ClockDivisorInteger() uint {
	if sm.clockDivisorInteger == 0 {
		return 65536
	} else {
		return uint(sm.clockDivisorInteger)
	}
}
func (sm *StateMachine) ClockDivisorFractional() uint8    { return sm.clockDivisorFractional }
func (sm *StateMachine) ClockDividerTicksRemaining() uint { return sm.clockDividerTicksRemaining }
func (sm *StateMachine) ClockDividerFractionAccumulator() uint8 {
	return sm.clockDividerFractionAccumulator
}

func (sm *StateMachine) Configurator() Configurator {
	return sm
}

func (sm *StateMachine) SetLabel(label string) { sm.label = label }

func (sm *StateMachine) Restart() {
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
func (sm *StateMachine) SetEnabled(enabled bool) {
	sm.enabled = enabled
}
func (sm *StateMachine) SetSidesetIsOptional(sidesetIsOptional bool) {
	sm.sidesetIsOptional = sidesetIsOptional
}
func (sm *StateMachine) SetSidesetControlsPinDirection(sidesetControlsPinDirection bool) {
	sm.sidesetControlsPinDirection = sidesetControlsPinDirection
}
func (sm *StateMachine) SetOutWriteEnableUsed(outWriteEnableUsed bool) {
	sm.outWriteEnableUsed = outWriteEnableUsed
}

var ErrValueOutOfRange = errors.New("value out of range")

func (sm *StateMachine) SetOutWriteEnableBitIndex(outWriteEnableBitIndex uint) error {
	if outWriteEnableBitIndex > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, outWriteEnableBitIndex)
	}
	sm.outWriteEnableBitIndex = outWriteEnableBitIndex
	return nil
}
func (sm *StateMachine) SetStickyOutSetAssertionEnabled(stickyOutSetAssertionEnabled bool) {
	sm.stickyOutSetAssertionEnabled = stickyOutSetAssertionEnabled
}
func (sm *StateMachine) SetWrapFromAddress(wrapFromAddress uint) error {
	if wrapFromAddress > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, wrapFromAddress)
	}
	sm.wrapFromAddress = wrapFromAddress
	return nil
}
func (sm *StateMachine) SetWrapToAddress(wrapToAddress uint) error {
	if wrapToAddress > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, wrapToAddress)
	}
	sm.wrapToAddress = wrapToAddress
	return nil
}
func (sm *StateMachine) SetStatusValueUsesRXFIFO(statusValueUsesRXFIFO bool) {
	sm.statusValueUsesRXFIFO = statusValueUsesRXFIFO
}
func (sm *StateMachine) SetStatusValueComparisonLevel(statusValueComparisonLevel uint) error {
	if statusValueComparisonLevel > 15 {
		return fmt.Errorf("%w: %d, should be < 16", ErrValueOutOfRange, statusValueComparisonLevel)
	}
	sm.statusValueComparisonLevel = statusValueComparisonLevel
	return nil
}
func (sm *StateMachine) SetPullThreshold(pullThreshold uint) error {
	if pullThreshold > 32 || pullThreshold < 1 {
		return fmt.Errorf("%w: %d, should be >0 and <32", ErrValueOutOfRange, pullThreshold)
	}
	sm.pullThreshold = pullThreshold
	return nil
}
func (sm *StateMachine) SetPushThreshold(pushThreshold uint) error {
	if pushThreshold > 32 || pushThreshold < 1 {
		return fmt.Errorf("%w: %d, should be >0 and <32", ErrValueOutOfRange, pushThreshold)
	}
	sm.pushThreshold = pushThreshold
	return nil
}
func (sm *StateMachine) SetOutShiftMovesRight(outShiftMovesRight bool) {
	sm.outShiftMovesRight = outShiftMovesRight
}
func (sm *StateMachine) SetInShiftMovesRight(inShiftMovesRight bool) {
	sm.inShiftMovesRight = inShiftMovesRight
}
func (sm *StateMachine) SetAutopullEnabled(autopullEnabled bool) {
	sm.autopullEnabled = autopullEnabled
}
func (sm *StateMachine) SetAutopushEnabled(autopushEnabled bool) {
	sm.autopushEnabled = autopushEnabled
}
func (sm *StateMachine) SetBaseSidesetPin(baseSidesetPin uint) error {
	if baseSidesetPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseSidesetPin)
	}
	sm.baseSidesetPin = baseSidesetPin
	return nil
}
func (sm *StateMachine) SetBitCountSideset(bitCountSideset uint) error {
	if bitCountSideset > 5 {
		return fmt.Errorf("%w: %d, should be < 6", ErrValueOutOfRange, bitCountSideset)
	}
	sm.bitCountSideset = bitCountSideset
	return nil
}
func (sm *StateMachine) SetBaseSetPin(baseSetPin uint) error {
	if baseSetPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseSetPin)
	}
	sm.baseSetPin = baseSetPin
	return nil
}
func (sm *StateMachine) SetPinCountSet(pinCountSet uint) error {
	if pinCountSet > 5 {
		return fmt.Errorf("%w: %d, should be < 6", ErrValueOutOfRange, pinCountSet)
	}
	sm.pinCountSet = pinCountSet
	return nil
}
func (sm *StateMachine) SetBaseInPin(baseInPin uint) error {
	if baseInPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseInPin)
	}
	sm.baseInPin = baseInPin
	return nil
}
func (sm *StateMachine) SetBaseOutPin(baseOutPin uint) error {
	if baseOutPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseOutPin)
	}
	sm.baseOutPin = baseOutPin
	return nil
}
func (sm *StateMachine) SetPinCountOut(pinCountOut uint) error {
	if pinCountOut > 32 {
		return fmt.Errorf("%w: %d, should be < 33", ErrValueOutOfRange, pinCountOut)
	}
	sm.pinCountOut = pinCountOut
	return nil
}
func (sm *StateMachine) SetJumpPin(jumpPin uint) error {
	if jumpPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, jumpPin)
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

var ErrFIFOJoinInvalid = errors.New("FIFO join invalid")

func (sm *StateMachine) SetFIFOJoin(join FIFOJoin) error {
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
		return ErrFIFOJoinInvalid
	}
	return nil
}

func (sm *StateMachine) ForceInstruction(forcedInstruction uint16) {
	sm.forcedInstruction = forcedInstruction
	sm.newForcedInstruction = true
}

func (sm *StateMachine) RestartClockDivider() {
	sm.clockDividerTicksRemaining = 0
	sm.clockDividerFractionAccumulator = 0
}
func (sm *StateMachine) SetClockDivisor(divider float32) error {
	divInt, divFrac, err := clockDivisorFromFloat32(divider)
	if err != nil {
		return err
	}
	sm.clockDivisorInteger = divInt
	sm.clockDivisorFractional = divFrac
	return nil
}
func (sm *StateMachine) SetClockDivisorInteger(divInt uint) error {
	if divInt < 1 || divInt > 65536 {
		return fmt.Errorf("%w: %d, should be >0 and <65537", ErrValueOutOfRange, divInt)
	}
	if divInt == 65536 {
		sm.clockDivisorInteger = 0
	} else {
		sm.clockDivisorInteger = uint16(divInt)
	}
	return nil
}
func (sm *StateMachine) SetClockDivisorFractional(divFrac uint8) {
	sm.clockDivisorFractional = divFrac
}

func (sm *StateMachine) Operator() Operator {
	return sm
}

func (sm *StateMachine) ID() simulation.ComponentIdentifier { return sm.id }
func (sm *StateMachine) Label() string                      { return sm.label }

func (sm *StateMachine) SetPinInputs(pinInputs uint32) {
	sm.pinInputs = pinInputs
}
func (sm *StateMachine) SetIRQInputs(irqInputs uint8) {
	sm.irqInputs = irqInputs
}

// ErrForcedInstructionStalledDuringEXECdInstruction represents an edge-case error in the PIO, where
// a forced instruction stalled while an EXEC'd instruction was pending,
// which would've (in real hardware) overwritten the EXEC'd instruction in the instruction latch.
var ErrForcedInstructionStalledDuringEXECdInstruction = errors.New("forced instruction stalled during EXEC'd instruction")

func (sm *StateMachine) Tick() error {
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
		jumped, stalled, stalledIRQ, err := sm.execute(sm.forcedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing forced instruction: %w", err)
		}
		sm.newForcedInstruction = false
		sm.jumped = jumped
		sm.forcedInstructionStalled = stalled || stalledIRQ
		if sm.forcedInstructionStalled {
			if sm.newEXECdInstruction || sm.execdInstructionStalled {
				return ErrForcedInstructionStalledDuringEXECdInstruction
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

func (sm *StateMachine) dividedTick() error {
	if sm.newEXECdInstruction {
		jumped, stalled, stalledIRQ, err := sm.execute(sm.latchedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing EXEC'd instruction: %w", err)
		}
		sm.newEXECdInstruction = false
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
			return fmt.Errorf("error reading address %d of instruction memory: %w", sm.programCounter, err)
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
			return fmt.Errorf("error reading address %d of instruction memory: %w", sm.programCounter, err)
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

var ErrInstructionTypeInvalid = errors.New("instruction type invalid")

func (sm *StateMachine) execute(instr uint16, currentlyStalled bool) (jumped bool, stalled bool, stalledIRQ bool, err error) {
	instructionType := instruction((instr >> 13) & 0b111)

	switch instructionType {
	case instructionJump:
		condition := jumpCondition((instr >> 5) & 0b111)
		address := uint(instr & 0b11111)
		jumped, err = sm.executeJump(condition, address)
		if err != nil {
			return false, false, false, fmt.Errorf("jump: %w", err)
		}
	case instructionWait:
		polarity := (instr>>7)&0b1 == 1
		source := waitSource((instr >> 5) & 0b11)
		index := uint(instr & 0b11111)
		stalled, err = sm.executeWait(polarity, source, index)
		if err != nil {
			return false, false, false, fmt.Errorf("wait: %w", err)
		}
	case instructionIn:
		source := inSource((instr >> 5) & 0b111)
		numberOfBits := uint(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		stalled, err = sm.executeIn(source, numberOfBits, currentlyStalled)
		if err != nil {
			return false, false, false, fmt.Errorf("in: %w", err)
		}
	case instructionOut:
		destination := outDestination((instr >> 5) & 0b111)
		numberOfBits := uint(instr & 0b11111)
		if numberOfBits == 0 {
			numberOfBits = 32
		}
		jumped, stalled, err = sm.executeOut(destination, numberOfBits)
		if err != nil {
			return false, false, false, fmt.Errorf("out: %w", err)
		}
	case instructionPushPull:
		isPull := (instr>>7)&0b1 == 1
		ifThreshold := (instr>>6)&0b1 == 1
		block := (instr>>5)&0b1 == 1
		stalled, err = sm.executePushOrPull(isPull, ifThreshold, block)
		if err != nil {
			return false, false, false, fmt.Errorf("push/pull: %w", err)
		}
	case instructionMove:
		destination := moveDestination((instr >> 5) & 0b111)
		operation := moveOperation((instr >> 3) & 0b11)
		source := moveSource(instr & 0b111)
		jumped, err = sm.executeMove(destination, operation, source)
		if err != nil {
			return false, false, false, fmt.Errorf("move: %w", err)
		}
	case instructionIRQ:
		clearIRQ := (instr>>6)&0b1 == 1
		waitIRQ := (instr>>5)&0b1 == 1
		index := uint(instr & 0b11111)
		stalledIRQ, err = sm.executeIRQ(clearIRQ, waitIRQ, index, currentlyStalled)
		if err != nil {
			return false, false, false, fmt.Errorf("irq: %w", err)
		}
	case instructionSet:
		destination := setDestination((instr >> 5) & 0b111)
		data := uint(instr & 0b11111)
		err := sm.executeSet(destination, data)
		if err != nil {
			return false, false, false, fmt.Errorf("set: %w", err)
		}
	default:
		return false, false, false, ErrInstructionTypeInvalid
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

func (sm *StateMachine) incrementProgramCounter() {
	if sm.programCounter == sm.wrapFromAddress {
		sm.programCounter = sm.wrapToAddress
	} else {
		sm.programCounter = (sm.programCounter + 1) % 32
	}
}

var ErrXRegisterNotInitialized = errors.New("X register not initialized")
var ErrYRegisterNotInitialized = errors.New("Y register not initialized")

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

var ErrJumpConditionInvalid = errors.New("jump condition invalid")

func (sm *StateMachine) executeJump(condition jumpCondition, address uint) (jumped bool, err error) {
	var shouldJump bool
	switch condition {
	case jumpAlways:
		shouldJump = true
	case jumpXZero:
		if !sm.xRegisterInitialized {
			return false, ErrXRegisterNotInitialized
		}
		shouldJump = (sm.xRegister == 0)
	case jumpXNonZeroThenDecrement:
		if !sm.xRegisterInitialized {
			return false, ErrXRegisterNotInitialized
		}
		shouldJump = (sm.xRegister != 0)
		sm.xRegister--
	case jumpYZero:
		if !sm.yRegisterInitialized {
			return false, ErrYRegisterNotInitialized
		}
		shouldJump = (sm.yRegister == 0)
	case jumpYNonZeroThenDecrement:
		if !sm.yRegisterInitialized {
			return false, ErrYRegisterNotInitialized
		}
		shouldJump = (sm.yRegister != 0)
		sm.yRegister--
	case jumpXNotEqualY:
		if !sm.xRegisterInitialized {
			return false, ErrXRegisterNotInitialized
		}
		if !sm.yRegisterInitialized {
			return false, ErrYRegisterNotInitialized
		}
		shouldJump = (sm.xRegister != sm.yRegister)
	case jumpPin:
		shouldJump = (sm.pinInputs>>sm.jumpPin)&0b1 == 1
	case jumpOSRENotEmpty:
		shouldJump = (sm.outputShiftRegisterCounter < sm.pullThreshold)
	default:
		return false, ErrJumpConditionInvalid
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

var ErrWaitSourceInvalid = errors.New("wait source invalid")

func (sm *StateMachine) executeWait(polarity bool, source waitSource, index uint) (stalled bool, err error) {
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
		return false, ErrWaitSourceInvalid
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

var ErrInSourceInvalid = errors.New("in source invalid")

func (sm *StateMachine) executeIn(source inSource, count uint, currentlyStalled bool) (stalled bool, err error) {
	if !currentlyStalled {
		var data uint32
		var mask uint32 = (0b1 << count) - 1
		switch source {
		case inSourcePins:
			data = bits.RotateLeft32(sm.pinInputs, -int(sm.baseInPin)) & mask
		case inSourceX:
			if !sm.xRegisterInitialized {
				return false, ErrXRegisterNotInitialized
			}
			data = (sm.xRegister & mask)
		case inSourceY:
			if !sm.yRegisterInitialized {
				return false, ErrYRegisterNotInitialized
			}
			data = (sm.yRegister & mask)
		case inSourceNull:
			data = 0
		case inSourceISR:
			data = (sm.inputShiftRegister & mask)
		case inSourceOSR:
			data = (sm.outputShiftRegister & mask)
		default:
			return false, ErrInSourceInvalid
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

var ErrOutDestinationInvalid = errors.New("out destination invalid")

func (sm *StateMachine) executeOut(destination outDestination, count uint) (jumped bool, stalled bool, err error) {
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
		return false, false, ErrOutDestinationInvalid
	}

	return jumped, false, nil
}

func (sm *StateMachine) executePushOrPull(isPull bool, ifThreshold bool, block bool) (stalled bool, err error) {
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
				return false, ErrXRegisterNotInitialized
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

var ErrMoveSourceInvalid = errors.New("move source invalid")
var ErrCannotMoveFromOSRWhileAutopullEnabled = errors.New("cannot move from OSR while autopull enabled")
var ErrMoveOperationInvalid = errors.New("move operation invalid")
var ErrMoveDestinationInvalid = errors.New("move destination invalid")

func (sm *StateMachine) executeMove(destination moveDestination, operation moveOperation, source moveSource) (jumped bool, err error) {
	var sourceData uint32
	switch source {
	case moveSourcePins:
		sourceData = bits.RotateLeft32(sm.pinInputs, -int(sm.baseInPin))
	case moveSourceX:
		if !sm.xRegisterInitialized {
			return false, ErrXRegisterNotInitialized
		}
		sourceData = sm.xRegister
	case moveSourceY:
		if !sm.yRegisterInitialized {
			return false, ErrYRegisterNotInitialized
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
			return false, ErrCannotMoveFromOSRWhileAutopullEnabled
		} else {
			sourceData = sm.outputShiftRegister
		}
	default:
		return false, ErrMoveSourceInvalid
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
		return false, ErrMoveOperationInvalid
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
		return false, ErrMoveDestinationInvalid
	}

	return jumped, nil
}

var ErrCannotBothClearAndWaitOnAnIRQ = errors.New("cannot both clear and wait on an IRQ")

func (sm *StateMachine) executeIRQ(clearIRQ bool, waitIRQ bool, index uint, currentlyStalled bool) (stalled bool, err error) {
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
			return false, ErrCannotBothClearAndWaitOnAnIRQ
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

var ErrSetDestinationInvalid = errors.New("set destination invalid")

func (sm *StateMachine) executeSet(destination setDestination, data uint) error {
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
		return ErrSetDestinationInvalid
	}
	return nil
}
