package statemachine

import (
	"errors"
	"fmt"
	"math/bits"

	"git.edwardh.dev/pio/fifo"
	"git.edwardh.dev/pio/instruction"
)

type OutInput struct {
	AutopullEnabled              bool
	BaseOutPin                   uint8
	OutShiftMovesRight           bool
	OutWriteEnableBitIndex       uint8
	OutWriteEnableUsed           bool
	OutputShiftRegister          uint32
	OutputShiftRegisterCounter   uint8
	PinCountOut                  uint8
	PullThreshold                uint8
	StickyOutSetAssertionEnabled bool

	FIFOTX fifo.Reader
}

type OutOutput struct {
	NewEXECdInstruction bool
	EXECdInstruction    uint16

	Jumped            bool
	NewProgramCounter uint8

	WrittenToInputShiftRegister  bool
	NewInputShiftRegister        uint32
	NewInputShiftRegisterCounter uint8

	NewOutputShiftRegister        uint32
	NewOutputShiftRegisterCounter uint8

	NewPinOutputEnables     uint32
	NewPinOutputEnablesMask uint32
	NewPinOutputs           uint32
	NewPinOutputsMask       uint32
	WrittenToXRegister      bool
	NewXRegister            uint32
	WrittenToYRegister      bool
	NewYRegister            uint32
	ShouldStall             bool
}

var ErrOutDestinationInvalid = errors.New("out destination invalid")

func ExecuteOut(instr instruction.Out, input OutInput) (OutOutput, error) {
	var output OutOutput = OutOutput{
		NewOutputShiftRegister:        input.OutputShiftRegister,
		NewOutputShiftRegisterCounter: input.OutputShiftRegisterCounter,
	}

	shouldPull := input.AutopullEnabled && (input.OutputShiftRegisterCounter >= input.PullThreshold)
	if shouldPull {
		output.ShouldStall = true
		if !input.FIFOTX.IsEmpty() {
			osr, err := input.FIFOTX.Read()
			if err != nil {
				return output, fmt.Errorf("error reading TX fifo for autopull: %w", err)
			}
			output.NewOutputShiftRegister = osr
			output.NewOutputShiftRegisterCounter = 0
		}
		return output, nil
	}

	var data uint32
	var mask uint32 = (0b1 << instr.BitCount()) - 1
	if input.OutShiftMovesRight {
		data = input.OutputShiftRegister & mask
		output.NewOutputShiftRegister >>= instr.BitCount()
	} else {
		data = (input.OutputShiftRegister >> (32 - instr.BitCount()))
		output.NewOutputShiftRegister <<= instr.BitCount()
	}
	output.NewOutputShiftRegisterCounter += instr.BitCount()
	if output.NewOutputShiftRegisterCounter > 32 {
		output.NewOutputShiftRegisterCounter = 32
	}

	isPinOperation := instr.Destination() == instruction.OutDestinationPins || instr.Destination() == instruction.OutDestinationPinDirections

	outWriteEnable := ((data >> input.OutWriteEnableBitIndex) & 0b1) == 1
	writePins := true
	if isPinOperation && input.OutWriteEnableUsed && !outWriteEnable {
		writePins = false
		if input.StickyOutSetAssertionEnabled {
			output.NewPinOutputEnables = 0
			output.NewPinOutputEnablesMask = 0
			output.NewPinOutputs = 0
			output.NewPinOutputsMask = 0
		}
	}

	var pinData = bits.RotateLeft32(data, int(input.BaseOutPin))
	var pinMask = bits.RotateLeft32((0b1<<input.PinCountOut)-1, int(input.BaseOutPin))

	switch instr.Destination() {
	case instruction.OutDestinationPins:
		if writePins {
			output.NewPinOutputs = pinData
			output.NewPinOutputsMask = pinMask
		}
	case instruction.OutDestinationX:
		output.NewXRegister = data
		output.WrittenToXRegister = true
	case instruction.OutDestinationY:
		output.NewYRegister = data
		output.WrittenToYRegister = true
	case instruction.OutDestinationNull:
		// discards data
	case instruction.OutDestinationPinDirections:
		if writePins {
			output.NewPinOutputEnables = pinData
			output.NewPinOutputEnablesMask = pinMask
		}
	case instruction.OutDestinationProgramCounter:
		output.Jumped = true
		output.NewProgramCounter = uint8(data % 32)
	case instruction.OutDestinationInputShiftRegister:
		output.WrittenToInputShiftRegister = true
		output.NewInputShiftRegister = data
		output.NewInputShiftRegisterCounter = instr.BitCount()
	case instruction.OutDestinationEXEC:
		output.NewEXECdInstruction = true
		output.EXECdInstruction = uint16(data)
	default:
		return output, ErrOutDestinationInvalid
	}

	return output, nil
}
