package statemachine

import (
	"errors"

	"git.edwardh.dev/pio/instr"
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

func executeJump(i instr.Jump, input jumpInput) (jumpOutput, error) {
	var shouldJump bool
	var output jumpOutput

	address := i.Address()

	switch i.Condition() {
	case instr.JumpAlways:
		shouldJump = true
	case instr.JumpXZero:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.xRegister == 0)
	case instr.JumpXNonZeroThenDecrement:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		shouldJump = (input.xRegister != 0)
		output.decrementXRegister = true
	case instr.JumpYZero:
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.yRegister == 0)
	case instr.JumpYNonZeroThenDecrement:
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.yRegister != 0)
		output.decrementYRegister = true
	case instr.JumpXNotEqualY:
		if !input.xRegisterInitialized {
			return output, ErrXRegisterNotInitialized
		}
		if !input.yRegisterInitialized {
			return output, ErrYRegisterNotInitialized
		}
		shouldJump = (input.xRegister != input.yRegister)
	case instr.JumpPin:
		shouldJump = (input.pinInputs>>input.jumpPin)&0b1 == 1
	case instr.JumpOSRENotEmpty:
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
