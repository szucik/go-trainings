package main

import (
	"regexp"
	"strings"
)

// Użyj tej jednej funkcji — przetrwa apokalipsę
// Zamienia każdy ALARM(...) na ALARM('nazwa') – zawsze z apostrofami
func NormalizeToSingleQuotes(rule string) string {
	re := regexp.MustCompile(`(?s)ALARM\s*\(\s*(?:\\"((?:\\.|[^\\])*)\\"|"((\\.|[^\\])*)"|([^)\s"'\\]+))\s*\)`)

	return re.ReplaceAllStringFunc(rule, func(m string) string {
		caps := re.FindStringSubmatch(m)
		name := ""

		if caps[1] != "" {
			name = strings.ReplaceAll(strings.ReplaceAll(caps[1], `\"`, `"`), `\\`, `\`)
		} else if caps[2] != "" {
			name = strings.ReplaceAll(strings.ReplaceAll(caps[2], `\"`, `"`), `\\`, `\`)
		} else if caps[3] != "" {
			name = strings.TrimSpace(caps[3])
		}

		return `ALARM('` + name + `')`
	})
}
