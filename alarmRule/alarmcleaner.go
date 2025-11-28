// alarmcleaner/cleaner.go
package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// NOWE: obsługa NOT ( STATE(...) ) i NOT STATE(...)
	notStateParensPattern = regexp.MustCompile(`(?i)NOT\s*\(\s*(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)\s*\)`)
	notStatePattern       = regexp.MustCompile(`(?i)NOT\s+(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)

	// Twoje stare, dobre wzorce
	statePattern            = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)
	notWithParensPattern    = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`)
	notWithoutParensPattern = regexp.MustCompile(`(?i)NOT\s+(TRUE|FALSE)\b`)
	truePattern             = regexp.MustCompile(`(?i)\bTRUE\b`)
	falsePattern            = regexp.MustCompile(`(?i)\bFALSE\b`)
)

func Format(rule string) string {
	s := rule

	// 1. NAJPIERW: NOT ( STATE(...) ) → !STATE('...')
	s = notStateParensPattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := notStateParensPattern.FindStringSubmatch(m)
		state := strings.ToUpper(caps[1])
		name := cleanName(caps[2])
		return "!" + state + `('` + name + `')`
	})

	// 2. POTEM: NOT STATE(...) → !STATE('...')
	s = notStatePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := notStatePattern.FindStringSubmatch(m)
		state := strings.ToUpper(caps[1])
		name := cleanName(caps[2])
		return "!" + state + `('` + name + `')`
	})

	// 3. NOT(TRUE) / NOT(FALSE) → !(true) / !(false)
	s = notWithParensPattern.ReplaceAllStringFunc(s, func(m string) string {
		val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
		return "!(" + val + ")"
	})

	// 4. NOT TRUE / NOT FALSE → !true / !false
	s = notWithoutParensPattern.ReplaceAllStringFunc(s, func(m string) string {
		val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
		return "!" + val
	})

	// 5. Zwykłe ALARM(...), OK(...) itd.
	s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := statePattern.FindStringSubmatch(m)
		if len(caps) < 3 {
			return m
		}
		state := strings.ToUpper(caps[1])
		name := cleanName(caps[2])
		return state + `('` + name + `')`
	})

	// 6. Luźne TRUE / FALSE → true / false
	s = truePattern.ReplaceAllString(s, "true")
	s = falsePattern.ReplaceAllString(s, "false")

	// 7. Formatowanie
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")

	return strings.TrimSpace(s)
}

// Najlepsza funkcja do czyszczenia nazwy – oparta na Twoim genialnym Unquote
func cleanName(raw string) string {
	trimmed := strings.TrimSpace(raw)

	// Najpierw próba strconv.Unquote – obsługuje "cpu", \"cpu\", \t"cpu "\t itd.
	if u, err := strconv.Unquote(`"` + trimmed + `"`); err == nil {
		return strings.TrimSpace(u)
	}

	// Fallback – ręczne czyszczenie
	s := strings.ReplaceAll(trimmed, `\"`, `"`)
	s = strings.Trim(s, `"`)
	return strings.TrimSpace(s)
}
