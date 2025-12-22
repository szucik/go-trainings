package alarmrule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

// GetAllTriggeringAlarms returns all alarms that are currently causing the composite alarm state.
//
// IMPORTANT: This function TRUSTS the provided compositeState. It does NOT verify that
// the composite state matches the evaluation of transformedRule with childStates.
// This is by design for performance - you should already know the correct composite state.
//
// Parameters:
//   - transformedRule: The alarm rule in govaluate format (from TransformAlarmRule)
//   - compositeState: Current state of composite alarm ("ALARM", "OK", or "INSUFFICIENT_DATA")
//   - childStates: Current states of all child alarms
//   - changedAlarm: The alarm that changed (optional, currently unused but kept for context)
//
// Returns:
//   - List of alarm names that are triggering the current composite state
//   - Error if evaluation fails
//
// Example:
//
//	rule := TransformAlarmRule("ALARM(cpu) AND ALARM(mem)")
//	states := map[string]string{"cpu": "ALARM", "mem": "ALARM"}
//	triggering, err := GetAllTriggeringAlarms(rule, "ALARM", states, "cpu")
//	// Returns: ["cpu", "mem"]
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

// getAllAlarmsCausingAlarmState finds all alarms causing composite to be ALARM.
// ASSUMES composite is already in ALARM state (no verification).
//
// Logic:
//   - For each alarm in ALARM or INSUFFICIENT_DATA state
//   - Simulate changing it to OK
//   - If composite would become OK, this alarm is triggering
func getAllAlarmsCausingAlarmState(transformedRule string, childStates map[string]string) ([]string, error) {
	triggeringAlarms := []string{}

	// For each alarm in problematic state (ALARM or INSUFFICIENT_DATA)
	for alarmName, actualState := range childStates {
		// Skip alarms that are OK (they can't cause ALARM)
		if actualState != "ALARM" && actualState != "INSUFFICIENT_DATA" {
			continue
		}

		// Simulate: change this alarm to OK
		modifiedStates := make(map[string]string)
		for k, v := range childStates {
			modifiedStates[k] = v
		}
		modifiedStates[alarmName] = "OK"

		// Re-evaluate with this alarm as OK
		modifiedResult, err := evaluateRule(transformedRule, modifiedStates)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate with modified states: %w", err)
		}

		// If composite would become OK (false), this alarm IS causing ALARM
		if !modifiedResult {
			triggeringAlarms = append(triggeringAlarms, alarmName)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsCausingOKState finds all alarms causing composite to be OK.
// ASSUMES composite is already in OK state (no verification).
//
// Logic:
//   - Parse expected states from the alarm rule
//   - For each alarm, check if it's in the expected state
//   - If yes, it's triggering the OK state
func getAllAlarmsCausingOKState(transformedRule string, childStates map[string]string) ([]string, error) {
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
		case ExpectedOK:
			isInExpectedState = (actualState == "OK")

		case ExpectedNotAlarm:
			isInExpectedState = (actualState != "ALARM")

		case ExpectedNotInsufficientData:
			isInExpectedState = (actualState != "INSUFFICIENT_DATA")

		case ExpectedInsufficientData:
			isInExpectedState = (actualState == "INSUFFICIENT_DATA")

		case ExpectedAlarm:
			isInExpectedState = (actualState != "ALARM")
		}

		if isInExpectedState {
			triggeringAlarms = append(triggeringAlarms, alarmName)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsInInsufficientDataState returns all alarms in INSUFFICIENT_DATA state.
// ASSUMES composite is in INSUFFICIENT_DATA state (no verification).
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

// parseExpectedStates extracts expected states from the rule.
// It parses the transformed rule to determine what state each alarm should be in
// for the composite to be OK.
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
				// For OK composite state, ALARM(x) in rule means x should NOT be ALARM
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

// evaluateRule evaluates the transformed alarm rule and returns true if composite should be ALARM.
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
