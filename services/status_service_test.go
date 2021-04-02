package services_test

import (
	"testing"

	"github.com/att-comdev/jarvis-connector/services"
)

func TestString(t *testing.T) {
	testData := []struct {
		input    services.StatusService
		expected string
		result   string
	}{
		{
			input:    services.StatusFail,
			expected: services.FailString,
		}, {
			input:    services.StatusIrrelevant,
			expected: services.IrrelevantString,
		}, {
			input:    services.StatusRunning,
			expected: services.RunningString,
		}, {
			input:    services.StatusSuccessful,
			expected: services.SuccessfulString,
		}, {
			input:    services.StatusUnset,
			expected: services.UnsetString,
		},
	}

	// Test to make sure the mapping works
	for _, test := range testData {
		test.result = test.input.String()
		if test.expected != test.result {
			t.Errorf(
				"Test input: %d does not produce expected mapping expected: %s result: %s",
				test.input, test.expected, test.result)
		}
	}
}
