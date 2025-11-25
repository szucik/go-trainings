func FormatCloudWatchAlarmRule(rule string) string {
	re := regexp.MustCompile(`(?s)(ALARM|OK|INSUFFICIENT_DATA|NOT)\s*\(\s*([^)]+?)\s*\)`)

	s := re.ReplaceAllStringFunc(rule, func(m string) string {
		caps := re.FindStringSubmatch(m)
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

	// Piękny format
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\s+(AND|OR|NOT)\s+`).ReplaceAllString(s, " $1 ")
	return strings.TrimSpace(s)
}