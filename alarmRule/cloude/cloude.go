package alarmrule

import (
	"fmt"
	"regexp"
	"strings"
)

// TransformAlarmRule converts AWS CloudWatch AlarmRule to govaluate format
func TransformAlarmRule(input string) (string, error) {
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	// Step 1: Transform boolean literals
	input = regexp.MustCompile(`\bTRUE\b`).ReplaceAllString(input, "true")
	input = regexp.MustCompile(`\bFALSE\b`).ReplaceAllString(input, "false")

	// Step 2: Transform operators
	// NOT → ! (remove space after !)
	input = regexp.MustCompile(`\bNOT\s+`).ReplaceAllString(input, "!")

	// AND → &&
	input = regexp.MustCompile(`\s+AND\s+`).ReplaceAllString(input, " && ")

	// OR → ||
	input = regexp.MustCompile(`\s+OR\s+`).ReplaceAllString(input, " || ")

	// Step 3: Transform alarm function calls
	input = transformAlarmCalls(input)

	return input, nil
}

// transformAlarmCalls finds and transforms all alarm function calls
func transformAlarmCalls(input string) string {
	result := strings.Builder{}
	pos := 0

	for pos < len(input) {
		// Try to find next alarm function
		found := false

		for _, funcName := range []string{"INSUFFICIENT_DATA", "ALARM", "OK"} {
			if strings.HasPrefix(input[pos:], funcName+"(") {
				// Found a function call
				startPos := pos + len(funcName) + 1 // after "FUNC("

				// Find the closing paren - look for last ) before next keyword or end
				endPos := findFunctionEnd(input, startPos)
				if endPos == -1 {
					// No closing paren, just copy the rest
					result.WriteString(input[pos:])
					return result.String()
				}

				// Extract content
				content := input[startPos:endPos]

				// Transform content
				transformedContent, err := transformAlarmContent(content)
				if err != nil {
					// On error, keep original
					result.WriteString(input[pos : endPos+1])
				} else {
					result.WriteString(funcName)
					result.WriteByte('(')
					result.WriteString(transformedContent)
					result.WriteByte(')')
				}

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

// findFunctionEnd finds the closing ) for a function
// Strategy: find ) that is followed by whitespace + keyword, another ), or end of string
func findFunctionEnd(input string, start int) int {
	pos := start

	for pos < len(input) {
		if input[pos] == ')' {
			// Found a ), check what comes next
			nextPos := pos + 1

			// Skip whitespace
			for nextPos < len(input) && isWhitespace(input[nextPos]) {
				nextPos++
			}

			// Check if this looks like end of function
			if nextPos >= len(input) {
				// End of string
				return pos
			}

			// Check if next is ) or keyword (already transformed to symbols)
			if input[nextPos] == ')' ||
				strings.HasPrefix(input[nextPos:], "&&") ||
				strings.HasPrefix(input[nextPos:], "||") ||
				strings.HasPrefix(input[nextPos:], "!") {
				return pos
			}
		}
		pos++
	}

	return -1
}

// transformAlarmContent transforms alarm name according to rules
func transformAlarmContent(content string) (string, error) {

	// Trim whitespace
	content = strings.TrimSpace(content)

	if content == "" {
		return "''", nil
	}

	// Step 1: If content starts and ends with ' - remove them
	if len(content) >= 2 && content[0] == '\'' && content[len(content)-1] == '\'' {
		content = content[1 : len(content)-1]
	}

	// Step 2: Escape all " and '
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, `'`, `\'`)

	// Step 3: Wrap in single quotes
	return fmt.Sprintf("'%s'", content), nil
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
