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
	NewPinOutputEnables     uint32
	NewPinOutputEnablesMask uint32
	NewPinOutputs           uint32
	NewPinOutputsMask       uint32
	WrittenToXRegister      bool
	NewXRegister            uint32
	WrittenToYRegister      bool
	NewYRegister            uint32
}

var ErrSetDestinationInvalid = errors.New("set destination invalid")

func ExecuteSet(instr instruction.Set, input SetInput) (SetOutput, error) {
	var output SetOutput

	var pinData = bits.RotateLeft32(uint32(instr.Data()), int(input.BaseSetPin))
	var pinMask = bits.RotateLeft32((0b1<<input.PinCountSet)-1, int(input.BaseSetPin))
	switch instr.Destination() {
	case instruction.SetDestinationPins:
		output.NewPinOutputs = pinData
		output.NewPinOutputsMask = pinMask
	case instruction.SetDestinationX:
		output.WrittenToXRegister = true
		output.NewXRegister = uint32(instr.Data())
	case instruction.SetDestinationY:
		output.WrittenToYRegister = true
		output.NewYRegister = uint32(instr.Data())
	case instruction.SetDestinationPinDirections:
		output.NewPinOutputEnables = pinData
		output.NewPinOutputEnablesMask = pinMask
	default:
		return output, ErrSetDestinationInvalid
	}

	return output, nil
}
