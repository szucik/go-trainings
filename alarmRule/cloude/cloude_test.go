package alarmfmt

import (
	"testing"
)

func TestFormatAlarmRule(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`OK("a")`, `OK('\"a\"')`},
	}

	for _, tt := range tests {
		result, err := FormatAlarmRule(tt.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != tt.expected {
			t.Errorf("FormatAlarmRule(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}
