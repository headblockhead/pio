package sm

import "errors"

var ErrClockDivisorInvalid = errors.New("invalid clock divisor")

func clockDivisorFromFloat32(divisor float32) (divisorInteger uint16, divisorFractional uint8, err error) {
	if divisor < 1 || divisor > 65536 {
		return 0, 0, ErrClockDivisorInvalid
	}
	if divisor == 65536 {
		return 0, 0, nil
	}
	divisorInteger = uint16(divisor)
	divisorFractional = uint8((divisor - float32(divisorInteger)) * 256)
	return divisorInteger, divisorFractional, nil
}

func clockDivisorToFloat32(divisorInteger uint16, divisorFractional uint8) float32 {
	if divisorInteger == 0 {
		return 65536
	}
	return float32(divisorInteger) + (float32(divisorFractional) / 256)
}

func clockDividerUpdateTicksRemaining(dividerFractionAccumulator uint8, divisorInteger uint16, divisorFractional uint8) (newDividerTicksRemaining uint, newDividerFractionAccumulator uint8) {
	var actualDivisorInteger uint
	// clockDivisorInteger as 0 represents a divisor of 65536, as dividing by 0 is not a useful operation.
	if divisorInteger == 0 {
		actualDivisorInteger = 65536
		divisorFractional = 0
	} else {
		actualDivisorInteger = uint(divisorInteger)
	}

	// Setup the amount of ticks to count until the next divided tick based on the integer divisor.
	newDividerTicksRemaining = actualDivisorInteger

	// Now, use the fractionAccumulator to determine whether to delay the next divided clock tick by an additional clock cycle.
	// Importantly, dividerFractionAccumulator is a uint8, so automatically wraps around every 256.
	// As the divisorFractional represents a number of 256ths, this makes calculations very easy.
	newDividerFractionAccumulator = dividerFractionAccumulator + divisorFractional
	// If adding the divisorFractional has caused the fractionAccumulator to wrap,
	// delay the next divided clock tick by an additional clock cycle.
	if newDividerFractionAccumulator < dividerFractionAccumulator {
		newDividerTicksRemaining++
	}

	return newDividerTicksRemaining, newDividerFractionAccumulator
}
