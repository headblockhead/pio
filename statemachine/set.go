package statemachine

import (
	"errors"
	"math/bits"

	"git.edwardh.dev/pio/instruction"
)

type SetInput struct {
	BaseSetPin  uint8
	PinCountSet uint8
}

type SetOutput struct {
	PinOutputEnables     uint32
	PinOutputEnablesMask uint32
	PinOutputs           uint32
	PinOutputsMask       uint32
	WrittenToXRegister   bool
	XRegister            uint32
	WrittenToYRegister   bool
	YRegister            uint32
}

var ErrSetDestinationInvalid = errors.New("set destination invalid")

func ExecuteSet(instr instruction.Set, input SetInput) (SetOutput, error) {
	var output SetOutput

	var pinData = bits.RotateLeft32(uint32(instr.Data()), int(input.BaseSetPin))
	var pinMask = bits.RotateLeft32((0b1<<input.PinCountSet)-1, int(input.BaseSetPin))
	switch instr.Destination() {
	case instruction.SetDestinationPins:
		output.PinOutputs = pinData
		output.PinOutputsMask = pinMask
	case instruction.SetDestinationX:
		output.WrittenToXRegister = true
		output.XRegister = uint32(instr.Data())
	case instruction.SetDestinationY:
		output.WrittenToYRegister = true
		output.YRegister = uint32(instr.Data())
	case instruction.SetDestinationPinDirections:
		output.PinOutputEnables = pinData
		output.PinOutputEnablesMask = pinMask
	default:
		return output, ErrSetDestinationInvalid
	}

	return output, nil
}
