package statemachine

import (
	"errors"
	"math/bits"

	"git.edwardh.dev/pio/instruction"
)

type setInput struct {
	baseSetPin  uint8
	pinCountSet uint8
}

type setOutput struct {
	newPinOutputEnables     uint32
	newPinOutputEnablesMask uint32
	newPinOutputs           uint32
	newPinOutputsMask       uint32
	writtenToXRegister      bool
	newXRegister            uint32
	writtenToYRegister      bool
	newYRegister            uint32
}

var ErrSetDestinationInvalid = errors.New("set destination invalid")

func executeSet(instr instruction.Set, input setInput) (setOutput, error) {
	var output setOutput

	var pinData = bits.RotateLeft32(uint32(instr.Data()), int(input.baseSetPin))
	var pinMask = bits.RotateLeft32((0b1<<input.pinCountSet)-1, int(input.baseSetPin))
	switch instr.Destination() {
	case instruction.SetDestinationPins:
		output.newPinOutputs = pinData
		output.newPinOutputsMask = pinMask
	case instruction.SetDestinationX:
		output.writtenToXRegister = true
		output.newXRegister = uint32(instr.Data())
	case instruction.SetDestinationY:
		output.writtenToYRegister = true
		output.newYRegister = uint32(instr.Data())
	case instruction.SetDestinationPinDirections:
		output.newPinOutputEnables = pinData
		output.newPinOutputEnablesMask = pinMask
	default:
		return output, ErrSetDestinationInvalid
	}

	return output, nil
}
