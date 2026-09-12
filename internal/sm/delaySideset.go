package sm

import "math/bits"

func delaySidesetUpdate(sidesetIsOptional bool, sidesetBasePin uint, sidesetBitCount uint, data uint16) (pinSidesets uint32, pinSidesetsMask uint32, delaysRemaining uint) {
	var delayMask uint16 = (0b1 << (5 - sidesetBitCount)) - 1
	var sidesetMask = ^delayMask
	if sidesetIsOptional {
		sidesetMask &= 0b01111
	}

	doSideset := (sidesetBitCount > 0) && (!sidesetIsOptional || ((data>>4)&0b1 == 1))
	if doSideset {
		sidesetData := (data & sidesetMask) >> (5 - sidesetBitCount)
		pinCount := sidesetBitCount
		if sidesetIsOptional {
			pinCount -= 1
		}
		var pinData uint32 = bits.RotateLeft32(uint32(sidesetData), int(sidesetBasePin))
		var pinMask uint32 = bits.RotateLeft32((0b1<<pinCount)-1, int(sidesetBasePin))
		pinSidesets = pinData
		pinSidesetsMask = pinMask
	}

	doDelay := (sidesetBitCount < 5)
	if doDelay {
		delayData := (data & delayMask)
		delaysRemaining = uint(delayData)
	}

	return pinSidesets, pinSidesetsMask, delaysRemaining
}
