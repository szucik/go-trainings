package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	statePattern            = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA|NOT)\s*\(\s*([^)]+?)\s*\)`)
	notWithParensPattern    = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`) // NOT(TRUE), NOT ( TRUE )
	notWithoutParensPattern = regexp.MustCompile(`(?i)NOT\s+(TRUE|FALSE)\b`)         // NOT TRUE, NOT  FALSE
	truePattern             = regexp.MustCompile(`(?i)\bTRUE\b`)
	falsePattern            = regexp.MustCompile(`(?i)\bFALSE\b`)
)

func Format(rule string) string {
	s := rule

	// 1. NOT(TRUE) / NOT(FALSE) / NOT (TRUE) → !(true) / !(false) – ZACHOWUJEMY NAWIAS!
	s = notWithParensPattern.ReplaceAllStringFunc(s, func(m string) string {
		val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
		return "!(" + val + ")"
	})

	// 2. NOT TRUE / NOT FALSE → !true / !false – BEZ NAWIASU!
	s = notWithoutParensPattern.ReplaceAllStringFunc(s, func(m string) string {
		val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
		return "!" + val
	})

	// 3. Zamień pozostałe TRUE/FALSE na małe litery
	s = truePattern.ReplaceAllString(s, "true")
	s = falsePattern.ReplaceAllString(s, "false")

	// 4. Formatuj ALARM('name'), OK('name'), itd.
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

	// 5. Czyść formatowanie
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
	return strings.TrimSpace(s)
}
