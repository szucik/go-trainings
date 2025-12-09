package alarmrule

import (
	"strings"
)

// Cleaner removes whitespace and unnecessary characters from AlarmRule
// without changing quotes or operators.
type Cleaner struct{}

// NewCleaner creates a new Cleaner instance.
func NewCleaner() *Cleaner {
	return &Cleaner{}
}

// Clean removes unnecessary whitespace from AlarmRule.
// It does NOT change:
//   - Double quotes to single quotes
//   - Operators (AND, OR, NOT stay as-is)
//   - Function names
//
// It DOES remove:
//   - Leading/trailing whitespace
//   - Newlines (\n)
//   - Tabs (\t)
//   - Carriage returns (\r)
//   - Multiple consecutive spaces (normalized to single space)
//
// Examples:
//
//	Input:  ALARM(\n"test1"\n)
//	Output: ALARM("test1")
//
//	Input:  ALARM("cpu")  AND   ALARM("mem")
//	Output: ALARM("cpu") AND ALARM("mem")
//
//	Input:  ALARM(  "test"  )
//	Output: ALARM("test")
func (c *Cleaner) Clean(input string) string {
	// 1. Trim leading/trailing whitespace
	result := strings.TrimSpace(input)

	// 2. Remove all types of whitespace inside the string
	result = c.normalizeWhitespace(result)

	// 3. Clean whitespace inside function calls
	result = c.cleanFunctionCalls(result)

	return result
}

// normalizeWhitespace removes newlines, tabs, carriage returns
// and normalizes multiple spaces to single space.
func (c *Cleaner) normalizeWhitespace(input string) string {
	result := strings.Builder{}
	lastWasSpace := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// Check if current char is whitespace
		if isWhitespace(ch) {
			// Skip if last char was already space (avoid multiple spaces)
			if !lastWasSpace {
				result.WriteByte(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteByte(ch)
			lastWasSpace = false
		}
	}

	return result.String()
}

// cleanFunctionCalls removes whitespace inside function parentheses.
//
// Examples:
//
//	ALARM( "test" )  → ALARM("test")
//	ALARM( \n"test"\n )  → ALARM("test")
func (c *Cleaner) cleanFunctionCalls(input string) string {
	result := strings.Builder{}
	pos := 0

	functionNames := []string{"INSUFFICIENT_DATA", "ALARM", "OK"}

	for pos < len(input) {
		found := false

		for _, funcName := range functionNames {
			if strings.HasPrefix(input[pos:], funcName+"(") {
				// Found function call
				startPos := pos + len(funcName) + 1 // Position after "FUNCTION("

				endPos := findFunctionEnd(input, startPos)
				if endPos == -1 {
					// Can't find end, write as-is
					result.WriteString(input[pos:])
					return result.String()
				}

				// Extract content
				content := input[startPos:endPos]

				// Clean content (remove whitespace around it)
				cleanedContent := c.cleanContent(content)

				// Write cleaned function call
				result.WriteString(funcName)
				result.WriteByte('(')
				result.WriteString(cleanedContent)
				result.WriteByte(')')

				pos = endPos + 1
				found = true
				break
			}
		}

		if !found {
			result.WriteByte(input[pos])
			pos++
		}
	}

	return result.String()
}

// cleanContent cleans the content inside function parentheses.
// Removes leading/trailing whitespace but preserves quotes and content.
//
// Examples:
//
//	 "test"   → "test"
//	\n"test"\n → "test"
//	  'test'   → 'test'
//	" a "      → " a "  (preserves spaces inside quotes)
func (c *Cleaner) cleanContent(content string) string {
	// Trim whitespace from start
	start := 0
	for start < len(content) && isWhitespace(content[start]) {
		start++
	}

	// Trim whitespace from end
	end := len(content) - 1
	for end >= start && isWhitespace(content[end]) {
		end--
	}

	if start > end {
		return ""
	}

	return content[start : end+1]
}

// findFunctionEnd finds the closing parenthesis for a function call.
// Handles cases where parentheses appear inside alarm names (e.g., "a)").
func findFunctionEnd(input string, start int) int {
	pos := start

	for pos < len(input) {
		if input[pos] == ')' {
			nextPos := pos + 1

			// Skip whitespace
			for nextPos < len(input) && isWhitespace(input[nextPos]) {
				nextPos++
			}

			// Check if this is the function's closing parenthesis
			if nextPos >= len(input) {
				return pos
			}

			if input[nextPos] == ')' ||
				strings.HasPrefix(input[nextPos:], "AND") ||
				strings.HasPrefix(input[nextPos:], "OR") ||
				strings.HasPrefix(input[nextPos:], "NOT") {
				return pos
			}
		}
		pos++
	}

	return -1
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
