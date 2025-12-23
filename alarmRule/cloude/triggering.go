package alarmrule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
)

// AlarmStateName represents the state of an alarm
type AlarmStateName string

const (
	AlarmStateNameOK                AlarmStateName = "OK"
	AlarmStateNameALARM             AlarmStateName = "ALARM"
	AlarmStateNameINSUFFICIENT_DATA AlarmStateName = "INSUFFICIENT_DATA"
)

// GetAllTriggeringAlarms returns all alarms that are currently causing the composite alarm state.
//
// This implementation uses smart pre-analysis to skip alarms that cannot be triggering,
// then falls back to testing remaining alarms.
//
// Parameters:
//   - transformedRule: The alarm rule in govaluate format (from TransformAlarmRule)
//   - compositeState: Current state of composite alarm ("ALARM", "OK", or "INSUFFICIENT_DATA")
//   - childStates: Map of alarm states where key is state name and value is set of alarm names
//   - changedAlarm: The alarm that changed (optional, used for logging/context)
//
// Returns:
//   - List of alarm names that are triggering the current composite state
//   - Error if evaluation fails
func GetAllTriggeringAlarms(
	transformedRule string,
	compositeState string,
	childStates map[AlarmStateName]map[string]struct{},
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
func getAllAlarmsCausingAlarmState(transformedRule string, childStates map[AlarmStateName]map[string]struct{}) ([]string, error) {
	// CRITICAL: Handle empty rule
	if transformedRule == "" {
		return nil, fmt.Errorf("empty rule provided")
	}

	triggeringAlarms := []string{}
	allAlarms := flattenChildStates(childStates)

	// Extract alarm conditions from rule
	alarmConditions := extractAlarmConditions(transformedRule)

	// Filter alarms - but be more permissive
	candidateAlarms := make(map[string]AlarmStateName)

	for alarmName, actualState := range allAlarms {
		condition, found := alarmConditions[alarmName]
		if !found {
			continue // Alarm not in rule
		}

		// CRITICAL FIX: Include alarms based on their actual state and condition
		shouldInclude := false

		// Case 1: Alarms with positive conditions (must be problematic to trigger ALARM state)
		if condition.HasPositiveALARM || condition.HasPositiveOK || condition.HasPositiveINSUFFICIENT {
			// Include if in problematic state
			if actualState == AlarmStateNameALARM || actualState == AlarmStateNameINSUFFICIENT_DATA {
				shouldInclude = true
			}
			// For OK state: only include if has positive OK (like OK(...) for alarm state trigger)
			if actualState == AlarmStateNameOK && condition.HasPositiveOK {
				shouldInclude = true
			}
		}

		// Case 2: Alarms ONLY in negative conditions (!OK(x), !ALARM(x), etc)
		// These trigger ALARM state when the negated condition becomes FALSE
		if condition.OnlyNegative {
			// !OK(x) triggers ALARM when x is NOT in OK state (i.e., ALARM or INSUFFICIENT_DATA)
			if condition.HasNegativeOK && (actualState == AlarmStateNameALARM || actualState == AlarmStateNameINSUFFICIENT_DATA) {
				shouldInclude = true
			}
			// !ALARM(x) triggers ALARM when x is NOT in ALARM state (i.e., OK or INSUFFICIENT_DATA)
			if condition.HasNegativeALARM && (actualState == AlarmStateNameOK || actualState == AlarmStateNameINSUFFICIENT_DATA) {
				shouldInclude = true
			}
			// !INSUFFICIENT_DATA(x) triggers ALARM when x is NOT in INSUFFICIENT_DATA state
			if condition.HasNegativeINSUFFICIENT && (actualState == AlarmStateNameOK || actualState == AlarmStateNameALARM) {
				shouldInclude = true
			}
		}

		if shouldInclude {
			candidateAlarms[alarmName] = actualState
		}
	}

	// Test only candidate alarms
	for alarmName, originalState := range candidateAlarms {
		var simulatedState AlarmStateName

		switch originalState {
		case AlarmStateNameOK:
			simulatedState = AlarmStateNameALARM
		case AlarmStateNameALARM:
			simulatedState = AlarmStateNameOK
		case AlarmStateNameINSUFFICIENT_DATA:
			simulatedState = AlarmStateNameOK
		}

		modifiedStates := copyChildStates(childStates)

		if modifiedStates[originalState] != nil {
			delete(modifiedStates[originalState], alarmName)
		}

		if modifiedStates[simulatedState] == nil {
			modifiedStates[simulatedState] = make(map[string]struct{})
		}
		modifiedStates[simulatedState][alarmName] = struct{}{}

		modifiedResult, err := evaluateRule(transformedRule, modifiedStates)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate with modified states for alarm %s: %w", alarmName, err)
		}

		if !modifiedResult {
			// For OnlyNegative alarms, check if they're truly triggering or just suppressors
			// A suppressor is an OnlyNegative alarm in a "safe" state (not triggering the negation)
			if cond, ok := alarmConditions[alarmName]; ok && cond.OnlyNegative {
				// Check if this is truly a trigger (problematic state) or a suppressor (safe state)
				isSuppressor := false

				if cond.HasNegativeOK && originalState == AlarmStateNameOK {
					// !OK(x) where x=OK is a suppressor (false negation)
					isSuppressor = true
				} else if cond.HasNegativeALARM && originalState == AlarmStateNameOK {
					// !ALARM(x) where x=OK is a suppressor (false negation)
					isSuppressor = true
				} else if cond.HasNegativeINSUFFICIENT && originalState == AlarmStateNameOK {
					// !INSUFFICIENT_DATA(x) where x=OK is a suppressor (false negation)
					isSuppressor = true
				}

				if isSuppressor {
					// Skip suppressors from triggering list
					continue
				}
			}

			triggeringAlarms = append(triggeringAlarms, alarmName)
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
	// Pattern: optional ! followed by function name and quoted argument
	pattern := regexp.MustCompile(`(!?)(ALARM|OK|INSUFFICIENT_DATA)\('([^']*(?:\\.[^']*)*)'\)`)
	matches := pattern.FindAllStringSubmatchIndex(transformedRule, -1)

	for _, matchIdx := range matches {
		// matchIdx[0:2] = full match
		// matchIdx[2:4] = group 1 (!)
		// matchIdx[4:6] = group 2 (function name)
		// matchIdx[6:8] = group 3 (alarm name inside quotes)

		isNegated := matchIdx[2] < matchIdx[3] // Check if ! group has content
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
func getAllAlarmsCausingOKState(transformedRule string, childStates map[AlarmStateName]map[string]struct{}) ([]string, error) {
	expectedStates := parseExpectedStates(transformedRule)
	triggeringAlarms := []string{}
	alarmToState := flattenChildStates(childStates)

	for alarmName, actualState := range alarmToState {
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
			triggeringAlarms = append(triggeringAlarms, alarmName)
		}
	}

	return triggeringAlarms, nil
}

// getAllAlarmsInInsufficientDataState returns all alarms in INSUFFICIENT_DATA state.
func getAllAlarmsInInsufficientDataState(childStates map[AlarmStateName]map[string]struct{}) ([]string, error) {
	insufficientDataAlarms := []string{}

	if alarms, ok := childStates[AlarmStateNameINSUFFICIENT_DATA]; ok {
		for alarmName := range alarms {
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

	reOK := regexp.MustCompile(`OK\('([^']+)'\)`)
	matchesOK := reOK.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesOK {
		if len(match) > 1 {
			alarmName := match[1]
			// Unescape name
			alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
			alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)
			expectedStates[alarmName] = ExpectedOK
		}
	}

	reAlarm := regexp.MustCompile(`!?ALARM\('([^']+)'\)`)
	matchesAlarm := reAlarm.FindAllStringSubmatch(transformedRule, -1)
	for _, match := range matchesAlarm {
		if len(match) > 1 {
			alarmName := match[1]
			alarmName = strings.ReplaceAll(alarmName, `\'`, `'`)
			alarmName = strings.ReplaceAll(alarmName, `\"`, `"`)
			fullMatch := match[0]
			if strings.HasPrefix(fullMatch, "!") {
				expectedStates[alarmName] = ExpectedNotAlarm
			} else {
				expectedStates[alarmName] = ExpectedNotAlarm
			}
		}
	}

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

// evaluateRule evaluates the transformed alarm rule
func evaluateRule(transformedRule string, childStates map[AlarmStateName]map[string]struct{}) (bool, error) {
	alarmToState := flattenChildStates(childStates)

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmToState[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == AlarmStateNameALARM, nil
		},
		"OK": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmToState[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == AlarmStateNameOK, nil
		},
		"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
			alarmName := args[0].(string)
			state, ok := alarmToState[alarmName]
			if !ok {
				return false, fmt.Errorf("alarm %s not found in states", alarmName)
			}
			return state == AlarmStateNameINSUFFICIENT_DATA, nil
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

// Helper functions

func flattenChildStates(childStates map[AlarmStateName]map[string]struct{}) map[string]AlarmStateName {
	result := make(map[string]AlarmStateName)

	for state, alarms := range childStates {
		for alarmName := range alarms {
			result[alarmName] = state
		}
	}

	return result
}

func copyChildStates(childStates map[AlarmStateName]map[string]struct{}) map[AlarmStateName]map[string]struct{} {
	result := make(map[AlarmStateName]map[string]struct{})

	for state, alarms := range childStates {
		result[state] = make(map[string]struct{})
		for alarmName := range alarms {
			result[state][alarmName] = struct{}{}
		}
	}

	return result
}
