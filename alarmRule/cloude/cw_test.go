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
			name:     "ALARM with single space",
			input:    "ALARM (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "ALARM with multiple spaces",
			input:    "ALARM   (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "ALARM with tab",
			input:    "ALARM\t(cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "ALARM with newline",
			input:    "ALARM\n(cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "OK with single space",
			input:    "OK (health)",
			expected: "OK(health)",
		},
		{
			name:     "OK with multiple spaces",
			input:    "OK    (health)",
			expected: "OK(health)",
		},
		{
			name:     "INSUFFICIENT_DATA with space",
			input:    "INSUFFICIENT_DATA (metrics)",
			expected: "INSUFFICIENT_DATA(metrics)",
		},
		{
			name:     "INSUFFICIENT_DATA with tab and spaces",
			input:    "INSUFFICIENT_DATA \t (metrics)",
			expected: "INSUFFICIENT_DATA(metrics)",
		},
		{
			name:     "mixed functions with spaces",
			input:    "ALARM (cpu) AND OK (health)",
			expected: "ALARM(cpu) AND OK(health)",
		},
		{
			name:     "all three functions with spaces",
			input:    "ALARM (cpu) OR OK (health) AND INSUFFICIENT_DATA (metrics)",
			expected: "ALARM(cpu) OR OK(health) AND INSUFFICIENT_DATA(metrics)",
		},
		{
			name:     "spaces inside parentheses preserved",
			input:    "ALARM( cpu )",
			expected: "ALARM( cpu )",
		},
		{
			name:     "spaces inside and outside parentheses",
			input:    "ALARM ( cpu )",
			expected: "ALARM( cpu )",
		},
		{
			name:     "no spaces - unchanged",
			input:    "ALARM(cpu) AND OK(health)",
			expected: "ALARM(cpu) AND OK(health)",
		},
		{
			name:     "negated ALARM with space",
			input:    "NOT ALARM (cpu)",
			expected: "NOT ALARM(cpu)",
		},
		{
			name:     "negated OK with space",
			input:    "!OK (health)",
			expected: "!OK(health)",
		},
		{
			name:     "complex rule with multiple spaces",
			input:    "ALARM  (cpu) OR OK   (health) AND NOT INSUFFICIENT_DATA  (metrics)",
			expected: "ALARM(cpu) OR OK(health) AND NOT INSUFFICIENT_DATA(metrics)",
		},
		{
			name:     "word boundary test - should NOT match MALARM",
			input:    "MALARM (test)",
			expected: "MALARM (test)",
		},
		{
			name:     "word boundary test - should NOT match xALARM",
			input:    "xALARM (test)",
			expected: "xALARM (test)",
		},
		{
			name:     "word boundary test - should NOT match ALARMx",
			input:    "ALARMx (test)",
			expected: "ALARMx (test)",
		},
		{
			name:     "parentheses at start",
			input:    "ALARM (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "quoted alarm name with spaces",
			input:    `ALARM ("my alarm")`,
			expected: `ALARM("my alarm")`,
		},
		{
			name:     "mixed whitespace types",
			input:    "ALARM \t\n (cpu)",
			expected: "ALARM(cpu)",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "real-world example 1",
			input:    "ALARM (\"test1\") OR ALARM (\"test2\")",
			expected: "ALARM(\"test1\") OR ALARM(\"test2\")",
		},
		{
			name:     "real-world example 2",
			input:    "(ALARM (a) OR ALARM (b)) AND NOT ALARM (maint)",
			expected: "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
		},
		{
			name:     "real-world example 3",
			input:    "OK (health) AND NOT INSUFFICIENT_DATA (metrics)",
			expected: "OK(health) AND NOT INSUFFICIENT_DATA(metrics)",
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

	t.Run("use with Cleaner", func(t *testing.T) {
		input := "ALARM (cpu) AND OK (health)"

		// Step 1: Remove spaces after functions
		step1 := cw.RemoveSpacesAfterFunctions(input)
		assert.Equal(t, "ALARM(cpu) AND OK(health)", step1)

		// Step 2: Clean with Cleaner
		cleaner := NewCleaner()
		result := cleaner.Clean(step1)
		assert.Equal(t, "ALARM(cpu) AND OK(health)", result)
	})

	t.Run("use with TransformAlarmRule", func(t *testing.T) {
		input := "ALARM (\"cpu\") AND OK (\"health\")"

		// Step 1: Remove spaces after functions
		step1 := cw.RemoveSpacesAfterFunctions(input)
		assert.Equal(t, "ALARM(\"cpu\") AND OK(\"health\")", step1)

		// Step 2: Transform
		result := TransformAlarmRule(step1)
		expected := "ALARM('cpu') && OK('health')"
		assert.Equal(t, expected, result)
	})

	t.Run("complex real-world scenario", func(t *testing.T) {
		input := `
			ALARM ("us-east-1-cpu") OR 
			ALARM ("eu-west-1-cpu") AND 
			NOT INSUFFICIENT_DATA ("metrics")
		`

		// Remove spaces after functions
		step1 := cw.RemoveSpacesAfterFunctions(input)

		// Should have no spaces between function names and parentheses
		assert.NotContains(t, step1, "ALARM (")
		assert.NotContains(t, step1, "INSUFFICIENT_DATA (")

		// Transform
		result := TransformAlarmRule(step1)

		// Should be properly transformed
		assert.Contains(t, result, "ALARM('us-east-1-cpu')")
		assert.Contains(t, result, "ALARM('eu-west-1-cpu')")
		assert.Contains(t, result, "!INSUFFICIENT_DATA('metrics')")
	})
}

func TestCW_RemoveSpacesAfterFunctions_Benchmarks(t *testing.T) {
	cw := NewCW()

	// Simple benchmark test
	input := "ALARM (cpu) AND OK (health) OR INSUFFICIENT_DATA (metrics)"

	for i := 0; i < 1000; i++ {
		_ = cw.RemoveSpacesAfterFunctions(input)
	}

}
