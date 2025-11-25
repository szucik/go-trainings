package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var statePattern = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA|NOT)\s*\(\s*([^)]+?)\s*\)`)

// GŁÓWNA FUNKCJA – TO JEST TWOJA BRONIA ATOMOWA
func Format(rule string) string {
	s := rule

	// 1. Najpierw zamień wszystkie TRUE/FALSE na małe litery (gdziekolwiek)
	s = regexp.MustCompile(`(?i)\b(TRUE|FALSE)\b`).ReplaceAllString(s, func(m string) string {
		return strings.ToLower(m)
	})

	// 2. Teraz zamień KAŻDE wystąpienie NOT(true) / NOT(false) → !(true) / !(false)
	//    Bez względu na spacje, nawiasy, wielkość liter
	s = regexp.MustCompile(`(?i)NOT\s*\(\s*(true|false)\s*\)`).
		ReplaceAllString(s, "(!$1)")

	s = regexp.MustCompile(`(?i)NOT\s+(true|false)`).
		ReplaceAllString(s, "(!$1)")

	// 3. Obsługa ALARM('name'), OK('name'), itd.
	s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := statePattern.FindStringSubmatch(m)
		if len(caps) < 3 {
			return m
		}
		state := caps[1]
		raw := strings.TrimSpace(caps[2])

		var name string
		if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
			name = strings.Trim(u, `"'`)
		} else {
			name = strings.Trim(strings.ReplaceAll(raw, `\"`, `"`), `"'`)
		}
		name = strings.TrimSpace(name)

		return state + `('` + name + `')`
	})

	// 4. Finalne czyszczenie formatu
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
	s = strings.TrimSpace(s)

	return s
}
