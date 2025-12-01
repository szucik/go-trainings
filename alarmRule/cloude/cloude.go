package alarmrule

import (
	"fmt"
	"regexp"
	"strings"
)

type AlarmState string

const (
	StateOK               AlarmState = "OK"
	StateAlarm            AlarmState = "ALARM"
	StateInsufficientData AlarmState = "INSUFFICIENT_DATA"
)

// FixAlarmRule konwertuje AlarmRule użytkownika do formatu wymaganego przez govaluate
func FixAlarmRule(input string) (string, error) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Pattern: STATE(content)
	pattern := regexp.MustCompile(`^(OK|ALARM|INSUFFICIENT_DATA)\((.+?)\)$`)

	matches := pattern.FindStringSubmatch(input)
	if matches == nil {
		return "", fmt.Errorf("nieprawidłowy format AlarmRule: %s", input)
	}

	state := matches[1]
	content := matches[2]

	// Krok 1: Jeśli content zaczyna i kończy się apostrofem ' - usuń je
	if len(content) >= 2 && content[0] == '\'' && content[len(content)-1] == '\'' {
		content = content[1 : len(content)-1]
	}

	// Krok 2: Escapuj wszystkie " i '
	content = strings.ReplaceAll(content, `"`, `\"`)
	content = strings.ReplaceAll(content, `'`, `\'`)

	// Krok 3: Opakuj w apostrofy
	result := fmt.Sprintf(`%s('%s')`, state, content)

	return result, nil
}

// ValidateAlarmState sprawdza czy podany stan jest prawidłowy
func ValidateAlarmState(state string) bool {
	validStates := []AlarmState{StateOK, StateAlarm, StateInsufficientData}
	for _, valid := range validStates {
		if state == string(valid) {
			return true
		}
	}
	return false
}
