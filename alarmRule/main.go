// Najlepsza funkcja na świecie do CloudWatch Composite Alarms
func FormatAlarmRule(rule string) string {
	// Krok 1: normalizuj każdy ALARM(...)
	normalized := regexp.MustCompile(`ALARM\s*\([^)]*\)`).ReplaceAllStringFunc(rule, func(m string) string {
		inner := strings.Trim(m, "ALARM()")
		inner = strings.TrimSpace(inner)

		if s, err := strconv.Unquote(`"` + inner + `"`); err == nil {
			name := strings.TrimSpace(strings.Trim(s, `"`))
			return `ALARM('` + name + `')`
		}
		name := strings.ReplaceAll(inner, `\"`, `"`)
		name = strings.Trim(name, `"`)
		return `ALARM('` + strings.TrimSpace(name) + `')`
	})

	// Krok 2: czyść białe znaki
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = regexp.MustCompile(`\s+(AND|OR|NOT)\s+`).ReplaceAllString(normalized, " $1 ")
	return strings.TrimSpace(normalized)
}