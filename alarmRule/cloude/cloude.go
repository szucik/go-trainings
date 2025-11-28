package alarmfmt

import (
	"fmt"
	"regexp"
)

func FormatAlarmRule(input string) (string, error) {
	re := regexp.MustCompile(`([A-Z_]+)\((.*)\)`)

	result := re.ReplaceAllStringFunc(input, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if len(matches) != 3 {
			return s
		}
		funcName := matches[1]
		inner := matches[2]

		escaped := ""
		for _, r := range inner {
			if r == '"' {
				escaped += `\"`
			} else {
				escaped += string(r)
			}
		}

		return fmt.Sprintf("%s('%s')", funcName, escaped)
	})

	return result, nil
}
