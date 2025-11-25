package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Dopasowuje: NOT ALARM(x), NOT OK(x), NOT INSUFFICIENT_DATA(x), NOT ( ALARM(x) ), itd.
	notStatePattern  = regexp.MustCompile(`(?i)NOT\s*\(\s*(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)\s*\)`)
	notStateNoParens = regexp.MustCompile(`(?i)NOT\s+(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)

	// Standardowe stany
	statePattern = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)

	// TRUE/FALSE – już mamy
	truePattern  = regexp.MustCompile(`(?i)\bTRUE\b`)
	falsePattern = regexp.MustCompile(`(?i)\bFALSE\b`)
)

func Format(rule string) string {
	s := rule

	// 1. Najpierw zamień NOT(TRUE)/NOT(FALSE) – jak wcześniej
	s = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`).
		ReplaceAllStringFunc(s, func(m string) string {
			val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
			return "!(" + val + ")"
		})
	s = regexp.MustCompile(`(?i)NOT\s+(TRUE|FALSE)\b`).
		ReplaceAllStringFunc(s, func(m string) string {
			val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
			return "!" + val
		})

	// 2. NOT OK(x), NOT ALARM(x) → !OK('x'), !ALARM('x')
	s = notStatePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := notStatePattern.FindStringSubmatch(m)
		state := caps[1]
		inner := strings.TrimSpace(caps[2])
		name := cleanName(inner)
		return "!" + state + `('` + name + `')`
	})

	s = notStateNoParens.ReplaceAllStringFunc(s, func(m string) string {
		caps := notStateNoParens.FindStringSubmatch(m)
		state := caps[1]
		inner := strings.TrimSpace(caps[2])
		name := cleanName(inner)
		return "!" + state + `('` + name + `')`
	})

	// 3. Formatuj zwykłe stany: ALARM(x) → ALARM('x')
	s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := statePattern.FindStringSubmatch(m)
		state := caps[1]
		raw := strings.TrimSpace(caps[2])
		name := cleanName(raw)
		return state + `('` + name + `')`
	})

	// 4. TRUE/FALSE → true/false
	s = truePattern.ReplaceAllString(s, "true")
	s = falsePattern.ReplaceAllString(s, "false")

	// 5. Czyść formatowanie
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
	return strings.TrimSpace(s)
}

// Pomocnicza funkcja – czyści nazwę alarmu
func cleanName(raw string) string {
	if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
		return strings.Trim(u, `"'`)
	}
	name := strings.ReplaceAll(raw, `\"`, `"`)
	return strings.Trim(name, `"'`)
}
