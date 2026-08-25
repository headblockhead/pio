package gpio

import "testing"

func TestOverride(t *testing.T) {
	var nonExistantOverride Override = Override(0)
	nonExistantOverride -= 1

	overrideTests := []struct {
		name           string
		override       Override
		input          bool
		expectedOutput bool
		expectedError  error
	}{
		{name: "none_true", override: OverrideNone, input: true, expectedOutput: true, expectedError: nil},
		{name: "none_false", override: OverrideNone, input: false, expectedOutput: false, expectedError: nil},
		{name: "invert_true", override: OverrideInvert, input: true, expectedOutput: false, expectedError: nil},
		{name: "invert_false", override: OverrideInvert, input: false, expectedOutput: true, expectedError: nil},
		{name: "always0_true", override: OverrideAlways0, input: true, expectedOutput: false, expectedError: nil},
		{name: "always0_false", override: OverrideAlways0, input: false, expectedOutput: false, expectedError: nil},
		{name: "always1_true", override: OverrideAlways1, input: true, expectedOutput: true, expectedError: nil},
		{name: "always1_false", override: OverrideAlways1, input: false, expectedOutput: true, expectedError: nil},
		{name: "error_true", override: nonExistantOverride, input: true, expectedOutput: true, expectedError: ErrOverrideInvalid},
		{name: "error_false", override: nonExistantOverride, input: false, expectedOutput: false, expectedError: ErrOverrideInvalid},
	}

	for _, testCase := range overrideTests {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput, actualError := testCase.override.ApplyTo(testCase.input)
			if testCase.expectedError != actualError {
				t.Fatalf("expected error to be %v, got %v", testCase.expectedError, actualError)
			}
			if testCase.expectedOutput != actualOutput {
				t.Errorf("expected output to be %v, got %v", testCase.expectedOutput, actualOutput)
			}
		})
	}
}
