package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instr"
	"git.edwardh.dev/pio/memory"
	"git.edwardh.dev/pio/simulation"
)

type Observer interface {
	Index() uint8

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
	OutWriteEnableBitIndex() uint8
	StickyOutSetAssertionEnabled() bool

	WrapFromAddress() uint8
	WrapToAddress() uint8

	StatusValueUsesRXFIFO() bool
	StatusValueComparisonLevel() uint8
	PullThreshold() uint8
	PushThreshold() uint8
	OutShiftMovesRight() bool
	InShiftMovesRight() bool
	AutopullEnabled() bool
	AutopushEnabled() bool

	BaseSidesetPin() uint8
	BitCountSideset() uint8
	BaseSetPin() uint8
	PinCountSet() uint8
	BaseInPin() uint8
	BaseOutPin() uint8
	PinCountOut() uint8
	JumpPin() uint8

	ProgramCounter() uint8
	Stalled() bool
	DelaysRemaining() uint8
	NewForcedInstruction() bool
	ForcedInstruction() uint16
	ForcedInstructionStalled() bool
	NewEXECdInstruction() bool
	EXECdInstructionStalled() bool
	LatchedInstruction() uint16

	OutputShiftRegister() uint32
	OutputShiftRegisterCounter() uint8
	InputShiftRegister() uint32
	InputShiftRegisterCounter() uint8
	XRegister() uint32
	XRegisterInitialized() bool
	YRegister() uint32
	YRegisterInitialized() bool

	ClockDivisor() float32
	ClockDivisorInteger() uint32
	ClockDivisorFractional() uint8

	ClockDividerTicksRemaining() uint32
	ClockDividerFractionAccumulator() uint8
}

type Configurator interface {
	SetLabel(string)

	FIFORXReader() fifo.Reader
	FIFOTXWriter() fifo.Writer

	ClearFIFOs()
	Init(uint8) error

	Restart()
	SetEnabled(bool)
	SetSidesetIsOptional(bool) error
	SetSidesetControlsPinDirection(bool)
	SetOutWriteEnableUsed(bool)
	SetOutWriteEnableBitIndex(uint8) error
	SetStickyOutSetAssertionEnabled(bool)

	SetWrapFromAddress(uint8) error
	SetWrapToAddress(uint8) error

	SetStatusValueUsesRXFIFO(bool)
	SetStatusValueComparisonLevel(uint8) error
	SetPullThreshold(uint8) error
	SetPushThreshold(uint8) error
	SetOutShiftMovesRight(bool)
	SetInShiftMovesRight(bool)
	SetAutopullEnabled(bool)
	SetAutopushEnabled(bool)

	SetBaseSidesetPin(uint8) error
	SetBitCountSideset(uint8) error
	SetBaseSetPin(uint8) error
	SetPinCountSet(uint8) error
	SetBaseInPin(uint8) error
	SetBaseOutPin(uint8) error
	SetPinCountOut(uint8) error
	SetJumpPin(uint8) error

	SetFIFOJoin(FIFOJoin) error

	ForceInstruction(uint16)

	RestartClockDivider()
	SetClockDivisor(float32) error
	SetClockDivisorInteger(uint32) error
	SetClockDivisorFractional(uint8)

	SetOutShift(bool, bool, uint8) error
	SetInShift(bool, bool, uint8) error
	SetSideset(uint8, bool, bool) error
	SetWrap(uint8, uint8) error
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

	index uint8

	memoryReader memory.Reader
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
	outWriteEnableBitIndex       uint8
	stickyOutSetAssertionEnabled bool

	wrapFromAddress uint8
	wrapToAddress   uint8

	statusValueUsesRXFIFO      bool
	statusValueComparisonLevel uint8
	pullThreshold              uint8
	pushThreshold              uint8
	outShiftMovesRight         bool
	inShiftMovesRight          bool
	autopullEnabled            bool
	autopushEnabled            bool

	baseSidesetPin  uint8
	bitCountSideset uint8
	baseSetPin      uint8
	pinCountSet     uint8
	baseInPin       uint8
	baseOutPin      uint8
	pinCountOut     uint8
	jumpPin         uint8

	programCounter           uint8
	stalled                  bool
	delaysRemaining          uint8
	newForcedInstruction     bool
	forcedInstruction        uint16
	forcedInstructionStalled bool
	newEXECdInstruction      bool
	execdInstructionStalled  bool
	latchedInstruction       uint16

	outputShiftRegister        uint32
	outputShiftRegisterCounter uint8
	inputShiftRegister         uint32
	inputShiftRegisterCounter  uint8
	xRegister                  uint32
	xRegisterInitialized       bool
	yRegister                  uint32
	yRegisterInitialized       bool

	clockDivisorInteger    uint16
	clockDivisorFractional uint8

	clockDividerTicksRemaining      uint32
	clockDividerFractionAccumulator uint8
}

func New(index uint8, label string, memoryReader memory.Reader) *StateMachine {
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

func (sm *StateMachine) FIFORX() *fifo.FIFO {
	return sm.fifoRX
}
func (sm *StateMachine) FIFOTX() *fifo.FIFO {
	return sm.fifoTX
}

func (sm *StateMachine) Observer() Observer {
	return sm
}

func (sm *StateMachine) Index() uint8 { return sm.index }

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
func (sm *StateMachine) OutWriteEnableBitIndex() uint8      { return sm.outWriteEnableBitIndex }
func (sm *StateMachine) StickyOutSetAssertionEnabled() bool { return sm.stickyOutSetAssertionEnabled }

func (sm *StateMachine) WrapFromAddress() uint8 { return sm.wrapFromAddress }
func (sm *StateMachine) WrapToAddress() uint8   { return sm.wrapToAddress }

func (sm *StateMachine) StatusValueUsesRXFIFO() bool       { return sm.statusValueUsesRXFIFO }
func (sm *StateMachine) StatusValueComparisonLevel() uint8 { return sm.statusValueComparisonLevel }
func (sm *StateMachine) PullThreshold() uint8              { return sm.pullThreshold }
func (sm *StateMachine) PushThreshold() uint8              { return sm.pushThreshold }
func (sm *StateMachine) OutShiftMovesRight() bool          { return sm.outShiftMovesRight }
func (sm *StateMachine) InShiftMovesRight() bool           { return sm.inShiftMovesRight }
func (sm *StateMachine) AutopullEnabled() bool             { return sm.autopullEnabled }
func (sm *StateMachine) AutopushEnabled() bool             { return sm.autopushEnabled }

func (sm *StateMachine) BaseSidesetPin() uint8  { return sm.baseSidesetPin }
func (sm *StateMachine) BitCountSideset() uint8 { return sm.bitCountSideset }
func (sm *StateMachine) BaseSetPin() uint8      { return sm.baseSetPin }
func (sm *StateMachine) PinCountSet() uint8     { return sm.pinCountSet }
func (sm *StateMachine) BaseInPin() uint8       { return sm.baseInPin }
func (sm *StateMachine) BaseOutPin() uint8      { return sm.baseOutPin }
func (sm *StateMachine) PinCountOut() uint8     { return sm.pinCountOut }
func (sm *StateMachine) JumpPin() uint8         { return sm.jumpPin }

func (sm *StateMachine) ProgramCounter() uint8          { return sm.programCounter }
func (sm *StateMachine) Stalled() bool                  { return sm.stalled }
func (sm *StateMachine) DelaysRemaining() uint8         { return sm.delaysRemaining }
func (sm *StateMachine) NewForcedInstruction() bool     { return sm.newForcedInstruction }
func (sm *StateMachine) ForcedInstruction() uint16      { return sm.forcedInstruction }
func (sm *StateMachine) ForcedInstructionStalled() bool { return sm.forcedInstructionStalled }
func (sm *StateMachine) NewEXECdInstruction() bool      { return sm.newEXECdInstruction }
func (sm *StateMachine) EXECdInstructionStalled() bool  { return sm.execdInstructionStalled }
func (sm *StateMachine) LatchedInstruction() uint16     { return sm.latchedInstruction }

func (sm *StateMachine) OutputShiftRegister() uint32       { return sm.outputShiftRegister }
func (sm *StateMachine) OutputShiftRegisterCounter() uint8 { return sm.outputShiftRegisterCounter }
func (sm *StateMachine) InputShiftRegister() uint32        { return sm.inputShiftRegister }
func (sm *StateMachine) InputShiftRegisterCounter() uint8  { return sm.inputShiftRegisterCounter }
func (sm *StateMachine) XRegister() uint32                 { return sm.xRegister }
func (sm *StateMachine) XRegisterInitialized() bool        { return sm.xRegisterInitialized }
func (sm *StateMachine) YRegister() uint32                 { return sm.yRegister }
func (sm *StateMachine) YRegisterInitialized() bool        { return sm.yRegisterInitialized }

func (sm *StateMachine) ClockDivisor() float32 {
	return clockDivisorToFloat32(sm.clockDivisorInteger, sm.clockDivisorFractional)
}
func (sm *StateMachine) ClockDivisorInteger() uint32 {
	if sm.clockDivisorInteger == 0 {
		return 65536
	} else {
		return uint32(sm.clockDivisorInteger)
	}
}
func (sm *StateMachine) ClockDivisorFractional() uint8      { return sm.clockDivisorFractional }
func (sm *StateMachine) ClockDividerTicksRemaining() uint32 { return sm.clockDividerTicksRemaining }
func (sm *StateMachine) ClockDividerFractionAccumulator() uint8 {
	return sm.clockDividerFractionAccumulator
}

func (sm *StateMachine) Configurator() Configurator {
	return sm
}

func (sm *StateMachine) SetLabel(label string) { sm.label = label }

func (sm *StateMachine) FIFORXReader() fifo.Reader { return sm.fifoRX.Reader() }
func (sm *StateMachine) FIFOTXWriter() fifo.Writer { return sm.fifoTX.Writer() }

func (sm *StateMachine) ClearFIFOs() {
	sm.fifoRX.Clear()
	sm.fifoTX.Clear()
}

func (sm *StateMachine) Init(programCounter uint8) error {
	if programCounter > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, programCounter)
	}
	sm.SetEnabled(false)
	sm.ClearFIFOs()
	sm.Restart()
	sm.RestartClockDivider()
	sm.programCounter = programCounter
	return nil
}

func (sm *StateMachine) Restart() {
	sm.inputShiftRegisterCounter = 0
	sm.outputShiftRegisterCounter = 32
	sm.inputShiftRegister = 0
	sm.delaysRemaining = 0
	sm.stalled = false
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
func (sm *StateMachine) SetSidesetIsOptional(sidesetIsOptional bool) error {
	if sm.bitCountSideset == 0 && sidesetIsOptional {
		return fmt.Errorf("%w: sideset should not be optional when bit count is 0", ErrValueOutOfRange)
	}
	sm.sidesetIsOptional = sidesetIsOptional
	return nil
}
func (sm *StateMachine) SetSidesetControlsPinDirection(sidesetControlsPinDirection bool) {
	sm.sidesetControlsPinDirection = sidesetControlsPinDirection
}
func (sm *StateMachine) SetOutWriteEnableUsed(outWriteEnableUsed bool) {
	sm.outWriteEnableUsed = outWriteEnableUsed
}

var ErrValueOutOfRange = errors.New("value out of range")

func (sm *StateMachine) SetOutWriteEnableBitIndex(outWriteEnableBitIndex uint8) error {
	if outWriteEnableBitIndex > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, outWriteEnableBitIndex)
	}
	sm.outWriteEnableBitIndex = outWriteEnableBitIndex
	return nil
}
func (sm *StateMachine) SetStickyOutSetAssertionEnabled(stickyOutSetAssertionEnabled bool) {
	sm.stickyOutSetAssertionEnabled = stickyOutSetAssertionEnabled
}
func (sm *StateMachine) SetWrapFromAddress(wrapFromAddress uint8) error {
	if wrapFromAddress > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, wrapFromAddress)
	}
	sm.wrapFromAddress = wrapFromAddress
	return nil
}
func (sm *StateMachine) SetWrapToAddress(wrapToAddress uint8) error {
	if wrapToAddress > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, wrapToAddress)
	}
	sm.wrapToAddress = wrapToAddress
	return nil
}
func (sm *StateMachine) SetStatusValueUsesRXFIFO(statusValueUsesRXFIFO bool) {
	sm.statusValueUsesRXFIFO = statusValueUsesRXFIFO
}
func (sm *StateMachine) SetStatusValueComparisonLevel(statusValueComparisonLevel uint8) error {
	if statusValueComparisonLevel > 15 {
		return fmt.Errorf("%w: %d, should be < 16", ErrValueOutOfRange, statusValueComparisonLevel)
	}
	sm.statusValueComparisonLevel = statusValueComparisonLevel
	return nil
}
func (sm *StateMachine) SetPullThreshold(pullThreshold uint8) error {
	if pullThreshold > 32 || pullThreshold < 1 {
		return fmt.Errorf("%w: %d, should be >0 and <32", ErrValueOutOfRange, pullThreshold)
	}
	sm.pullThreshold = pullThreshold
	return nil
}
func (sm *StateMachine) SetPushThreshold(pushThreshold uint8) error {
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
func (sm *StateMachine) SetBaseSidesetPin(baseSidesetPin uint8) error {
	if baseSidesetPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseSidesetPin)
	}
	sm.baseSidesetPin = baseSidesetPin
	return nil
}
func (sm *StateMachine) SetBitCountSideset(bitCountSideset uint8) error {
	if bitCountSideset > 5 {
		return fmt.Errorf("%w: %d, should be < 6", ErrValueOutOfRange, bitCountSideset)
	}
	if sm.sidesetIsOptional && bitCountSideset == 0 {
		return fmt.Errorf("%w: %d, should be > 0 when sideset is optional", ErrValueOutOfRange, bitCountSideset)
	}
	sm.bitCountSideset = bitCountSideset
	return nil
}
func (sm *StateMachine) SetBaseSetPin(baseSetPin uint8) error {
	if baseSetPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseSetPin)
	}
	sm.baseSetPin = baseSetPin
	return nil
}
func (sm *StateMachine) SetPinCountSet(pinCountSet uint8) error {
	if pinCountSet > 5 {
		return fmt.Errorf("%w: %d, should be < 6", ErrValueOutOfRange, pinCountSet)
	}
	sm.pinCountSet = pinCountSet
	return nil
}
func (sm *StateMachine) SetBaseInPin(baseInPin uint8) error {
	if baseInPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseInPin)
	}
	sm.baseInPin = baseInPin
	return nil
}
func (sm *StateMachine) SetBaseOutPin(baseOutPin uint8) error {
	if baseOutPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, baseOutPin)
	}
	sm.baseOutPin = baseOutPin
	return nil
}
func (sm *StateMachine) SetPinCountOut(pinCountOut uint8) error {
	if pinCountOut > 32 {
		return fmt.Errorf("%w: %d, should be < 33", ErrValueOutOfRange, pinCountOut)
	}
	sm.pinCountOut = pinCountOut
	return nil
}
func (sm *StateMachine) SetJumpPin(jumpPin uint8) error {
	if jumpPin > 31 {
		return fmt.Errorf("%w: %d, should be < 32", ErrValueOutOfRange, jumpPin)
	}
	sm.jumpPin = jumpPin
	return nil
}

type FIFOJoin uint8

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
func (sm *StateMachine) SetClockDivisorInteger(divInt uint32) error {
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

func (sm *StateMachine) SetOutShift(movesRight bool, autopullEnabled bool, pullThreshold uint8) error {
	sm.SetOutShiftMovesRight(movesRight)
	sm.SetAutopullEnabled(autopullEnabled)
	err := sm.SetPullThreshold(pullThreshold)
	if err != nil {
		return fmt.Errorf("error setting pull threshold: %w", err)
	}
	return nil
}

func (sm *StateMachine) SetInShift(movesRight bool, autopushEnabled bool, pushThreshold uint8) error {
	sm.SetInShiftMovesRight(movesRight)
	sm.SetAutopushEnabled(autopushEnabled)
	err := sm.SetPushThreshold(pushThreshold)
	if err != nil {
		return fmt.Errorf("error setting push threshold: %w", err)
	}
	return nil
}

func (sm *StateMachine) SetSideset(bitCountSideset uint8, sidesetIsOptional bool, sidesetControlsPinDirection bool) error {
	if bitCountSideset > 5 {
		return fmt.Errorf("bit count sideset %w: %d, should be < 6", ErrValueOutOfRange, bitCountSideset)
	}
	if sidesetIsOptional && bitCountSideset == 0 {
		return fmt.Errorf("bit count sideset %w: %d, should be > 0 when sideset is optional", ErrValueOutOfRange, bitCountSideset)
	}
	sm.bitCountSideset = bitCountSideset
	sm.sidesetIsOptional = sidesetIsOptional
	sm.SetSidesetControlsPinDirection(sidesetControlsPinDirection)
	return nil
}

func (sm *StateMachine) SetWrap(wrapToAddress uint8, wrapFromAddress uint8) error {
	err := sm.SetWrapToAddress(wrapToAddress)
	if err != nil {
		return fmt.Errorf("error setting wrap to address: %w", err)
	}
	err = sm.SetWrapFromAddress(wrapFromAddress)
	if err != nil {
		return fmt.Errorf("error setting wrap from address: %w", err)
	}
	return nil
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
		jumped, stalled, err := sm.execute(sm.forcedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing forced instruction %016b: %w", sm.forcedInstruction, err)
		}
		sm.newForcedInstruction = false
		if jumped {
			sm.stalled = false
		}
		sm.forcedInstructionStalled = stalled
		if sm.forcedInstructionStalled {
			if sm.newEXECdInstruction || sm.execdInstructionStalled {
				return ErrForcedInstructionStalledDuringEXECdInstruction
			}
			sm.latchedInstruction = sm.forcedInstruction
		}
	} else if sm.forcedInstructionStalled {
		_, stalled, err := sm.execute(sm.latchedInstruction, true)
		if err != nil {
			return fmt.Errorf("error executing stalled forced instruction %016b: %w", sm.latchedInstruction, err)
		}
		sm.forcedInstructionStalled = stalled
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
		sm.newEXECdInstruction = false
		_, stalled, err := sm.execute(sm.latchedInstruction, false)
		if err != nil {
			return fmt.Errorf("error executing EXEC'd instruction %016b: %w", sm.latchedInstruction, err)
		}
		sm.execdInstructionStalled = stalled
	} else if sm.execdInstructionStalled {
		_, stalled, err := sm.execute(sm.latchedInstruction, true)
		if err != nil {
			return fmt.Errorf("error executing stalled EXEC'd instruction %016b: %w", sm.latchedInstruction, err)
		}
		sm.execdInstructionStalled = stalled
	} else if sm.stalled {
		instruction, err := sm.memoryReader.Read(sm.programCounter)
		if err != nil {
			return fmt.Errorf("error reading instruction memory: %w", err)
		}
		jumped, stalled, err := sm.execute(instruction, true)
		if err != nil {
			return fmt.Errorf("error executing stalled instruction %016b: %w", instruction, err)
		}
		if !jumped && !stalled {
			sm.incrementProgramCounter()
		}
		sm.stalled = stalled
	} else if sm.delaysRemaining > 0 {
		err := sm.tryAutopull()
		if err != nil {
			return fmt.Errorf("error performing autopull: %w", err)
		}
		sm.delaysRemaining--
	} else {
		instruction, err := sm.memoryReader.Read(sm.programCounter)
		if err != nil {
			return fmt.Errorf("error reading instruction memory: %w", err)
		}
		jumped, stalled, err := sm.execute(instruction, false)
		if err != nil {
			return fmt.Errorf("error executing instruction %016b at address %d: %w", instruction, sm.programCounter, err)
		}
		if !jumped && !stalled {
			sm.incrementProgramCounter()
		}
		sm.stalled = stalled
	}
	return nil
}

var ErrInstructionTypeInvalid = errors.New("instruction type invalid")

func (sm *StateMachine) execute(instruction uint16, currentlyStalled bool) (jumped bool, stalled bool, err error) {
	switch instr.DecodeType(instruction) {
	case instr.TypeJump:
		decoded := instr.DecodeJump(instruction)
		output, err := executeJump(decoded, jumpInput{
			jumpPin:                    sm.jumpPin,
			outputShiftRegisterCounter: sm.outputShiftRegisterCounter,
			pinInputs:                  sm.pinInputs,
			pullThreshold:              sm.pullThreshold,
			xRegister:                  sm.xRegister,
			xRegisterInitialized:       sm.xRegisterInitialized,
			yRegister:                  sm.yRegister,
			yRegisterInitialized:       sm.yRegisterInitialized,
		})
		if err != nil {
			return false, false, fmt.Errorf("jump: %w", err)
		}
		if output.decrementXRegister {
			sm.xRegister--
		}
		if output.decrementYRegister {
			sm.yRegister--
		}
		if output.jumped {
			jumped = true
			sm.programCounter = output.newProgramCounter
		}
	case instr.TypeWait:
		decoded := instr.DecodeWait(instruction)
		output, err := executeWait(decoded, waitInput{
			baseInPin:         sm.baseInPin,
			irqInputs:         sm.irqInputs,
			pinInputs:         sm.pinInputs,
			stateMachineIndex: sm.index,
		})
		if err != nil {
			return false, false, fmt.Errorf("wait: %w", err)
		}
		if output.clearIRQsMask != 0 {
			sm.irqWrites &= ^output.clearIRQsMask
			sm.irqWritesMask |= output.clearIRQsMask
		}
		if output.shouldStall {
			stalled = true
		}
	case instr.TypeIn:
		decoded := instr.DecodeIn(instruction)
		output, err := executeIn(decoded, inInput{
			alreadyStalled:            currentlyStalled,
			autopushEnabled:           sm.autopushEnabled,
			baseInPin:                 sm.baseInPin,
			inShiftMovesRight:         sm.inShiftMovesRight,
			inputShiftRegister:        sm.inputShiftRegister,
			inputShiftRegisterCounter: sm.inputShiftRegisterCounter,
			outputShiftRegister:       sm.outputShiftRegister,
			pinInputs:                 sm.pinInputs,
			pushThreshold:             sm.pushThreshold,
			xRegister:                 sm.xRegister,
			xRegisterInitialized:      sm.xRegisterInitialized,
			yRegister:                 sm.yRegister,
			yRegisterInitialized:      sm.yRegisterInitialized,

			fifoRX: sm.fifoRX,
		})
		if err != nil {
			return false, false, fmt.Errorf("in: %w", err)
		}
		sm.inputShiftRegister = output.newInputShiftRegister
		sm.inputShiftRegisterCounter = output.newInputShiftRegisterCounter
		if output.shouldStall {
			stalled = true
		}
	case instr.TypeOut:
		decoded := instr.DecodeOut(instruction)
		output, err := executeOut(decoded, outInput{
			autopullEnabled:              sm.autopullEnabled,
			baseOutPin:                   sm.baseOutPin,
			outShiftMovesRight:           sm.outShiftMovesRight,
			outWriteEnableBitIndex:       sm.outWriteEnableBitIndex,
			outWriteEnableUsed:           sm.outWriteEnableUsed,
			outputShiftRegister:          sm.outputShiftRegister,
			outputShiftRegisterCounter:   sm.outputShiftRegisterCounter,
			pinCountOut:                  sm.pinCountOut,
			pullThreshold:                sm.pullThreshold,
			stickyOutSetAssertionEnabled: sm.stickyOutSetAssertionEnabled,

			fifoTX: sm.fifoTX,
		})
		if err != nil {
			return false, false, fmt.Errorf("out: %w", err)
		}
		if output.newEXECdInstruction {
			sm.newEXECdInstruction = true
			sm.latchedInstruction = output.execdInstruction
		}
		if output.jumped {
			jumped = true
			sm.programCounter = output.newProgramCounter
		}
		if output.writtenToInputShiftRegister {
			sm.inputShiftRegister = output.newInputShiftRegister
			sm.inputShiftRegisterCounter = output.newInputShiftRegisterCounter
		}
		sm.outputShiftRegister = output.newOutputShiftRegister
		sm.outputShiftRegisterCounter = output.newOutputShiftRegisterCounter
		if output.newPinOutputEnablesMask != 0 {
			sm.pinOutputEnables &= ^output.newPinOutputEnablesMask
			sm.pinOutputEnables |= output.newPinOutputEnables
			sm.pinOutputEnablesMask |= output.newPinOutputEnablesMask
		}
		if output.newPinOutputsMask != 0 {
			sm.pinOutputs &= ^output.newPinOutputsMask
			sm.pinOutputs |= output.newPinOutputs
			sm.pinOutputsMask |= output.newPinOutputsMask
		}
		if output.writtenToXRegister {
			sm.xRegister = output.newXRegister
			sm.xRegisterInitialized = true
		}
		if output.writtenToYRegister {
			sm.yRegister = output.newYRegister
			sm.yRegisterInitialized = true
		}
		if output.shouldStall {
			stalled = true
		}
	case instr.TypePushPull:
		decoded := instr.DecodePushPull(instruction)
		output, err := executePushPull(decoded, pushPullInput{
			autopullEnabled:            sm.autopullEnabled,
			inputShiftRegister:         sm.inputShiftRegister,
			inputShiftRegisterCounter:  sm.inputShiftRegisterCounter,
			outputShiftRegister:        sm.outputShiftRegister,
			outputShiftRegisterCounter: sm.outputShiftRegisterCounter,
			pullThreshold:              sm.pullThreshold,
			pushThreshold:              sm.pushThreshold,
			xRegister:                  sm.xRegister,
			xRegisterInitialized:       sm.xRegisterInitialized,

			fifoTX: sm.fifoTX,
			fifoRX: sm.fifoRX,
		})
		if err != nil {
			return false, false, fmt.Errorf("push/pull: %w", err)
		}
		sm.inputShiftRegister = output.newInputShiftRegister
		sm.inputShiftRegisterCounter = output.newInputShiftRegisterCounter
		sm.outputShiftRegister = output.newOutputShiftRegister
		sm.outputShiftRegisterCounter = output.newOutputShiftRegisterCounter
		if output.shouldStall {
			stalled = true
		}
	case instr.TypeMove:
		decoded := instr.DecodeMove(instruction)
		output, err := executeMove(decoded, moveInput{
			autopullEnabled:              sm.autopullEnabled,
			baseInPin:                    sm.baseInPin,
			baseOutPin:                   sm.baseOutPin,
			inputShiftRegister:           sm.inputShiftRegister,
			outWriteEnableBitIndex:       sm.outWriteEnableBitIndex,
			outWriteEnableUsed:           sm.outWriteEnableUsed,
			outputShiftRegister:          sm.outputShiftRegister,
			pinCountOut:                  sm.pinCountOut,
			pinInputs:                    sm.pinInputs,
			statusValueComparisonLevel:   sm.statusValueComparisonLevel,
			statusValueUsesRXFIFO:        sm.statusValueUsesRXFIFO,
			stickyOutSetAssertionEnabled: sm.stickyOutSetAssertionEnabled,
			xRegister:                    sm.xRegister,
			xRegisterInitialized:         sm.xRegisterInitialized,
			yRegister:                    sm.yRegister,
			yRegisterInitialized:         sm.yRegisterInitialized,

			fifoRX: sm.fifoRX,
			fifoTX: sm.fifoTX,
		})
		if err != nil {
			return false, false, fmt.Errorf("move: %w", err)
		}
		if output.newEXECdInstruction {
			sm.newEXECdInstruction = true
			sm.latchedInstruction = output.execdInstruction
		}
		if output.jumped {
			jumped = true
			sm.programCounter = output.newProgramCounter
		}
		if output.writtenToInputShiftRegister {
			sm.inputShiftRegister = output.newInputShiftRegister
			sm.inputShiftRegisterCounter = output.newInputShiftRegisterCounter
		}
		if output.writtenToOutputShiftRegister {
			sm.outputShiftRegister = output.newOutputShiftRegister
			sm.outputShiftRegisterCounter = output.newOutputShiftRegisterCounter
		}
		if output.newPinOutputEnablesMask != 0 {
			sm.pinOutputEnables &= ^output.newPinOutputEnablesMask
			sm.pinOutputEnables |= output.newPinOutputEnables
			sm.pinOutputEnablesMask |= output.newPinOutputEnablesMask
		}
		if output.newPinOutputsMask != 0 {
			sm.pinOutputs &= ^output.newPinOutputsMask
			sm.pinOutputs |= output.newPinOutputs
			sm.pinOutputsMask |= output.newPinOutputsMask
		}
		if output.writtenToXRegister {
			sm.xRegister = output.newXRegister
			sm.xRegisterInitialized = true
		}
		if output.writtenToYRegister {
			sm.yRegister = output.newYRegister
			sm.yRegisterInitialized = true
		}
	case instr.TypeIRQ:
		decoded := instr.DecodeIRQ(instruction)
		output, err := executeIRQ(decoded, irqInput{
			alreadyStalled:    currentlyStalled,
			irqInputs:         sm.irqInputs,
			stateMachineIndex: sm.index,
		})
		if err != nil {
			return false, false, fmt.Errorf("irq: %w", err)
		}
		if output.clearIRQsMask != 0 {
			sm.irqWrites &= ^output.clearIRQsMask
			sm.irqWritesMask |= output.clearIRQsMask
		}
		if output.setIRQsMask != 0 {
			sm.irqWrites |= output.setIRQsMask
			sm.irqWritesMask |= output.setIRQsMask
		}
		if output.shouldStall {
			stalled = true
		}
	case instr.TypeSet:
		decoded := instr.DecodeSet(instruction)
		output, err := executeSet(decoded, setInput{
			baseSetPin:  sm.baseSetPin,
			pinCountSet: sm.pinCountSet,
		})
		if err != nil {
			return false, false, fmt.Errorf("set: %w", err)
		}
		if output.newPinOutputEnablesMask != 0 {
			sm.pinOutputEnables &= ^output.newPinOutputEnablesMask
			sm.pinOutputEnables |= output.newPinOutputEnables
			sm.pinOutputEnablesMask |= output.newPinOutputEnablesMask
		}
		if output.newPinOutputsMask != 0 {
			sm.pinOutputs &= ^output.newPinOutputsMask
			sm.pinOutputs |= output.newPinOutputs
			sm.pinOutputsMask |= output.newPinOutputsMask
		}
		if output.writtenToXRegister {
			sm.xRegister = output.newXRegister
			sm.xRegisterInitialized = true
		}
		if output.writtenToYRegister {
			sm.yRegister = output.newYRegister
			sm.yRegisterInitialized = true
		}
	default:
		return false, false, ErrInstructionTypeInvalid
	}

	err = sm.tryAutopull()
	if err != nil {
		return false, false, fmt.Errorf("error performing autopull: %w", err)
	}

	if sm.bitCountSideset > 0 {
		pinCountSideset := sm.bitCountSideset
		if sm.sidesetIsOptional {
			pinCountSideset -= 1
		}
		decodedSideSet := instr.DecodeSideset(instruction, pinCountSideset, sm.sidesetIsOptional)
		if decodedSideSet.Active() {
			sidesetData := decodedSideSet.Data()
			sm.pinSidesets = bits.RotateLeft32(uint32(sidesetData), int(sm.baseSidesetPin))
			sm.pinSidesetsMask = bits.RotateLeft32((0b1<<pinCountSideset)-1, int(sm.baseSidesetPin))
		}
	}

	if !sm.newForcedInstruction && !sm.forcedInstructionStalled {
		if sm.bitCountSideset < 5 {
			decodedDelay := instr.DecodeDelay(instruction, sm.bitCountSideset)
			sm.delaysRemaining = decodedDelay.Cycles()
		} else {
			sm.delaysRemaining = 0
		}
	}

	return jumped, stalled, nil
}

func (sm *StateMachine) incrementProgramCounter() {
	if sm.programCounter == sm.wrapFromAddress {
		sm.programCounter = sm.wrapToAddress
	} else {
		sm.programCounter = (sm.programCounter + 1) % 32
	}
}

func (sm *StateMachine) tryAutopull() error {
	if !sm.autopullEnabled || sm.outputShiftRegisterCounter < sm.pullThreshold || sm.fifoTX.IsEmpty() {
		return nil
	}

	osr, err := sm.fifoTX.Read()
	if err != nil {
		return fmt.Errorf("error reading TX FIFO: %w", err)
	}
	sm.outputShiftRegister = osr
	sm.outputShiftRegisterCounter = 0
	return nil
}

var ErrXRegisterNotInitialized = errors.New("x register not initialized")
var ErrYRegisterNotInitialized = errors.New("y register not initialized")
