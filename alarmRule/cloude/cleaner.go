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
func (c *Cleaner) Clean(input string) string {
	result := strings.TrimSpace(input)
	result = c.normalizeWhitespace(result)
	result = c.cleanFunctionCalls(result)
	return result
}

// normalizeWhitespace removes newlines, tabs, carriage returns
// and normalizes multiple spaces to single space.
// Also removes spaces right after ( and right before ).
func (c *Cleaner) normalizeWhitespace(input string) string {
	result := strings.Builder{}
	lastWasSpace := false
	prevChar := byte(0)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		if isWhitespace(ch) {
			// Look ahead for next non-whitespace char
			nextChar := byte(0)
			for j := i + 1; j < len(input); j++ {
				if !isWhitespace(input[j]) {
					nextChar = input[j]
					break
				}
			}

			// Skip space if:
			// - Last char was already space
			// - Previous char was (
			// - Next char is )
			if lastWasSpace || prevChar == '(' || nextChar == ')' {
				continue
			}

			result.WriteByte(' ')
			lastWasSpace = true
		} else {
			result.WriteByte(ch)
			lastWasSpace = false
			prevChar = ch
		}
	}

	return result.String()
}

// cleanFunctionCalls removes whitespace inside function parentheses.
func (c *Cleaner) cleanFunctionCalls(input string) string {
	result := strings.Builder{}
	pos := 0

	functionNames := []string{"INSUFFICIENT_DATA", "ALARM", "OK"}

	for pos < len(input) {
		found := false

		for _, funcName := range functionNames {
			if strings.HasPrefix(input[pos:], funcName+"(") {
				startPos := pos + len(funcName) + 1

				endPos := findFunctionEnd(input, startPos)
				if endPos == -1 {
					result.WriteString(input[pos:])
					return result.String()
				}

				content := input[startPos:endPos]
				cleanedContent := c.cleanContent(content)

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
func (c *Cleaner) cleanContent(content string) string {
	start := 0
	for start < len(content) && isWhitespace(content[start]) {
		start++
	}

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
func findFunctionEnd(input string, start int) int {
	pos := start

	for pos < len(input) {
		if input[pos] == ')' {
			nextPos := pos + 1

			for nextPos < len(input) && isWhitespace(input[nextPos]) {
				nextPos++
			}

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
