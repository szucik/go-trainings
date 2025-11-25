package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// === TYLKO poprawne, ODDZIELONE spacja lub nawiasem NOT + stan ===
	notStateParens = regexp.MustCompile(`(?i)\bNOT\s*\(\s*(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)\s*\)`)
	notStateSpaced = regexp.MustCompile(`(?i)\bNOT\s+(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)

	// === NOT + true/false ===
	notTrueFalseParens = regexp.MustCompile(`(?i)\bNOT\s*\(\s*(TRUE|FALSE)\s*\)`)
	notTrueFalseSpaced = regexp.MustCompile(`(?i)\bNOT\s+(TRUE|FALSE)\b`)

	// === Zwykłe stany (ALARM, OK, itd.) ===
	statePattern = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)

	truePattern  = regexp.MustCompile(`(?i)\bTRUE\b`)
	falsePattern = regexp.MustCompile(`(?i)\bFALSE\b`)
)

func Format(rule string) string {
	s := rule

	// 1. NOT + stan – tylko gdy NOT jest oddzielone (spacja lub nawias)
	s = notStateParens.ReplaceAllStringFunc(s, replaceNotState)
	s = notStateSpaced.ReplaceAllStringFunc(s, replaceNotState)

	// 2. NOT + true/false
	s = notTrueFalseParens.ReplaceAllStringFunc(s, replaceNotBoolParens)
	s = notTrueFalseSpaced.ReplaceAllStringFunc(s, replaceNotBoolSpaced)

	// 3. Formatuj zwykłe stany
	s = statePattern.ReplaceAllStringFunc(s, formatState)

	// 4. TRUE/FALSE → true/false
	s = truePattern.ReplaceAllString(s, "true")
	s = falsePattern.ReplaceAllString(s, "false")

	// 5. Czyść format
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
	return strings.TrimSpace(s)
}

// Pomocnicze funkcje – żeby nie było inline func w regexie
func replaceNotState(m string) string {
	// Wyciągnij stan i nazwę
	re := regexp.MustCompile(`(?i)(ALARM|OK|INSUFFICIENT_DATA)\s*\(\s*([^)]+?)\s*\)`)
	caps := re.FindStringSubmatch(m)
	if len(caps) < 3 {
		return m
	}
	state := caps[1]
	name := cleanName(caps[2])
	return "!" + state + `('` + name + `')`
}

func replaceNotBoolParens(m string) string {
	val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
	return "!(" + val + ")"
}

func replaceNotBoolSpaced(m string) string {
	val := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m)))
	return "!" + val
}

func formatState(m string) string {
	caps := statePattern.FindStringSubmatch(m)
	if len(caps) < 3 {
		return m
	}
	state := caps[1]
	name := cleanName(caps[2])
	return state + `('` + name + `')`
}

func cleanName(raw string) string {
	if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
		return strings.Trim(u, `"'`)
	}
	name := strings.ReplaceAll(raw, `\"`, `"`)
	return strings.Trim(name, `"'`)
}
