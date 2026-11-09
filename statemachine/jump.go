package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type JumpInput struct {
	JumpPin                    uint8
	OutputShiftRegisterCounter uint8
	PinInputs                  uint32
	PullThreshold              uint8
	XRegister                  uint32
	XRegisterInitialized       bool
	YRegister                  uint32
	YRegisterInitialized       bool
}

type JumpOutput struct {
	DecrementXRegister bool
	DecrementYRegister bool
	Jumped             bool
	NewProgramCounter  uint8
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
		shouldJump = (input.PinInputs>>input.JumpPin)&0b1 == 1
	case instruction.JumpOSRENotEmpty:
		shouldJump = input.OutputShiftRegisterCounter < input.PullThreshold
	default:
		return output, ErrJumpConditionInvalid
	}

	if shouldJump {
		output.Jumped = true
		output.NewProgramCounter = address
	}

	return output, nil
}
