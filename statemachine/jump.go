package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instruction"
)

type jumpInput struct {
	jumpPin                    uint8
	outputShiftRegisterCounter uint8
	pinInputs                  uint32
	pullThreshold              uint8
	xRegister                  uint32
	xRegisterInitialized       bool
	yRegister                  uint32
	yRegisterInitialized       bool
}

type jumpOutput struct {
	decrementXRegister bool
	decrementYRegister bool
	jumped             bool
	newProgramCounter  uint8
}

var ErrJumpConditionInvalid = errors.New("jump condition invalid")

func executeJump(instr instruction.Jump, input jumpInput) (jumpOutput, error) {
	var shouldJump bool
	var output jumpOutput

	address := instr.Address()

	switch instr.Condition() {
	case instruction.JumpAlways:
		shouldJump = true
	case instruction.JumpXZero:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.xRegister == 0)
	case instruction.JumpXNonZeroThenDecrement:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.xRegister != 0)
		output.decrementXRegister = true
	case instruction.JumpYZero:
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.yRegister == 0)
	case instruction.JumpYNonZeroThenDecrement:
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.yRegister != 0)
		output.decrementYRegister = true
	case instruction.JumpXNotEqualY:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.xRegister != input.yRegister)
	case instruction.JumpPin:
		shouldJump = (input.pinInputs>>input.jumpPin)&0b1 == 1
	case instruction.JumpOSRENotEmpty:
		shouldJump = input.outputShiftRegisterCounter < input.pullThreshold
	default:
		return output, ErrJumpConditionInvalid
	}

	if shouldJump {
		output.jumped = true
		output.newProgramCounter = address
	}

	return output, nil
}
