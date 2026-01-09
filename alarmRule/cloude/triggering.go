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

// AlarmName represents the name of an alarm
type AlarmName string

// AlarmState represents a single alarm with its name and current state
type AlarmState struct {
	Name  AlarmName
	State AlarmStateName
}

// GetAllTriggeringAlarms returns all alarms that are currently causing the composite alarm state.
//
// This implementation uses smart pre-analysis to skip alarms that cannot be triggering,
// then falls back to testing remaining alarms.
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

// getAllAlarmsCausingAlarmState finds all alarms causing composite to be ALARM.
func getAllAlarmsCausingAlarmState(transformedRule string, childStates []AlarmState) ([]AlarmName, error) {
	// CRITICAL: Handle empty rule
	if transformedRule == "" {
		return nil, fmt.Errorf("empty rule provided")
	}

	// Convert to map for faster lookup (but preserve order for output)
	// alarmToState := makeAlarmStateMap(childStates)

	// Extract alarm conditions from rule
	alarmConditions := extractAlarmConditions(transformedRule)

	// Filter alarms - but be more permissive
	candidateAlarms := []AlarmState{} // Preserve order

	for _, alarmState := range childStates {
		alarmName := string(alarmState.Name)
		actualState := alarmState.State

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
			candidateAlarms = append(candidateAlarms, alarmState)
		}
	}

	// Test only candidate alarms (preserve order)
	triggeringAlarms := []AlarmName{}

	for _, candidate := range candidateAlarms {
		alarmName := string(candidate.Name)
		originalState := candidate.State

		var simulatedState AlarmStateName

		switch originalState {
		case AlarmStateNameOK:
			simulatedState = AlarmStateNameALARM
		case AlarmStateNameALARM:
			simulatedState = AlarmStateNameOK
		case AlarmStateNameINSUFFICIENT_DATA:
			simulatedState = AlarmStateNameOK
		}

		modifiedStates := make([]AlarmState, len(childStates))
		copy(modifiedStates, childStates)

		// Update the specific alarm's state
		for i := range modifiedStates {
			if modifiedStates[i].Name == candidate.Name {
				modifiedStates[i].State = simulatedState
				break
			}
		}

		modifiedResult, err := evaluateRule(transformedRule, modifiedStates)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate with modified states for alarm %s: %w", alarmName, err)
		}

		if !modifiedResult {
			// For OnlyNegative alarms, check if they're truly triggering or just suppressors
			if cond, ok := alarmConditions[alarmName]; ok && cond.OnlyNegative {
				// Check if this is truly a trigger (problematic state) or a suppressor (safe state)
				isSuppressor := false

				if cond.HasNegativeOK && originalState == AlarmStateNameOK {
					isSuppressor = true
				} else if cond.HasNegativeALARM && originalState == AlarmStateNameOK {
					isSuppressor = true
				} else if cond.HasNegativeINSUFFICIENT && originalState == AlarmStateNameOK {
					isSuppressor = true
				}

				if isSuppressor {
					continue
				}
			}

			triggeringAlarms = append(triggeringAlarms, candidate.Name)
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
func evaluateRule(transformedRule string, childStates []AlarmState) (bool, error) {
	alarmToState := makeAlarmStateMap(childStates)

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

// makeAlarmStateMap converts slice to map for faster lookup
func makeAlarmStateMap(childStates []AlarmState) map[string]AlarmStateName {
	result := make(map[string]AlarmStateName)
	for _, alarmState := range childStates {
		result[string(alarmState.Name)] = alarmState.State
	}
	return result
}
