package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		description    string
	}{
		{
			name:           "should return [m3] when only m3 is in ALARM with OR logic",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m3": "ALARM",
				"m1": "ALARM",
				"m2": "ALARM",
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3"},
			description:    "Only m3 in ALARM causes composite ALARM (first condition satisfied)",
		},
		{
			name:           "should return [cpu, mem] when both alarms are ALARM with AND logic",
			awsRule:        "ALARM(cpu) AND ALARM(mem)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu": "ALARM",
				"mem": "ALARM",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu", "mem"},
			description:    "Both cpu and mem must be ALARM, so both are triggering",
		},
		{
			name:           "should return [a] when only a is ALARM in complex OR expression",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a":     "ALARM",
				"b":     "OK",
				"maint": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
			description:    "a is ALARM and not in maintenance, so only a is triggering",
		},
		{
			name:           "should return empty when both a and b are ALARM with OR logic",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a":     "ALARM",
				"b":     "ALARM",
				"maint": "OK",
			},
			changedAlarm:   "b",
			wantTriggering: []string{},
			description:    "Both a and b are ALARM; OR means either is sufficient, so neither individually required",
		},
		{
			name:           "should return [a, b, c] when all three alarms are ALARM with AND logic",
			awsRule:        "ALARM(a) AND ALARM(b) AND ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "ALARM",
				"b": "ALARM",
				"c": "ALARM",
			},
			changedAlarm:   "c",
			wantTriggering: []string{"a", "b", "c"},
			description:    "All three must be ALARM (AND), so all are triggering",
		},
		{
			name:           "should return [b] when only b is ALARM with three-way OR logic",
			awsRule:        "ALARM(a) OR ALARM(b) OR ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "OK",
				"b": "ALARM",
				"c": "OK",
			},
			changedAlarm:   "b",
			wantTriggering: []string{"b"},
			description:    "Only b is ALARM, so only b is triggering",
		},
		{
			name:           "should return [m2] when m2 is ALARM and m1 is INSUFFICIENT_DATA",
			awsRule:        "ALARM(m1) OR ALARM(m2)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "ALARM",
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m2"},
			description:    "m2 is ALARM (triggering), m1 is INSUFFICIENT_DATA but doesn't match ALARM()",
		},
		{
			name:           "should return [cpu] when single alarm is in ALARM state",
			awsRule:        "ALARM(cpu)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu": "ALARM",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
			description:    "Single alarm in ALARM state",
		},
		{
			name:           "should return [cpu] when cpu is ALARM and maintenance is NOT ALARM",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu":         "ALARM",
				"maintenance": "OK",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
			description:    "cpu is ALARM and maintenance is not ALARM, so cpu triggers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)
			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Transformed: %s", transformed)
			t.Logf("Changed Alarm: %s", tt.changedAlarm)
			t.Logf("Child States: %v", tt.childStates)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				tt.childStates,
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
		childStates    map[string]string
		changedAlarm   string
		wantTriggering []string
		description    string
	}{
		{
			name:           "should return [m1, m2] when both are OK with OR logic",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 and m2 are OK, both contribute to OK state",
		},
		{
			name:           "should return [m1] when only m1 is OK with OR logic",
			awsRule:        "OK(m1) OR OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "ALARM",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
			description:    "Only m1 is OK, it's keeping composite OK",
		},
		{
			name:           "should return [m1, m2] when both must be OK with AND logic",
			awsRule:        "OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m2",
			wantTriggering: []string{"m1", "m2"},
			description:    "Both m1 AND m2 must be OK, so both are triggering",
		},
		{
			name:           "should return [m3, m1, m2] when all satisfy OK conditions",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m3": "OK",
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m3", "m1", "m2"},
			description:    "m3 is not ALARM (expected for OK), m1 AND m2 are OK (both conditions satisfied)",
		},
		{
			name:           "should return [maintenance] when maintenance is NOT ALARM",
			awsRule:        "NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]string{
				"maintenance": "OK",
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{"maintenance"},
			description:    "maintenance is NOT ALARM (it's OK), satisfying the condition",
		},
		{
			name:           "should return empty when maintenance suppresses alarm",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]string{
				"cpu":         "ALARM",
				"maintenance": "ALARM",
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{},
			description:    "maintenance being ALARM prevents composite ALARM (composite is OK, no clear triggering)",
		},
		{
			name:           "should return [a, b, c] when all three are OK with AND logic",
			awsRule:        "OK(a) AND OK(b) AND OK(c)",
			compositeState: "OK",
			childStates: map[string]string{
				"a": "OK",
				"b": "OK",
				"c": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b", "c"},
			description:    "All three must be OK (AND), so all are triggering",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)
			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Changed Alarm: %s", tt.changedAlarm)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				tt.childStates,
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
		childStates    map[string]string
		changedAlarm   string
		wantTriggering []string
		description    string
	}{
		{
			name:           "should return [m1] when m1 has insufficient data",
			awsRule:        "ALARM(m1)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
			},
			changedAlarm:   "m1",
			wantTriggering: []string{"m1"},
			description:    "m1 is in INSUFFICIENT_DATA state",
		},
		{
			name:           "should return [a, b] when both alarms have insufficient data",
			awsRule:        "ALARM(a) OR ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"a": "INSUFFICIENT_DATA",
				"b": "INSUFFICIENT_DATA",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a", "b"},
			description:    "Both alarms are in INSUFFICIENT_DATA state",
		},
		{
			name:           "should return [a] when only a has insufficient data",
			awsRule:        "ALARM(a) AND ALARM(b)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"a": "INSUFFICIENT_DATA",
				"b": "OK",
			},
			changedAlarm:   "a",
			wantTriggering: []string{"a"},
			description:    "Only a is in INSUFFICIENT_DATA state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				tt.childStates,
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
		childStates    map[string]string
		wantTriggering []string
		description    string
	}{
		{
			name:           "should return [m1] when m1 has insufficient data with OR logic",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "OK",
			},
			wantTriggering: []string{"m1"},
			description:    "m1 is INSUFFICIENT_DATA (expected), m2 is not",
		},
		{
			name:           "should return [m1, m2] when both have insufficient data",
			awsRule:        "INSUFFICIENT_DATA(m1) OR INSUFFICIENT_DATA(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
				"m2": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{"m1", "m2"},
			description:    "Both have INSUFFICIENT_DATA (both expected)",
		},
		{
			name:           "should return [metric] when metric is NOT INSUFFICIENT_DATA",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]string{
				"metric": "OK",
			},
			wantTriggering: []string{"metric"},
			description:    "metric is NOT INSUFFICIENT_DATA (as expected)",
		},
		{
			name:           "should return empty when metric IS INSUFFICIENT_DATA with NOT rule",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"metric": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{},
			description:    "metric IS INSUFFICIENT_DATA, but with NOT rule and ALARM state, no alarm is triggering (edge case)",
		},
		{
			name:           "should return [health, metric] when both are in expected states",
			awsRule:        "OK(health) AND NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]string{
				"health": "OK",
				"metric": "OK",
			},
			wantTriggering: []string{"health", "metric"},
			description:    "Both in expected states (health=OK, metric not INSUFFICIENT_DATA)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)
			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Transformed: %s", transformed)

			result, err := GetAllTriggeringAlarms(
				transformed,
				tt.compositeState,
				tt.childStates,
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
		states := map[string]string{"m1": "ALARM"}

		_, err := GetAllTriggeringAlarms(rule, "INVALID_STATE", states, "m1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown composite state")
	})

	t.Run("should return empty when alarm is missing in states", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{} // m1 missing

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "m1")

		// For ALARM composite state with missing alarm, function returns empty list
		// because it skips alarms not in childStates during iteration
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("should return [m1] when changed alarm is empty", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{"m1": "ALARM"}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1"}, result)
	})

	t.Run("should return error when rule is empty", func(t *testing.T) {
		states := map[string]string{"m1": "ALARM"}

		_, err := GetAllTriggeringAlarms("", "ALARM", states, "")
		assert.Error(t, err)
	})

	t.Run("should return empty when alarm in OK state can't trigger ALARM", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{"m1": "OK"}

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
		childStates := map[string]string{
			"primary": "ALARM",
			"backup":  "ALARM",
		}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", childStates, "backup")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"primary", "backup"}, result)
		t.Logf("Both servers down - complete outage")
	})

	t.Run("should return empty when multiple regions are down with OR logic", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(us-east) OR ALARM(eu-west) OR ALARM(ap-south)")
		childStates := map[string]string{
			"us-east":  "ALARM",
			"eu-west":  "OK",
			"ap-south": "ALARM",
		}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", childStates, "ap-south")
		require.NoError(t, err)

		// With OR logic, if multiple alarms are in ALARM, removing one still leaves ALARM
		// So neither is individually required (both contribute but neither is critical)
		assert.Empty(t, result)
		t.Logf("Two regions down with OR - neither individually critical")
	})

	t.Run("should return [backend-1] when only one backend is healthy", func(t *testing.T) {
		rule := TransformAlarmRule("OK(backend-1) OR OK(backend-2) OR OK(backend-3)")
		childStates := map[string]string{
			"backend-1": "OK",
			"backend-2": "ALARM",
			"backend-3": "ALARM",
		}

		result, err := GetAllTriggeringAlarms(rule, "OK", childStates, "backend-1")
		require.NoError(t, err)

		assert.Equal(t, []string{"backend-1"}, result)
		t.Logf("Only backend-1 is healthy and keeping system operational")
	})

	t.Run("should return [database, api, frontend] when all components are healthy", func(t *testing.T) {
		rule := TransformAlarmRule("OK(database) AND OK(api) AND OK(frontend)")
		childStates := map[string]string{
			"database": "OK",
			"api":      "OK",
			"frontend": "OK",
		}

		result, err := GetAllTriggeringAlarms(rule, "OK", childStates, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "api", "frontend"}, result)
		t.Logf("All critical components healthy")
	})

	t.Run("should return [api, database, metrics] when all systems have data", func(t *testing.T) {
		rule := TransformAlarmRule("OK(api) AND OK(database) AND NOT INSUFFICIENT_DATA(metrics)")
		childStates := map[string]string{
			"api":      "OK",
			"database": "OK",
			"metrics":  "OK",
		}

		result, err := GetAllTriggeringAlarms(rule, "OK", childStates, "")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"api", "database", "metrics"}, result)
		t.Logf("All systems healthy and have sufficient data")
	})

	t.Run("should return [database, app-1] when cascading failure detected", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(database) AND (ALARM(app-1) OR ALARM(app-2))")
		childStates := map[string]string{
			"database": "ALARM",
			"app-1":    "ALARM",
			"app-2":    "OK",
		}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", childStates, "database")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"database", "app-1"}, result)
		t.Logf("Database and app-1 both down - cascading failure")
	})
}
