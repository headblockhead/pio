package gpio

import "testing"

func TestNewGPIO(t *testing.T) {
	gpio := NewGPIO()
	if gpio.outputEnableOverride != OverrideNone {
		t.Errorf("expected outputEnableOverride to be OverrideNone, got %v", gpio.outputEnableOverride)
	}
	if gpio.outputOverride != OverrideNone {
		t.Errorf("expected outputOverride to be OverrideNone, got %v", gpio.outputOverride)
	}
	if gpio.inputOverride != OverrideNone {
		t.Errorf("expected inputOverride to be OverrideNone, got %v", gpio.inputOverride)
	}
	if gpio.function != FunctionNone {
		t.Errorf("expected function to be FunctionNone, got %v", gpio.function)
	}
}
