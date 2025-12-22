package alarmrule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

// GetAllTriggeringAlarms returns all alarms that are currently causing the composite alarm state.
func GetAllTriggeringAlarms(
	transformedRule string,
	compositeState string,
	childStates map[string]string,
	changedAlarm string,
) ([]string, error) {

	switch compositeState {
	case "ALARM":
		return getAllAlarmsCausingAlarmState(transformedRule, childStates)

	case "OK":
		return getAllAlarmsCausingOKState(transformedRule, childStates)

	case "INSUFFICIENT_DATA":
		return getAllAlarmsInInsufficientDataState(childStates)

	default:
		return nil, fmt.Errorf("unknown composite state: %s", compositeState)
	}
}

// getAllAlarmsCausingAlarmState finds all alarms causing composite to be ALARM
func getAllAlarmsCausingAlarmState(transformedRule string, childStates map[string]string) ([]string, error) {
	compositeIsAlarm, err := evaluateRule(transformedRule, childStates)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rule: %w", err)
	}

	if !compositeIsAlarm {
		return []string{}, nil
	}

	triggeringAlarms := []string{}

	for alarmName, actualState := range childStates {
		if actualState != "ALARM" && actualState != "INSUFFICIENT_DATA" {
			continue
		}

		modifiedStates := make(map[string]string)
		for k, v := range childStates {
			modifiedStates[k] = v
		}
		modifiedStates[alarmName] = "OK"

		modifiedResult, err := evaluateRule(transformedRule, modifiedStates)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate with modified states: %w", err)
		}

		if !modifiedResult {
			triggeringAlarms = append(triggeringAlarms, alarmName)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsCausingOKState finds all alarms causing composite to be OK
func getAllAlarmsCausingOKState(transformedRule string, childStates map[string]string) ([]string, error) {
	compositeIsAlarm, err := evaluateRule(transformedRule, childStates)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate rule: %w", err)
	}

	if compositeIsAlarm {
		return []string{}, nil
	}

	// Parse expected states from the rule
	expectedStates := parseExpectedStates(transformedRule)

	triggeringAlarms := []string{}

	// Check which alarms are in their expected state
	for alarmName, actualState := range childStates {
		expected, found := expectedStates[alarmName]
		if !found {
			continue
		}

		// Check if alarm is in expected state
		isInExpectedState := false

		switch expected {
		case "OK":
			isInExpectedState = (actualState == "OK")

		case "NOT_ALARM":
			isInExpectedState = (actualState != "ALARM")

		case "NOT_INSUFFICIENT_DATA":
			isInExpectedState = (actualState != "INSUFFICIENT_DATA")

		case "INSUFFICIENT_DATA":
			isInExpectedState = (actualState == "INSUFFICIENT_DATA")

		case "ALARM":
			// For OK composite state, we don't expect ALARM
			// But if rule has ALARM(x) and we're OK, it means alarm is NOT in ALARM
			isInExpectedState = (actualState != "ALARM")
		}

		if isInExpectedState {
			triggeringAlarms = append(triggeringAlarms, alarmName)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsInInsufficientDataState returns all alarms in INSUFFICIENT_DATA state
func getAllAlarmsInInsufficientDataState(childStates map[string]string) ([]string, error) {
	insufficientDataAlarms := []string{}

	for alarmName, state := range childStates {
		if state == "INSUFFICIENT_DATA" {
			insufficientDataAlarms = append(insufficientDataAlarms, alarmName)
		}
	}

	return insufficientDataAlarms, nil
}

// ExpectedState represents what state an alarm should be in
type ExpectedState string

const (
	ExpectedOK                  ExpectedState = "OK"
	ExpectedAlarm               ExpectedState = "ALARM"
	ExpectedInsufficientData    ExpectedState = "INSUFFICIENT_DATA"
	ExpectedNotAlarm            ExpectedState = "NOT_ALARM"
	ExpectedNotInsufficientData ExpectedState = "NOT_INSUFFICIENT_DATA"
)

// parseExpectedStates extracts expected states from the rule
func parseExpectedStates(transformedRule string) map[string]ExpectedState {
	expectedStates := make(map[string]ExpectedState)

	// Pattern: OK('alarm_name')
	reOK := regexp.MustCompile(`OK\('([^']+)'\)`)
	matchesOK := reOK.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesOK {
		if len(match) > 1 {
			alarmName := match[1]
			expectedStates[alarmName] = ExpectedOK
		}
	}

	// Pattern: ALARM('alarm_name')
	reAlarm := regexp.MustCompile(`ALARM\('([^']+)'\)`)
	matchesAlarm := reAlarm.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesAlarm {
		if len(match) > 1 {
			alarmName := match[1]
			// Check if it's negated
			idx := strings.Index(transformedRule, match[0])
			if idx > 0 && transformedRule[idx-1] == '!' {
				expectedStates[alarmName] = ExpectedNotAlarm
			} else {
				// For OK composite, ALARM(x) means we expect x NOT to be ALARM
				expectedStates[alarmName] = ExpectedNotAlarm
			}
		}
	}

	// Pattern: INSUFFICIENT_DATA('alarm_name')
	reInsufficient := regexp.MustCompile(`INSUFFICIENT_DATA\('([^']+)'\)`)
	matchesInsufficient := reInsufficient.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesInsufficient {
		if len(match) > 1 {
			alarmName := match[1]
			// Check if it's negated
			idx := strings.Index(transformedRule, match[0])
			if idx > 0 && transformedRule[idx-1] == '!' {
				expectedStates[alarmName] = ExpectedNotInsufficientData
			} else {
				expectedStates[alarmName] = ExpectedInsufficientData
			}
		}
	}

	return expectedStates
}

// evaluateRule evaluates the transformed alarm rule
func evaluateRule(transformedRule string, alarmStates map[string]string) (bool, error) {
	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmStates[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == "ALARM", nil
		},

		"OK": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmStates[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == "OK", nil
		},

		"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmStates[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == "INSUFFICIENT_DATA", nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformedRule, functions)
	if err != nil {
		return false, fmt.Errorf("failed to parse expression: %w", err)
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result.(bool), nil
}
