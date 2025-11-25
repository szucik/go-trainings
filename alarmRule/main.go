func Format(rule string) string {
	s := rule

	// 1. NOT(TRUE) → !true, NOT(FALSE) → !false
	s = regexp.MustCompile(`(?i)NOT\s*\(\s*(TRUE|FALSE)\s*\)`).ReplaceAllStringFunc(s, func(m string) string {
		return "!" + strings.ToLower(regexp.MustCompile(`(?i)TRUE|FALSE`).FindString(m))
	})

	// 2. Uprość ! ( !x ) → x
	for strings.Contains(s, "! ( !") {
		s = regexp.MustCompile(`!\s*\(\s*!\s*([^)]+)\)`).ReplaceAllString(s, "$1")
	}

	// 3. Zamień luźne TRUE/FALSE na true/false
	s = regexp.MustCompile(`(?i)\bTRUE\b`).ReplaceAllString(s, "true")
	s = regexp.MustCompile(`(?i)\bFALSE\b`).ReplaceAllString(s, "false")

	// 4. Formatuj stany: ALARM('name') itd.
	s = statePattern.ReplaceAllStringFunc(s, func(m string) string {
		caps := statePattern.FindStringSubmatch(m)
		state := caps[1]
		raw := strings.TrimSpace(caps[2])

		var name string
		if u, err := strconv.Unquote(`"` + raw + `"`); err == nil {
			name = strings.Trim(u, `"'`)
		} else {
			name = strings.Trim(strings.ReplaceAll(raw, `\"`, `"`), `"'`)
		}

		return state + `('` + name + `')`
	})

	// 5. Czyść format
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR)\s+`).ReplaceAllString(s, " $1 ")
	return strings.TrimSpace(s)
}