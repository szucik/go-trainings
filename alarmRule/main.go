package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Użyj tej jednej funkcji — przetrwa apokalipsę
// Zamienia każdy ALARM(...) na ALARM('nazwa') – zawsze z apostrofami
func FixAlarmRule(rule string) string {
	re := regexp.MustCompile(`ALARM\s*\([^)]*\)`)

	return re.ReplaceAllStringFunc(rule, func(part string) string {
		inner := part
		inner = strings.TrimPrefix(inner, "ALARM")
		inner = strings.Trim(inner, "() \t\n\r")

		// Magia: próbujemy zinterpretować jako string Go
		if s, err := strconv.Unquote(`"` + inner + `"`); err == nil {
			return "ALARM('" + strings.Trim(s, `"`) + "')"
		}

		// Fallback
		s := strings.ReplaceAll(inner, `\"`, `"`)
		s = strings.Trim(s, `"`)
		return "ALARM('" + s + "')"
	})
}
