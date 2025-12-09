package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleaner_Clean(t *testing.T) {
	cleaner := NewCleaner()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple with newlines",
			input: "ALARM(\n\"test1\"\n)",
			want:  `ALARM("test1")`,
		},
		{
			name:  "Multiple spaces",
			input: "ALARM(\"cpu\")  AND   ALARM(\"mem\")",
			want:  `ALARM("cpu") AND ALARM("mem")`,
		},
		{
			name:  "Tabs and newlines",
			input: "ALARM(\t\"test\"\t)",
			want:  `ALARM("test")`,
		},
		{
			name:  "Leading/trailing whitespace",
			input: "  ALARM(\"test\")  ",
			want:  `ALARM("test")`,
		},
		{
			name:  "Newlines everywhere",
			input: "ALARM(\n\"cpu\"\n)\nAND\nALARM(\n\"mem\"\n)",
			want:  `ALARM("cpu") AND ALARM("mem")`,
		},
		{
			name:  "Single quotes preserved",
			input: "OK(\n'test'\n)",
			want:  `OK('test')`,
		},
		{
			name:  "Spaces inside quotes preserved",
			input: `ALARM(" a ")`,
			want:  `ALARM(" a ")`,
		},
		{
			name:  "Complex expression",
			input: "(\nALARM(\"cpu\")\nOR\nALARM(\"mem\")\n)\nAND\nNOT\nALARM(\"maintenance\")",
			want:  `(ALARM("cpu") OR ALARM("mem")) AND NOT ALARM("maintenance")`,
		},
		{
			name:  "INSUFFICIENT_DATA with newlines",
			input: "INSUFFICIENT_DATA(\n\"metric\"\n)",
			want:  `INSUFFICIENT_DATA("metric")`,
		},
		{
			name:  "Multiple newlines",
			input: "ALARM(\n\n\"test\"\n\n)",
			want:  `ALARM("test")`,
		},
		{
			name:  "Mixed whitespace",
			input: "ALARM( \n\t \"test\" \t\n )",
			want:  `ALARM("test")`,
		},
		{
			name:  "No quotes",
			input: "ALARM(\ntest\n)",
			want:  `ALARM(test)`,
		},
		{
			name:  "Parenthesis in name",
			input: "ALARM(\n\"a)\"\n)",
			want:  `ALARM("a)")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleaner.Clean(tt.input)
			assert.Equal(t, tt.want, got, "Input: %q", tt.input)
		})
	}
}

func TestCleaner_CleanContent(t *testing.T) {
	cleaner := NewCleaner()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "With newlines",
			input: "\n\"test\"\n",
			want:  `"test"`,
		},
		{
			name:  "With tabs",
			input: "\t\"test\"\t",
			want:  `"test"`,
		},
		{
			name:  "With spaces",
			input: "  \"test\"  ",
			want:  `"test"`,
		},
		{
			name:  "Mixed whitespace",
			input: " \n\t \"test\" \t\n ",
			want:  `"test"`,
		},
		{
			name:  "No whitespace",
			input: `"test"`,
			want:  `"test"`,
		},
		{
			name:  "Only whitespace",
			input: " \n\t ",
			want:  "",
		},
		{
			name:  "Spaces inside quotes preserved",
			input: `" a "`,
			want:  `" a "`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleaner.cleanContent(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCleaner_NormalizeWhitespace(t *testing.T) {
	cleaner := NewCleaner()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Multiple spaces",
			input: "ALARM    AND    OK",
			want:  "ALARM AND OK",
		},
		{
			name:  "Newlines to spaces",
			input: "ALARM\nAND\nOK",
			want:  "ALARM AND OK",
		},
		{
			name:  "Tabs to spaces",
			input: "ALARM\tAND\tOK",
			want:  "ALARM AND OK",
		},
		{
			name:  "Mixed whitespace",
			input: "ALARM  \n\t  AND  \t\n  OK",
			want:  "ALARM AND OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleaner.normalizeWhitespace(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
