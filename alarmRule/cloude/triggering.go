package alarmrule

import (
	"fmt"
	"regexp"
	"strings"
)

// AlarmStateName represents the state of an alarm
type AlarmStateName string

const (
	AlarmStateNameOK                AlarmStateName = "OK"
	AlarmStateNameALARM             AlarmStateName = "ALARM"
	AlarmStateNameINSUFFICIENT_DATA AlarmStateName = "INSUFFICIENT_DATA"
)

// AlarmName represents the name of an alarm
type AlarmName string

// AlarmState represents a single alarm with its name and current state
type AlarmState struct {
	Name  AlarmName
	State AlarmStateName
}

// GetAllTriggeringAlarms returns all alarms that are in the state specified by the rule,
// causing composite to be in the given state.
//
// For composite state ALARM: returns ALL alarms satisfying their conditions (including negative-only)
// For composite state OK: returns alarms keeping the composite in OK state
// For composite state INSUFFICIENT_DATA: returns all alarms in INSUFFICIENT_DATA state
//
// Parameters:
//   - transformedRule: The alarm rule in govaluate format (from TransformAlarmRule)
//   - compositeState: Current state of composite alarm ("ALARM", "OK", or "INSUFFICIENT_DATA")
//   - childStates: Ordered list of alarm states
//   - changedAlarm: The alarm that changed (optional, used for logging/context)
//
// Returns:
//   - Ordered list of alarm names that are triggering the current composite state
//   - Error if evaluation fails
func GetAllTriggeringAlarms(
	transformedRule string,
	compositeState string,
	childStates []AlarmState,
	changedAlarm string,
) ([]AlarmName, error) {

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

// getAllAlarmsCausingAlarmState returns ALL alarms that satisfy their conditions in the rule,
// causing composite to be ALARM.
//
// For OK(x) → alarm must be in OK state
// For ALARM(x) → alarm must be in ALARM state
// For !OK(x) → alarm must NOT be in OK state
// For !ALARM(x) → alarm must NOT be in ALARM state
//
// This includes negative-only conditions (e.g., !ALARM(maintenance))
func getAllAlarmsCausingAlarmState(transformedRule string, childStates []AlarmState) ([]AlarmName, error) {
	if transformedRule == "" {
		return nil, fmt.Errorf("empty rule provided")
	}

	alarmConditions := extractAlarmConditions(transformedRule)
	triggeringAlarms := []AlarmName{}

	for _, alarmState := range childStates {
		alarmName := string(alarmState.Name)
		actualState := alarmState.State

		condition, found := alarmConditions[alarmName]
		if !found {
			continue // Alarm not in rule
		}

		satisfiesCondition := false

		// Positive conditions: alarm must BE in specific state
		if (condition.HasPositiveOK && actualState == AlarmStateNameOK) ||
			(condition.HasPositiveALARM && actualState == AlarmStateNameALARM) ||
			(condition.HasPositiveINSUFFICIENT && actualState == AlarmStateNameINSUFFICIENT_DATA) {
			satisfiesCondition = true
		}

		// Negative conditions: alarm must NOT be in specific state
		// This includes OnlyNegative alarms (e.g., !ALARM(maintenance))
		if condition.OnlyNegative {
			if (condition.HasNegativeOK && actualState != AlarmStateNameOK) ||
				(condition.HasNegativeALARM && actualState != AlarmStateNameALARM) ||
				(condition.HasNegativeINSUFFICIENT && actualState != AlarmStateNameINSUFFICIENT_DATA) {
				satisfiesCondition = true
			}
		}

		if satisfiesCondition {
			triggeringAlarms = append(triggeringAlarms, alarmState.Name)
		}
	}

	return triggeringAlarms, nil
}

// AlarmCondition describes how an alarm appears in the rule
type AlarmCondition struct {
	HasPositiveALARM        bool // ALARM(x) - alarm must be ALARM
	HasPositiveOK           bool // OK(x) - alarm must be OK
	HasPositiveINSUFFICIENT bool // INSUFFICIENT_DATA(x)
	HasNegativeALARM        bool // !ALARM(x) - alarm must NOT be ALARM
	HasNegativeOK           bool // !OK(x) - alarm must NOT be OK
	HasNegativeINSUFFICIENT bool // !INSUFFICIENT_DATA(x)
	OnlyNegative            bool // Only appears in negative conditions
}

// extractAlarmConditions analyzes the rule to determine how each alarm is used
func extractAlarmConditions(transformedRule string) map[string]*AlarmCondition {
	conditions := make(map[string]*AlarmCondition)

	// Use regex to find all function calls with their context
	pattern := regexp.MustCompile(`(!?)(ALARM|OK|INSUFFICIENT_DATA)\('([^']*(?:\\.[^']*)*)'\)`)
	matches := pattern.FindAllStringSubmatchIndex(transformedRule, -1)

	for _, matchIdx := range matches {
		isNegated := matchIdx[2] < matchIdx[3]
		funcName := transformedRule[matchIdx[4]:matchIdx[5]]
		alarmName := transformedRule[matchIdx[6]:matchIdx[7]]

		// Unescape alarm name
		alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
		alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)

		if alarmName == "" {
			continue
		}

		if conditions[alarmName] == nil {
			conditions[alarmName] = &AlarmCondition{}
		}

		// Record how this alarm is used
		switch funcName {
		case "ALARM":
			if isNegated {
				conditions[alarmName].HasNegativeALARM = true
			} else {
				conditions[alarmName].HasPositiveALARM = true
			}
		case "OK":
			if isNegated {
				conditions[alarmName].HasNegativeOK = true
			} else {
				conditions[alarmName].HasPositiveOK = true
			}
		case "INSUFFICIENT_DATA":
			if isNegated {
				conditions[alarmName].HasNegativeINSUFFICIENT = true
			} else {
				conditions[alarmName].HasPositiveINSUFFICIENT = true
			}
		}
	}

	// Determine if alarms are only in negative conditions
	for _, condition := range conditions {
		hasPositive := condition.HasPositiveALARM || condition.HasPositiveOK || condition.HasPositiveINSUFFICIENT
		hasNegative := condition.HasNegativeALARM || condition.HasNegativeOK || condition.HasNegativeINSUFFICIENT
		condition.OnlyNegative = hasNegative && !hasPositive
	}

	return conditions
}

// getAllAlarmsCausingOKState finds all alarms causing composite to be OK.
func getAllAlarmsCausingOKState(transformedRule string, childStates []AlarmState) ([]AlarmName, error) {
	expectedStates := parseExpectedStates(transformedRule)
	triggeringAlarms := []AlarmName{}

	for _, alarmState := range childStates {
		alarmName := string(alarmState.Name)
		actualState := alarmState.State

		expected, found := expectedStates[alarmName]
		if !found {
			continue
		}

		isInExpectedState := false

		switch expected {
		case ExpectedOK:
			isInExpectedState = (actualState == AlarmStateNameOK)
		case ExpectedNotAlarm:
			isInExpectedState = (actualState != AlarmStateNameALARM)
		case ExpectedNotInsufficientData:
			isInExpectedState = (actualState != AlarmStateNameINSUFFICIENT_DATA)
		case ExpectedInsufficientData:
			isInExpectedState = (actualState == AlarmStateNameINSUFFICIENT_DATA)
		case ExpectedAlarm:
			isInExpectedState = (actualState != AlarmStateNameALARM)
		}

		if isInExpectedState {
			triggeringAlarms = append(triggeringAlarms, alarmState.Name)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsInInsufficientDataState returns all alarms in INSUFFICIENT_DATA state.
func getAllAlarmsInInsufficientDataState(childStates []AlarmState) ([]AlarmName, error) {
	insufficientDataAlarms := []AlarmName{}

	for _, alarmState := range childStates {
		if alarmState.State == AlarmStateNameINSUFFICIENT_DATA {
			insufficientDataAlarms = append(insufficientDataAlarms, alarmState.Name)
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

	// OK() calls
	reOK := regexp.MustCompile(`OK\('([^']+)'\)`)
	matchesOK := reOK.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesOK {
		if len(match) > 1 {
			alarmName := match[1]
			alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
			alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)
			expectedStates[alarmName] = ExpectedOK
		}
	}

	// ALARM() and !ALARM() calls
	// For composite=OK state, both ALARM(x) and !ALARM(x) mean the alarm should NOT be in ALARM state
	reAlarm := regexp.MustCompile(`!?ALARM\('([^']+)'\)`)
	matchesAlarm := reAlarm.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesAlarm {
		if len(match) > 1 {
			alarmName := match[1]
			alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
			alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)
			expectedStates[alarmName] = ExpectedNotAlarm
		}
	}

	// INSUFFICIENT_DATA() and !INSUFFICIENT_DATA() calls
	reInsufficient := regexp.MustCompile(`!?INSUFFICIENT_DATA\('([^']+)'\)`)
	matchesInsufficient := reInsufficient.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesInsufficient {
		if len(match) > 1 {
			alarmName := match[1]
			alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
			alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)
			fullMatch := match[0]
			if strings.HasPrefix(fullMatch, "!") {
				expectedStates[alarmName] = ExpectedNotInsufficientData
			} else {
				expectedStates[alarmName] = ExpectedInsufficientData
			}
		}
	}

	return expectedStates
}
