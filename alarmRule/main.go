package alarmcleaner

import (
	"regexp"
	"strconv"
	"strings"
)

var statePattern = regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA|NOT)\s*\(\s*([^)]+?)\s*\)`)
var boolPattern = regexp.MustCompile(`(?i)\b(TRUE|FALSE)\b`)
var notTrueFalse = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`)

func Format(rule string) string {
	s := rule

	// 1. Najpierw zamień NOT(TRUE) / NOT(FALSE) na !true / !false
	s = notTrueFalse.ReplaceAllStringFunc(s, func(m string) string {
		caps := notTrueFalse.FindStringSubmatch(m)
		val := strings.ToLower(caps[1])
		return "!" + val
	})

	// 2. Zamień wszystkie TRUE/FALSE (same) na małe litery
	s = boolPattern.ReplaceAllStringFunc(s, func(m string) string {
		return strings.ToLower(m)
	})

	// 3. Teraz obsługa ALARM('name'), OK('name'), itd. – jak wcześniej
	s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := statePattern.FindStringSubmatch(m)
		if len(caps) != 3 {
			return m
		}

		state := caps[1]
		raw := caps[2]

		var name string
		if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
			name = strings.Trim(u, `"'`)
		} else {
			name = strings.ReplaceAll(raw, `\"`, `"`)
			name = strings.Trim(name, `"'`)
		}
		name = strings.TrimSpace(name)

		return state + `('` + name + `')`
	})

	// 4. Czyść białe znaki i formatuj pięknie
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR|NOT)\s+`).ReplaceAllString(s, " $1 ")
	s = strings.TrimSpace(s)

	return s
}

// for strings.Contains(s, "! ( !") || strings.Contains(s, "NOT ( NOT") {
//     s = regexp.MustCompile(`!\s*\(\s*!\s*([^)]+)\s*\)`).ReplaceAllString(s, "$1")
//     s = regexp.MustCompile(`NOT\s*\(\s*NOT\s*\(\s*([^)]+)\s*\)\s*\)`).ReplaceAllString(s, "$1")
// }

// func Format(rule string) string {
//     s := rule

//     // 1. NOT(TRUE) → !true, NOT(FALSE) → !false
//     s = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`).ReplaceAllStringFunc(s, func(m string) string {
//         return "!" + strings.ToLower(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m))
//     })

//     // 2. Uprość ! ( !x ) → x
//     for strings.Contains(s, "! ( !") {
//         s = regexp.MustCompile(`!\s*\(\s*!\s*([^)]+)\)`).ReplaceAllString(s, "$1")
//     }

//     // 3. Zamień luźne TRUE/FALSE na true/false
//     s = regexp.MustCompile(`(?i)\bTRUE\b`).ReplaceAllString(s, "true")
//     s = regexp.MustCompile(`(?i)\bFALSE\b`).ReplaceAllString(s, "false")

//     // 4. Formatuj stany: ALARM('name') itd.
//     s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
//         caps := statePattern.FindStringSubmatch(m)
//         state := caps[1]
//         raw := strings.TrimSpace(caps[2])

//         var name string
//         if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
//             name = strings.Trim(u, `"'`)
//         } else {
//             name = strings.Trim(strings.ReplaceAll(raw, `\"`, `"`), `"'`)
//         }

//         return state + `('` + name + `')`
//     })

//     // 5. Czyść format
//     s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
//     s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
//     return strings.TrimSpace(s)
// }
