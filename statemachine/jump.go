package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type JumpInput struct {
	XRegister                   uint32
	XRegisterInitialized        bool
	YRegister                   uint32
	YRegisterInitialized        bool
	JumpPinValue                bool
	OutputShiftRegisterNotEmpty bool
}

type JumpOutput struct {
	DecrementXRegister bool
	DecrementYRegister bool
	NewPC              *uint8
}

var ErrJumpConditionInvalid = errors.New("jump condition invalid")

func ExecuteJump(instr instruction.Jump, input JumpInput) (JumpOutput, error) {
	var shouldJump bool
	var output JumpOutput

	address := instr.Address()

	switch instr.Condition() {
	case instruction.JumpAlways:
		shouldJump = true
	case instruction.JumpXZero:
		if !input.XRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.XRegister == 0)
	case instruction.JumpXNonZeroThenDecrement:
		if !input.XRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.XRegister != 0)
		output.DecrementXRegister = true
	case instruction.JumpYZero:
		if !input.YRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.YRegister == 0)
	case instruction.JumpYNonZeroThenDecrement:
		if !input.YRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.YRegister != 0)
		output.DecrementYRegister = true
	case instruction.JumpXNotEqualY:
		if !input.XRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		if !input.YRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.XRegister != input.YRegister)
	case instruction.JumpPin:
		shouldJump = input.JumpPinValue
	case instruction.JumpOSRENotEmpty:
		shouldJump = input.OutputShiftRegisterNotEmpty
	default:
		return output, ErrJumpConditionInvalid
	}

	if shouldJump {
		output.NewPC = &address
	}

	return output, nil
}
