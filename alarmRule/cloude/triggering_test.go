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
		description    string
	}{
		{
			name:           "should return [m3] when only m3 is in ALARM with OR logic",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m3": AlarmStateNameALARM,
				"m1": AlarmStateNameALARM,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3"},
			description:    "Only m3 in ALARM causes composite ALARM (first condition satisfied)",
		},
		{
			name:           "should return [cpu, mem] when both alarms are ALARM with AND logic",
			awsRule:        "ALARM(cpu) AND ALARM(mem)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu": AlarmStateNameALARM,
				"mem": AlarmStateNameALARM,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu", "mem"},
			description:    "Both cpu and mem must be ALARM, so both are triggering",
		},
		{
			name:           "should return [a] when only a is ALARM in complex OR expression",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a":     AlarmStateNameALARM,
				"b":     AlarmStateNameOK,
				"maint": AlarmStateNameOK,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
			description:    "a is ALARM and not in maintenance, so only a is triggering",
		},
		{
			name:           "should return empty when both a and b are ALARM with OR logic",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a":     AlarmStateNameALARM,
				"b":     AlarmStateNameALARM,
				"maint": AlarmStateNameOK,
			},
			changedAlarm:   "b",
			wantTriggering: []string{},
			description:    "Both a and b are ALARM; OR means either is sufficient, so neither individually required",
		},
		{
			name:           "should return [a, b, c] when all three alarms are ALARM with AND logic",
			awsRule:        "ALARM(a) AND ALARM(b) AND ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameALARM,
				"b": AlarmStateNameALARM,
				"c": AlarmStateNameALARM,
			},
			changedAlarm:   "c",
			wantTriggering: []string{"a", "b", "c"},
			description:    "All three must be ALARM (AND), so all are triggering",
		},
		{
			name:           "should return [b] when only b is ALARM with three-way OR logic",
			awsRule:        "ALARM(a) OR ALARM(b) OR ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameALARM,
				"c": AlarmStateNameOK,
			},
			changedAlarm:   "b",
			wantTriggering: []string{"b"},
			description:    "Only b is ALARM, so only b is triggering",
		},
		{
			name:           "should return [m2] when m2 is ALARM and m1 is INSUFFICIENT_DATA",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m2"},
			description:    "m2 is ALARM (triggering), m1 is INSUFFICIENT_DATA but doesn't match ALARM()",
		},
		{
			name:           "should return [cpu] when single alarm is in ALARM state",
			awsRule:        "ALARM(cpu)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu": AlarmStateNameALARM,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
			description:    "Single alarm in ALARM state",
		},
		{
			name:           "should return [cpu] when cpu is ALARM and maintenance is NOT ALARM",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"cpu":         AlarmStateNameALARM,
				"maintenance": AlarmStateNameOK,
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
			description:    "cpu is ALARM and maintenance is not ALARM, so cpu triggers (maintenance is negative-only)",
		},
		// ════════════════════════════════════════════════════════════
		// NEW TESTS - Inverted Logic (OK triggers ALARM)
		// ════════════════════════════════════════════════════════════
		{
			name:           "should return [m1, m2] when both are OK with AND logic (inverted)",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
			description:    "Inverted logic: both m1 AND m2 being OK causes composite ALARM",
		},
		{
			name:           "should return [m1] when only m1 is OK with OR logic (inverted)",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
			description:    "Inverted logic: m1 being OK causes composite ALARM (OR satisfied)",
		},
		{
			name:           "should return [health] when health is OK (inverted single alarm)",
			awsRule:        "OK(health)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"health": AlarmStateNameOK,
			},
			changedAlarm:   "health",
			wantTriggering: []string{"health"},
			description:    "Inverted logic: single alarm being OK causes composite ALARM",
		},
		{
			name:           "should return [a, b, c] when all three are OK (inverted)",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameOK,
				"c": AlarmStateNameOK,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b", "c"},
			description:    "Inverted logic: all three being OK causes composite ALARM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("Description: %s", tt.description)
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
		description    string
	}{
		{
			name:           "should return [m1, m2] when both are OK with OR logic",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 and m2 are OK, both contribute to OK state",
		},
		{
			name:           "should return [m1] when only m1 is OK with OR logic",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameALARM,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
			description:    "Only m1 is OK, it's keeping composite OK",
		},
		{
			name:           "should return [m1, m2] when both must be OK with AND logic",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 AND m2 must be OK, so both are triggering",
		},
		{
			name:           "should return [m3, m1, m2] when all satisfy OK conditions",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m3": AlarmStateNameOK,
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3", "m1", "m2"},
			description:    "m3 is not ALARM (expected for OK), m1 AND m2 are OK (both conditions satisfied)",
		},
		{
			name:           "should return [maintenance] when maintenance is NOT ALARM",
			awsRule:        "NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"maintenance": AlarmStateNameOK,
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{"maintenance"},
			description:    "maintenance is NOT ALARM (it's OK), satisfying the condition",
		},
		{
			name:           "should return empty when maintenance suppresses alarm",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"cpu":         AlarmStateNameALARM,
				"maintenance": AlarmStateNameALARM,
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{},
			description:    "maintenance being ALARM prevents composite ALARM (composite is OK, no clear triggering)",
		},
		{
			name:           "should return [a, b, c] when all three are OK with AND logic",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameOK,
				"b": AlarmStateNameOK,
				"c": AlarmStateNameOK,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b", "c"},
			description:    "All three must be OK (AND), so all are triggering",
		},
		// ════════════════════════════════════════════════════════════
		// NEW TESTS - ALARM triggers OK (inverted)
		// ════════════════════════════════════════════════════════════
		{
			name:           "should return [m1, m2] when both are NOT ALARM with AND logic",
			awsRule:        "ALARM(m1) AND ALARM(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 and m2 being NOT ALARM (OK) keeps composite OK",
		},
		{
			name:           "should return [m1, m2] when both are NOT ALARM with OR logic",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameOK,
				"m2": AlarmStateNameOK,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 and m2 being NOT ALARM keeps composite OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("Description: %s", tt.description)
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
		description    string
	}{
		{
			name:           "should return [m1] when m1 has insufficient data",
			awsRule:        "ALARM(m1)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
			description:    "m1 is in INSUFFICIENT_DATA state",
		},
		{
			name:           "should return [a, b] when both alarms have insufficient data",
			awsRule:        "ALARM(a) OR ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameINSUFFICIENT_DATA,
				"b": AlarmStateNameINSUFFICIENT_DATA,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b"},
			description:    "Both alarms are in INSUFFICIENT_DATA state",
		},
		{
			name:           "should return [a] when only a has insufficient data",
			awsRule:        "ALARM(a) AND ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]AlarmStateName{
				"a": AlarmStateNameINSUFFICIENT_DATA,
				"b": AlarmStateNameOK,
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
			description:    "Only a is in INSUFFICIENT_DATA state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("Description: %s", tt.description)

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
		description    string
	}{
		{
			name:           "should return [m1] when m1 has insufficient data with OR logic",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameOK,
			},
			wantTriggering: []string{"m1"},
			description:    "m1 is INSUFFICIENT_DATA (expected), m2 is not",
		},
		{
			name:           "should return [m1, m2] when both have insufficient data",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"m1": AlarmStateNameINSUFFICIENT_DATA,
				"m2": AlarmStateNameINSUFFICIENT_DATA,
			},
			wantTriggering: []string{"m1", "m2"},
			description:    "Both have INSUFFICIENT_DATA (both expected)",
		},
		{
			name:           "should return [metric] when metric is NOT INSUFFICIENT_DATA",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"metric": AlarmStateNameOK,
			},
			wantTriggering: []string{"metric"},
			description:    "metric is NOT INSUFFICIENT_DATA (as expected)",
		},
		{
			name:           "should return empty when metric IS INSUFFICIENT_DATA with NOT rule",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "ALARM",
			childStates: map[string]AlarmStateName{
				"metric": AlarmStateNameINSUFFICIENT_DATA,
			},
			wantTriggering: []string{},
			description:    "metric IS INSUFFICIENT_DATA, but with NOT rule and ALARM state, no alarm is triggering (edge case)",
		},
		{
			name:           "should return [health, metric] when both are in expected states",
			awsRule:        "OK(health) AND NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]AlarmStateName{
				"health": AlarmStateNameOK,
				"metric": AlarmStateNameOK,
			},
			wantTriggering: []string{"health", "metric"},
			description:    "Both in expected states (health=OK, metric not INSUFFICIENT_DATA)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)
			states := makeChildStates(tt.childStates)

			t.Logf("Description: %s", tt.description)
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
	t.Run("should return [primary, backup] when both servers are down", func(t *testing.T) {
		rule := TransformAlarmRule("NOT OK(primary) AND NOT OK(backup)")
		states := makeChildStates(map[string]AlarmStateName{
			"primary": AlarmStateNameALARM,
			"backup":  AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "backup")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"primary", "backup"}, result)
		t.Logf("Both servers down - complete outage")
	})

	t.Run("should return empty when multiple regions are down with OR logic", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(us-east) OR ALARM(eu-west) OR ALARM(ap-south)")
		states := makeChildStates(map[string]AlarmStateName{
			"us-east":  AlarmStateNameALARM,
			"eu-west":  AlarmStateNameOK,
			"ap-south": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "ap-south")
		require.NoError(t, err)

		assert.Empty(t, result)
		t.Logf("Two regions down with OR - neither individually critical")
	})

	t.Run("should return [backend-1] when only one backend is healthy", func(t *testing.T) {
		rule := TransformAlarmRule("OK(backend-1) OR OK(backend-2) OR OK(backend-3)")
		states := makeChildStates(map[string]AlarmStateName{
			"backend-1": AlarmStateNameOK,
			"backend-2": AlarmStateNameALARM,
			"backend-3": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "backend-1")
		require.NoError(t, err)

		assert.Equal(t, []string{"backend-1"}, result)
		t.Logf("Only backend-1 is healthy and keeping system operational")
	})

	t.Run("should return [database, api, frontend] when all components are healthy", func(t *testing.T) {
		rule := TransformAlarmRule("OK(database) AND OK(api) AND OK(frontend)")
		states := makeChildStates(map[string]AlarmStateName{
			"database": AlarmStateNameOK,
			"api":      AlarmStateNameOK,
			"frontend": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "api", "frontend"}, result)
		t.Logf("All critical components healthy")
	})

	t.Run("should return [api, database, metrics] when all systems have data", func(t *testing.T) {
		rule := TransformAlarmRule("OK(api) AND OK(database) AND NOT INSUFFICIENT_DATA(metrics)")
		states := makeChildStates(map[string]AlarmStateName{
			"api":      AlarmStateNameOK,
			"database": AlarmStateNameOK,
			"metrics":  AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"api", "database", "metrics"}, result)
		t.Logf("All systems healthy and have sufficient data")
	})

	t.Run("should return [database, app-1] when cascading failure detected", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(database) AND (ALARM(app-1) OR ALARM(app-2))")
		states := makeChildStates(map[string]AlarmStateName{
			"database": AlarmStateNameALARM,
			"app-1":    AlarmStateNameALARM,
			"app-2":    AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "app-1"}, result)
		t.Logf("Database and app-1 both down - cascading failure")
	})
	// ════════════════════════════════════════════════════════════
	// NEW REAL-WORLD TESTS - Inverted Logic
	// ════════════════════════════════════════════════════════════

	t.Run("should return [app-1, app-2] when both apps are healthy but composite alarms (inverted)", func(t *testing.T) {
		rule := TransformAlarmRule("OK(app-1) AND OK(app-2)")
		states := makeChildStates(map[string]AlarmStateName{
			"app-1": AlarmStateNameOK,
			"app-2": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "app-1")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"app-1", "app-2"}, result)
		t.Logf("Inverted logic: Both apps healthy triggers composite ALARM")
	})

	t.Run("should return [monitor] when monitor is healthy but should alarm (inverted)", func(t *testing.T) {
		rule := TransformAlarmRule("OK(monitor)")
		states := makeChildStates(map[string]AlarmStateName{
			"monitor": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "monitor")
		require.NoError(t, err)

		assert.Equal(t, []string{"monitor"}, result)
		t.Logf("Inverted logic: Monitor healthy triggers composite ALARM")
	})

	t.Run("should return empty when all services are up with OR logic (inverted)", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]AlarmStateName{
			"service-a": AlarmStateNameOK,
			"service-b": AlarmStateNameOK,
			"service-c": AlarmStateNameOK,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.Empty(t, result)
		t.Logf("Inverted logic with OR: All services healthy but none individually required")
	})

	t.Run("should return [service-a] when only one service is up with OR logic (inverted)", func(t *testing.T) {
		rule := TransformAlarmRule("OK(service-a) OR OK(service-b) OR OK(service-c)")
		states := makeChildStates(map[string]AlarmStateName{
			"service-a": AlarmStateNameOK,
			"service-b": AlarmStateNameALARM,
			"service-c": AlarmStateNameALARM,
		})

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "service-a")
		require.NoError(t, err)

		assert.Equal(t, []string{"service-a"}, result)
		t.Logf("Inverted logic: Only service-a healthy and sufficient for composite ALARM")
	})
}
