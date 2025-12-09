package alarmrule

import (
	"fmt"
	"regexp"
	"strings"
)

// TransformAlarmRule converts AWS CloudWatch AlarmRule syntax to govaluate-compatible format.
// It first cleans the input using Cleaner, then transforms quotes and operators.
func TransformAlarmRule(input string) string {
	// KROK 1: Wyczyść whitespace (używamy Cleaner!)
	cleaner := NewCleaner()
	input = cleaner.Clean(input)

	// KROK 2: Transformuj cudzysłowy i operatory
	input = transformAlarmCalls(input)
	input = normalizeOperators(input)
	input = transformBooleanLiterals(input)
	input = replaceOperators(input)

	return input
}

// transformAlarmCalls finds and transforms all alarm function calls.
// Changes double quotes to single quotes and escapes content.
func transformAlarmCalls(input string) string {
	result := strings.Builder{}
	pos := 0

	for pos < len(input) {
		found := false

		for _, funcName := range []string{"INSUFFICIENT_DATA", "ALARM", "OK"} {
			if strings.HasPrefix(input[pos:], funcName+"(") {
				startPos := pos + len(funcName) + 1

				endPos := findFunctionEnd(input, startPos)
				if endPos == -1 {
					result.WriteString(input[pos:])
					return result.String()
				}

				content := input[startPos:endPos]

				transformedContent, err := transformAlarmContent(content)
				if err != nil {
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

// transformAlarmContent transforms alarm name to govaluate-compatible format.
// Changes double quotes to single quotes and escapes apostrophes.
//
// Examples:
//   - "test" → 'test'
//   - "'test'" → '\'test\”
//   - test'name → 'test\'name'
func transformAlarmContent(content string) (string, error) {
	content = strings.TrimSpace(content)

	if content == "" {
		return "''", nil
	}

	// Remove only ONE outer pair of quotes (AWS behavior)
	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
		content = content[1 : len(content)-1]
	} else if len(content) >= 2 && content[0] == '\'' && content[len(content)-1] == '\'' {
		content = content[1 : len(content)-1]
	}

	// Escape quotes for govaluate
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, `'`, `\'`)

	return fmt.Sprintf("'%s'", content), nil
}

// normalizeOperators adds spaces around operators if missing.
func normalizeOperators(input string) string {
	input = regexp.MustCompile(`\)AND\b`).ReplaceAllString(input, ") AND")
	input = regexp.MustCompile(`\)OR\b`).ReplaceAllString(input, ") OR")
	input = regexp.MustCompile(`\)NOT\b`).ReplaceAllString(input, ") NOT")
	input = regexp.MustCompile(`\bAND\(`).ReplaceAllString(input, "AND (")
	input = regexp.MustCompile(`\bOR\(`).ReplaceAllString(input, "OR (")
	input = regexp.MustCompile(`\bNOT\(`).ReplaceAllString(input, "NOT (")
	return input
}

// transformBooleanLiterals converts TRUE/FALSE to lowercase.
func transformBooleanLiterals(input string) string {
	input = regexp.MustCompile(`\bTRUE\b`).ReplaceAllString(input, "true")
	input = regexp.MustCompile(`\bFALSE\b`).ReplaceAllString(input, "false")
	return input
}

// replaceOperators converts AWS operators to govaluate operators.
func replaceOperators(input string) string {
	input = regexp.MustCompile(`\bNOT\s+`).ReplaceAllString(input, "!")
	input = regexp.MustCompile(`\s+AND\s+`).ReplaceAllString(input, " && ")
	input = regexp.MustCompile(`\s+OR\s+`).ReplaceAllString(input, " || ")
	return input
}
