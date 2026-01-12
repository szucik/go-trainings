package alarmrule

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create childStates structure as ordered slice
// Sorts alarm names alphabetically for deterministic order
func makeChildStates(states map[string]string) []AlarmState {
	// Extract keys and sort them for deterministic order
	keys := make([]string, 0, len(states))
	for key := range states {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := []AlarmState{}
	for _, alarmName := range keys {
		result = append(result, AlarmState{
			Name:  AlarmName(alarmName),
			State: AlarmStateName(states[alarmName]),
		})
	}

	return result
}

// Helper function to convert []AlarmName to []string for test assertions
func alarmNamesToStrings(alarms []AlarmName) []string {
	result := make([]string, len(alarms))
	for i, alarm := range alarms {
		result[i] = string(alarm)
	}
	return result
}

// ═══════════════════════════════════════════════════════════
// Tests for Composite State = ALARM
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_CompositeALARM(t *testing.T) {
	tests := []struct {
		name           string
		awsRule        string
		compositeState string
		childStates    map[string]string
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive single alarm in ALARM state with OR logic should trigger",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m3": "ALARM",
				"m1": "ALARM",
				"m2": "ALARM",
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3"},
		},
		{
			name:           "positive both alarms in ALARM state with AND logic should both trigger",
			awsRule:        "ALARM(cpu) AND ALARM(mem)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu": "ALARM",
				"mem": "ALARM",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu", "mem"},
		},
		{
			name:           "positive one alarm in ALARM state with complex OR expression should trigger with negative",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a":     "ALARM",
				"b":     "OK",
				"maint": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "maint"}, // ✅ ZMIENIONE: maint też jest liczony
		},
		{
			name:           "positive both alarms in ALARM state with OR logic should both trigger with negative",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a":     "ALARM",
				"b":     "ALARM",
				"maint": "OK",
			},
			changedAlarm:   "b",
			wantTriggering: []string{"a", "b", "maint"}, // ✅ ZMIENIONE: wszystkie spełniające
		},
		{
			name:           "positive all three alarms in ALARM state with AND logic should all trigger",
			awsRule:        "ALARM(a) AND ALARM(b) AND ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "ALARM",
				"b": "ALARM",
				"c": "ALARM",
			},
			changedAlarm:   "c",
			wantTriggering: []string{"a", "b", "c"},
		},
		{
			name:           "positive single alarm in ALARM state with three-way OR logic should trigger",
			awsRule:        "ALARM(a) OR ALARM(b) OR ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "OK",
				"b": "ALARM",
				"c": "OK",
			},
			changedAlarm:   "b",
			wantTriggering: []string{"b"},
		},
		{
			name:           "positive one alarm in ALARM state and other in INSUFFICIENT_DATA should trigger only ALARM",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "ALARM",
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m2"},
		},
		{
			name:           "positive single alarm in ALARM state should trigger",
			awsRule:        "ALARM(cpu)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu": "ALARM",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
		},
		{
			name:           "positive alarm in ALARM state with NOT condition on maintenance should trigger both",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu":         "ALARM",
				"maintenance": "OK",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu", "maintenance"}, // ✅ ZMIENIONE: maintenance też liczony
		},
		// ════════════════════════════════════════════════════════════
		// Inverted Logic (OK triggers ALARM)
		// ════════════════════════════════════════════════════════════
		{
			name:           "positive both alarms in OK state with AND logic (inverted) should trigger",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive single alarm in OK state with OR logic (inverted) should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "ALARM",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive single alarm in OK state (inverted) should trigger",
			awsRule:        "OK(health)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"health": "OK",
			},
			changedAlarm:   "health",
			wantTriggering: []string{"health"},
		},
		{
			name:           "positive all three alarms in OK state with AND logic (inverted) should trigger",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "OK",
				"b": "OK",
				"c": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Transformed: %s", transformed)
			t.Logf("Changed Alarm: %s", tt.changedAlarm)
			t.Logf("Child States: %v", tt.childStates)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				states,
				tt.changedAlarm,
			)
			require.NoError(t, err)

			t.Logf("Triggering Alarms: %v", result)

			assert.ElementsMatch(t, tt.wantTriggering, alarmNamesToStrings(result))
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Tests for Composite State = OK
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_CompositeOK(t *testing.T) {
	tests := []struct {
		name           string
		awsRule        string
		compositeState string
		childStates    map[string]string
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive both alarms in OK state with OR logic should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive single alarm in OK state with OR logic should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "ALARM",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms in OK state with AND logic should trigger",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive all alarms satisfying OK conditions should trigger",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m3": "OK",
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3", "m1", "m2"},
		},
		{
			name:           "positive alarm in NOT ALARM state should trigger",
			awsRule:        "NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]string{
				"maintenance": "OK",
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{"maintenance"},
		},
		{
			name:           "negative suppressor alarm should not trigger",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]string{
				"cpu":         "ALARM",
				"maintenance": "ALARM",
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{},
		},
		{
			name:           "positive all three alarms in OK state with AND logic should trigger",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "OK",
			childStates: map[string]string{
				"a": "OK",
				"b": "OK",
				"c": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b", "c"},
		},
		// ════════════════════════════════════════════════════════════
		// ALARM triggers OK (inverted)
		// ════════════════════════════════════════════════════════════
		{
			name:           "positive both alarms in OK state (NOT ALARM) with AND logic should trigger",
			awsRule:        "ALARM(m1) AND ALARM(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive both alarms in OK state (NOT ALARM) with OR logic should trigger",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Changed Alarm: %s", tt.changedAlarm)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				states,
				tt.changedAlarm,
			)
			require.NoError(t, err)

			t.Logf("Triggering Alarms: %v", result)

			assert.ElementsMatch(t, tt.wantTriggering, alarmNamesToStrings(result))
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Tests for Composite State = INSUFFICIENT_DATA
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_CompositeINSUFFICIENT_DATA(t *testing.T) {
	tests := []struct {
		name           string
		awsRule        string
		compositeState string
		childStates    map[string]string
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive single alarm with insufficient data should trigger",
			awsRule:        "ALARM(m1)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms with insufficient data should trigger",
			awsRule:        "ALARM(a) OR ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"a": "INSUFFICIENT_DATA",
				"b": "INSUFFICIENT_DATA",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b"},
		},
		{
			name:           "positive single alarm with insufficient data (AND logic) should trigger",
			awsRule:        "ALARM(a) AND ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"a": "INSUFFICIENT_DATA",
				"b": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				states,
				tt.changedAlarm,
			)
			require.NoError(t, err)

			t.Logf("Triggering Alarms: %v", result)

			assert.ElementsMatch(t, tt.wantTriggering, alarmNamesToStrings(result))
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Tests for INSUFFICIENT_DATA in AlarmRule
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_INSUFFICIENT_DATA_InRule(t *testing.T) {
	tests := []struct {
		name           string
		awsRule        string
		compositeState string
		childStates    map[string]string
		wantTriggering []string
	}{
		{
			name:           "positive single alarm with insufficient data in OR logic should trigger",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "OK",
			},
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms with insufficient data in OR logic should trigger",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive alarm not having insufficient data with NOT should trigger",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]string{
				"metric": "OK",
			},
			wantTriggering: []string{"metric"},
		},
		{
			name:           "negative alarm having insufficient data with NOT and ALARM state should not trigger",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"metric": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{},
		},
		{
			name:           "positive multiple alarms in expected states should trigger",
			awsRule:        "OK(health) AND NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]string{
				"health": "OK",
				"metric": "OK",
			},
			wantTriggering: []string{"health", "metric"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Transformed: %s", transformed)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				states,
				"",
			)
			require.NoError(t, err)

			t.Logf("Triggering Alarms: %v", result)

			assert.ElementsMatch(t, tt.wantTriggering, alarmNamesToStrings(result))
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Edge Cases
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_EdgeCases(t *testing.T) {
	t.Run("should return error when composite state is invalid", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]string{"m1": "ALARM"})

		_, err := GetAllTriggeringAlarms(rule, "INVALID_STATE", states, "m1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown composite state")
	})

	t.Run("should return empty when alarm is missing in states", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]string{}) // m1 missing

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "m1")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("should return [m1] when changed alarm is empty", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]string{"m1": "ALARM"})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1"}, alarmNamesToStrings(result))
	})

	t.Run("should return error when rule is empty", func(t *testing.T) {
		states := makeChildStates(map[string]string{"m1": "ALARM"})

		_, err := GetAllTriggeringAlarms("", "ALARM", states, "")
		assert.Error(t, err)
	})

	t.Run("should return empty when alarm in OK state can't trigger ALARM", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]string{"m1": "OK"})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "m1")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// ═══════════════════════════════════════════════════════════
// Real-World Scenarios
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_RealWorldScenarios(t *testing.T) {
	t.Run("positive both servers down should return all triggering alarms", func(t *testing.T) {
		rule := TransformAlarmRule("NOT OK(primary) AND NOT OK(backup)")
		states := makeChildStates(map[string]string{
			"primary": "ALARM",
			"backup":  "ALARM",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "backup")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"primary", "backup"}, alarmNamesToStrings(result))
	})

	t.Run("positive multiple regions down with OR logic should show all satisfying", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(us-east) OR ALARM(eu-west) OR ALARM(ap-south)")
		states := makeChildStates(map[string]string{
			"us-east":  "ALARM",
			"eu-west":  "OK",
			"ap-south": "ALARM",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "ap-south")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"us-east", "ap-south"}, alarmNamesToStrings(result)) // ✅ ZMIENIONE
	})

	t.Run("positive single healthy backend should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(backend-1) OR OK(backend-2) OR OK(backend-3)")
		states := makeChildStates(map[string]string{
			"backend-1": "OK",
			"backend-2": "ALARM",
			"backend-3": "ALARM",
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "backend-1")
		require.NoError(t, err)

		assert.Equal(t, []string{"backend-1"}, alarmNamesToStrings(result))
	})

	t.Run("positive all components healthy should all trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(database) AND OK(api) AND OK(frontend)")
		states := makeChildStates(map[string]string{
			"database": "OK",
			"api":      "OK",
			"frontend": "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "api", "frontend"}, alarmNamesToStrings(result))
	})

	t.Run("positive all systems healthy and with data should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(api) AND OK(database) AND NOT INSUFFICIENT_DATA(metrics)")
		states := makeChildStates(map[string]string{
			"api":      "OK",
			"database": "OK",
			"metrics":  "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"api", "database", "metrics"}, alarmNamesToStrings(result))
	})

	t.Run("positive cascading failure should return triggering alarms", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(database) AND (ALARM(app-1) OR ALARM(app-2))")
		states := makeChildStates(map[string]string{
			"database": "ALARM",
			"app-1":    "ALARM",
			"app-2":    "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "app-1"}, alarmNamesToStrings(result))
	})

	t.Run("positive both healthy apps with inverted logic should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(app-1) AND OK(app-2)")
		states := makeChildStates(map[string]string{
			"app-1": "OK",
			"app-2": "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "app-1")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"app-1", "app-2"}, alarmNamesToStrings(result))
	})

	t.Run("positive healthy monitor with inverted logic should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(monitor)")
		states := makeChildStates(map[string]string{
			"monitor": "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "monitor")
		require.NoError(t, err)

		assert.Equal(t, []string{"monitor"}, alarmNamesToStrings(result))
	})

	t.Run("positive all services healthy with OR logic (inverted) should show all", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]string{
			"service-a": "OK",
			"service-b": "OK",
			"service-c": "OK",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"service-a", "service-b", "service-c"}, alarmNamesToStrings(result)) // ✅ ZMIENIONE
	})

	t.Run("positive single healthy service with OR logic (inverted) should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]string{
			"service-a": "OK",
			"service-b": "ALARM",
			"service-c": "ALARM",
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.Equal(t, []string{"service-a"}, alarmNamesToStrings(result))
	})
}
