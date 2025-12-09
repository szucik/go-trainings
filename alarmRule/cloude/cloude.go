package alarmrule

//
//import (
//	"fmt"
//	"regexp"
//	"strings"
//)
//
//// TransformAlarmRule converts AWS CloudWatch AlarmRule syntax to govaluate-compatible format.
//// It transforms alarm function calls, boolean operators, and literals.
//func TransformAlarmRule(input string) string {
//	input = strings.TrimSpace(input)
//
//	// Transform alarm function calls first
//	input = transformAlarmCalls(input)
//
//	// Normalize operators - add spaces if missing
//	input = regexp.MustCompile(`\)AND\b`).ReplaceAllString(input, ") AND")
//	input = regexp.MustCompile(`\)OR\b`).ReplaceAllString(input, ") OR")
//	input = regexp.MustCompile(`\)NOT\b`).ReplaceAllString(input, ") NOT")
//
//	input = regexp.MustCompile(`\bAND\(`).ReplaceAllString(input, "AND (")
//	input = regexp.MustCompile(`\bOR\(`).ReplaceAllString(input, "OR (")
//	input = regexp.MustCompile(`\bNOT\(`).ReplaceAllString(input, "NOT (")
//
//	// Transform boolean literals
//	input = regexp.MustCompile(`\bTRUE\b`).ReplaceAllString(input, "true")
//	input = regexp.MustCompile(`\bFALSE\b`).ReplaceAllString(input, "false")
//
//	// Transform operators
//	input = regexp.MustCompile(`\bNOT\s+`).ReplaceAllString(input, "!")
//	input = regexp.MustCompile(`\s+AND\s+`).ReplaceAllString(input, " && ")
//	input = regexp.MustCompile(`\s+OR\s+`).ReplaceAllString(input, " || ")
//
//	return input
//}
//
//// transformAlarmCalls finds and transforms all alarm function calls (ALARM, OK, INSUFFICIENT_DATA).
//func transformAlarmCalls(input string) string {
//	result := strings.Builder{}
//	pos := 0
//
//	for pos < len(input) {
//		found := false
//
//		for _, funcName := range []string{"INSUFFICIENT_DATA", "ALARM", "OK"} {
//			if strings.HasPrefix(input[pos:], funcName+"(") {
//				startPos := pos + len(funcName) + 1
//
//				endPos := findFunctionEnd(input, startPos)
//				if endPos == -1 {
//					result.WriteString(input[pos:])
//					return result.String()
//				}
//
//				content := input[startPos:endPos]
//
//				transformedContent, err := transformAlarmContent(content)
//				if err != nil {
//					result.WriteString(input[pos : endPos+1])
//				} else {
//					result.WriteString(funcName)
//					result.WriteByte('(')
//					result.WriteString(transformedContent)
//					result.WriteByte(')')
//				}
//
//				pos = endPos + 1
//				found = true
//				break
//			}
//		}
//
//		if !found {
//			result.WriteByte(input[pos])
//			pos++
//		}
//	}
//
//	return result.String()
//}
//
//// findFunctionEnd finds the closing parenthesis for an alarm function.
//// It handles cases where parentheses appear inside alarm names (e.g., "a)").
//func findFunctionEnd(input string, start int) int {
//	pos := start
//
//	for pos < len(input) {
//		if input[pos] == ')' {
//			nextPos := pos + 1
//
//			// Skip whitespace
//			for nextPos < len(input) && isWhitespace(input[nextPos]) {
//				nextPos++
//			}
//
//			// Check if this is the function's closing parenthesis
//			if nextPos >= len(input) {
//				return pos
//			}
//
//			if input[nextPos] == ')' ||
//				strings.HasPrefix(input[nextPos:], "AND") ||
//				strings.HasPrefix(input[nextPos:], "OR") ||
//				strings.HasPrefix(input[nextPos:], "NOT") {
//				return pos
//			}
//		}
//		pos++
//	}
//
//	return -1
//}
//
//// transformAlarmContent transforms alarm name to govaluate-compatible format.
//// AWS behavior: removes only ONE outer pair of matching quotes.
//// Examples:
////   - ALARM("test") → 'test'
////   - OK("'test'") → '\'test\” (inner quotes preserved and escaped)
////   - OK(test'name) → 'test\'name' (apostrophe escaped)
//func transformAlarmContent(content string) (string, error) {
//	content = strings.TrimSpace(content)
//
//	if content == "" {
//		return "''", nil
//	}
//
//	// Remove only ONE outer pair of quotes (AWS behavior)
//	if len(content) >= 2 && content[0] == '"' && content[len(content)-1] == '"' {
//		content = content[1 : len(content)-1]
//	} else if len(content) >= 2 && content[0] == '\'' && content[len(content)-1] == '\'' {
//		content = content[1 : len(content)-1]
//	}
//
//	// Escape quotes for govaluate
//	content = strings.ReplaceAll(content, `"`, `\"`)
//	content = strings.ReplaceAll(content, `'`, `\'`)
//
//	return fmt.Sprintf("'%s'", content), nil
//}
//
//func isWhitespace(ch byte) bool {
//	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
//}
