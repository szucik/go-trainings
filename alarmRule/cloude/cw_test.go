package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCW_RemoveSpacesAfterFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "positive ALARM with single space should remove space",
			input:    "ALARM (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "positive ALARM with multiple spaces should remove all spaces",
			input:    "ALARM   (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "positive OK with space should remove space",
			input:    "OK (health)",
			expected: "OK(health)",
		},
		{
			name:     "positive INSUFFICIENT_DATA with space should remove space",
			input:    "INSUFFICIENT_DATA (metrics)",
			expected: "INSUFFICIENT_DATA(metrics)",
		},
		{
			name:     "positive mixed functions with spaces should remove all spaces",
			input:    "ALARM (cpu) AND OK (health)",
			expected: "ALARM(cpu) AND OK(health)",
		},
		{
			name:     "positive spaces inside parentheses should be preserved",
			input:    "ALARM( cpu )",
			expected: "ALARM( cpu )",
		},
		{
			name:     "positive no spaces should remain unchanged",
			input:    "ALARM(cpu) AND OK(health)",
			expected: "ALARM(cpu) AND OK(health)",
		},
		{
			name:     "positive negated ALARM with space should remove space",
			input:    "NOT ALARM (cpu)",
			expected: "NOT ALARM(cpu)",
		},
		{
			name:     "negative MALARM should not match and remain unchanged",
			input:    "MALARM (test)",
			expected: "MALARM (test)",
		},
		{
			name:     "positive real-world example should normalize spacing",
			input:    "(ALARM (a) OR ALARM (b)) AND NOT ALARM (maint)",
			expected: "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := NewCW()
			result := cw.RemoveSpacesAfterFunctions(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCW_RemoveSpacesAfterFunctions_Integration(t *testing.T) {
	cw := NewCW()

	t.Run("positive use with TransformAlarmRule should transform correctly", func(t *testing.T) {
		input := "ALARM (\"cpu\") AND OK (\"health\")"

		// Step 1: Remove spaces after functions
		step1 := cw.RemoveSpacesAfterFunctions(input)
		assert.Equal(t, "ALARM(\"cpu\") AND OK(\"health\")", step1)

		// Step 2: Transform
		result := TransformAlarmRule(step1)
		expected := "ALARM('cpu') && OK('health')"
		assert.Equal(t, expected, result)
	})
}
