package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create childStates structure
func makeChildStates(states map[string]AlarmStateName) map[AlarmStateName]map[string]struct{} {
	result := make(map[AlarmStateName]map[string]struct{})

	for alarmName, state := range states {
		if result[state] == nil {
			result[state] = make(map[string]struct{})
		}
		result[state][alarmName] = struct{}{}
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
		childStates    map[string]AlarmStateName
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive single alarm in ALARM state with OR logic should trigger",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m3": AlarmStateNameALARM,
				"m1": AlarmStateNameALARM,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3"},
		},
		{
			name:           "positive both alarms in ALARM state with AND logic should both trigger",
			awsRule:        "ALARM(cpu) AND ALARM(mem)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu": AlarmStateNameALARM,
				"mem": AlarmStateNameALARM,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu", "mem"},
		},
		{
			name:           "positive one alarm in ALARM state with complex OR expression should trigger",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a":     AlarmStateNameALARM,
				"b":     AlarmStateNameOK,
				"maint": AlarmStateNameOK,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
		},
		{
			name:           "negative both alarms in ALARM state with OR logic should not trigger individually",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a":     AlarmStateNameALARM,
				"b":     AlarmStateNameALARM,
				"maint": AlarmStateNameOK,
			},
			changedAlarm:   "b",
			wantTriggering: []string{},
		},
		{
			name:           "positive all three alarms in ALARM state with AND logic should all trigger",
			awsRule:        "ALARM(a) AND ALARM(b) AND ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameALARM,
				"b": AlarmStateNameALARM,
				"c": AlarmStateNameALARM,
			},
			changedAlarm:   "c",
			wantTriggering: []string{"a", "b", "c"},
		},
		{
			name:           "positive single alarm in ALARM state with three-way OR logic should trigger",
			awsRule:        "ALARM(a) OR ALARM(b) OR ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameALARM,
				"c": AlarmStateNameOK,
			},
			changedAlarm:   "b",
			wantTriggering: []string{"b"},
		},
		{
			name:           "positive one alarm in ALARM state and other in INSUFFICIENT_DATA should trigger only ALARM",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m2"},
		},
		{
			name:           "positive single alarm in ALARM state should trigger",
			awsRule:        "ALARM(cpu)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu": AlarmStateNameALARM,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
		},
		{
			name:           "positive alarm in ALARM state with NOT condition on maintenance should trigger",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu":         AlarmStateNameALARM,
				"maintenance": AlarmStateNameOK,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
		},
		// ════════════════════════════════════════════════════════════
		// Inverted Logic (OK triggers ALARM)
		// ════════════════════════════════════════════════════════════
		{
			name:           "positive both alarms in OK state with AND logic (inverted) should trigger",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive single alarm in OK state with OR logic (inverted) should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive single alarm in OK state (inverted) should trigger",
			awsRule:        "OK(health)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"health": AlarmStateNameOK,
			},
			changedAlarm:   "health",
			wantTriggering: []string{"health"},
		},
		{
			name:           "positive all three alarms in OK state with AND logic (inverted) should trigger",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameOK,
				"c": AlarmStateNameOK,
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

			assert.ElementsMatch(t, tt.wantTriggering, result)
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
		childStates    map[string]AlarmStateName
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive both alarms in OK state with OR logic should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive single alarm in OK state with OR logic should trigger",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms in OK state with AND logic should trigger",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive all alarms satisfying OK conditions should trigger",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m3": AlarmStateNameOK,
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3", "m1", "m2"},
		},
		{
			name:           "positive alarm in NOT ALARM state should trigger",
			awsRule:        "NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"maintenance": AlarmStateNameOK,
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{"maintenance"},
		},
		{
			name:           "negative suppressor alarm should not trigger",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"cpu":         AlarmStateNameALARM,
				"maintenance": AlarmStateNameALARM,
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{},
		},
		{
			name:           "positive all three alarms in OK state with AND logic should trigger",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameOK,
				"c": AlarmStateNameOK,
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
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive both alarms in OK state (NOT ALARM) with OR logic should trigger",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
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

			assert.ElementsMatch(t, tt.wantTriggering, result)
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
		childStates    map[string]AlarmStateName
		changedAlarm   string
		wantTriggering []string
	}{
		{
			name:           "positive single alarm with insufficient data should trigger",
			awsRule:        "ALARM(m1)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms with insufficient data should trigger",
			awsRule:        "ALARM(a) OR ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameINSUFFICIENT_DATA,
				"b": AlarmStateNameINSUFFICIENT_DATA,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b"},
		},
		{
			name:           "positive single alarm with insufficient data (AND logic) should trigger",
			awsRule:        "ALARM(a) AND ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameINSUFFICIENT_DATA,
				"b": AlarmStateNameOK,
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

			assert.ElementsMatch(t, tt.wantTriggering, result)
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
		childStates    map[string]AlarmStateName
		wantTriggering []string
	}{
		{
			name:           "positive single alarm with insufficient data in OR logic should trigger",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameOK,
			},
			wantTriggering: []string{"m1"},
		},
		{
			name:           "positive both alarms with insufficient data in OR logic should trigger",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameINSUFFICIENT_DATA,
			},
			wantTriggering: []string{"m1", "m2"},
		},
		{
			name:           "positive alarm not having insufficient data with NOT should trigger",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"metric": AlarmStateNameOK,
			},
			wantTriggering: []string{"metric"},
		},
		{
			name:           "negative alarm having insufficient data with NOT and ALARM state should not trigger",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"metric": AlarmStateNameINSUFFICIENT_DATA,
			},
			wantTriggering: []string{},
		},
		{
			name:           "positive multiple alarms in expected states should trigger",
			awsRule:        "OK(health) AND NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"health": AlarmStateNameOK,
				"metric": AlarmStateNameOK,
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

			assert.ElementsMatch(t, tt.wantTriggering, result)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// Edge Cases
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_EdgeCases(t *testing.T) {
	t.Run("should return error when composite state is invalid", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]AlarmStateName{"m1": AlarmStateNameALARM})

		_, err := GetAllTriggeringAlarms(rule, "INVALID_STATE", states, "m1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown composite state")
	})

	t.Run("should return empty when alarm is missing in states", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]AlarmStateName{}) // m1 missing

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "m1")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("should return [m1] when changed alarm is empty", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]AlarmStateName{"m1": AlarmStateNameALARM})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1"}, result)
	})

	t.Run("should return error when rule is empty", func(t *testing.T) {
		states := makeChildStates(map[string]AlarmStateName{"m1": AlarmStateNameALARM})

		_, err := GetAllTriggeringAlarms("", "ALARM", states, "")
		assert.Error(t, err)
	})

	t.Run("should return empty when alarm in OK state can't trigger ALARM", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := makeChildStates(map[string]AlarmStateName{"m1": AlarmStateNameOK})

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
		states := makeChildStates(map[string]AlarmStateName{
			"primary": AlarmStateNameALARM,
			"backup":  AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "backup")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"primary", "backup"}, result)
	})

	t.Run("negative multiple regions down with OR logic should not trigger individually", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(us-east) OR ALARM(eu-west) OR ALARM(ap-south)")
		states := makeChildStates(map[string]AlarmStateName{
			"us-east":  AlarmStateNameALARM,
			"eu-west":  AlarmStateNameOK,
			"ap-south": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "ap-south")
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("positive single healthy backend should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(backend-1) OR OK(backend-2) OR OK(backend-3)")
		states := makeChildStates(map[string]AlarmStateName{
			"backend-1": AlarmStateNameOK,
			"backend-2": AlarmStateNameALARM,
			"backend-3": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "backend-1")
		require.NoError(t, err)

		assert.Equal(t, []string{"backend-1"}, result)
	})

	t.Run("positive all components healthy should all trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(database) AND OK(api) AND OK(frontend)")
		states := makeChildStates(map[string]AlarmStateName{
			"database": AlarmStateNameOK,
			"api":      AlarmStateNameOK,
			"frontend": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "api", "frontend"}, result)
	})

	t.Run("positive all systems healthy and with data should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(api) AND OK(database) AND NOT INSUFFICIENT_DATA(metrics)")
		states := makeChildStates(map[string]AlarmStateName{
			"api":      AlarmStateNameOK,
			"database": AlarmStateNameOK,
			"metrics":  AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"api", "database", "metrics"}, result)
	})

	t.Run("positive cascading failure should return triggering alarms", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(database) AND (ALARM(app-1) OR ALARM(app-2))")
		states := makeChildStates(map[string]AlarmStateName{
			"database": AlarmStateNameALARM,
			"app-1":    AlarmStateNameALARM,
			"app-2":    AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "app-1"}, result)
	})

	t.Run("positive both healthy apps with inverted logic should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(app-1) AND OK(app-2)")
		states := makeChildStates(map[string]AlarmStateName{
			"app-1": AlarmStateNameOK,
			"app-2": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "app-1")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"app-1", "app-2"}, result)
	})

	t.Run("positive healthy monitor with inverted logic should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(monitor)")
		states := makeChildStates(map[string]AlarmStateName{
			"monitor": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "monitor")
		require.NoError(t, err)

		assert.Equal(t, []string{"monitor"}, result)
	})

	t.Run("negative all services healthy with OR logic (inverted) should not trigger individually", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]AlarmStateName{
			"service-a": AlarmStateNameOK,
			"service-b": AlarmStateNameOK,
			"service-c": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("positive single healthy service with OR logic (inverted) should trigger", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]AlarmStateName{
			"service-a": AlarmStateNameOK,
			"service-b": AlarmStateNameALARM,
			"service-c": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.Equal(t, []string{"service-a"}, result)
	})
}
