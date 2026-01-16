package alarmrule

import (
	"regexp"
)

// CW provides CloudWatch-specific utilities for alarm rule processing
type CW struct{}

// NewCW creates a new CW instance
func NewCW() *CW {
	return &CW{}
}

// RemoveSpacesAfterFunctions removes spaces immediately after CloudWatch function names
// (ALARM, OK, INSUFFICIENT_DATA) before the opening parenthesis.
//
// This is useful for normalizing CloudWatch alarm rules that may have inconsistent spacing.
//
// Examples:
//   - "ALARM (cpu)" → "ALARM(cpu)"
//   - "OK (health)" → "OK(health)"
//   - "INSUFFICIENT_DATA (metrics)" → "INSUFFICIENT_DATA(metrics)"
//   - "ALARM( cpu )" → "ALARM( cpu )" (spaces inside parentheses are preserved)
//   - "ALARM   (cpu)" → "ALARM(cpu)" (multiple spaces removed)
//   - "NOT ALARM (cpu)" → "NOT ALARM(cpu)" (works with negation)
//
// Parameters:
//   - input: The alarm rule string with potential spaces after function names
//
// Returns:
//   - The normalized alarm rule with spaces removed after function names
func (cw *CW) RemoveSpacesAfterFunctions(input string) string {
	result := input

	// ALARM (...) → ALARM(...)
	// \b ensures word boundary (won't match MALARM, xALARM, etc.)
	// \s+ matches one or more whitespace characters (space, tab, newline)
	result = regexp.MustCompile(`\bALARM\s+\(`).ReplaceAllString(result, "ALARM(")

	// OK (...) → OK(...)
	result = regexp.MustCompile(`\bOK\s+\(`).ReplaceAllString(result, "OK(")

	// INSUFFICIENT_DATA (...) → INSUFFICIENT_DATA(...)
	result = regexp.MustCompile(`\bINSUFFICIENT_DATA\s+\(`).ReplaceAllString(result, "INSUFFICIENT_DATA(")

	return result
}
