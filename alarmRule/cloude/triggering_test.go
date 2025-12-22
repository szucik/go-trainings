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
			name:           "OR - m3 causes ALARM",
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
			name:           "AND - both alarms cause ALARM",
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
			name:           "Complex - only a causes ALARM",
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
			name:           "Complex - both a and b cause ALARM",
			awsRule:        "(ALARM(a) OR ALARM(b)) AND NOT ALARM(maint)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a":     "ALARM",
				"b":     "ALARM",
				"maint": "OK",
			},
			changedAlarm:   "b",
			wantTriggering: []string{"a", "b"},
			description:    "Both a and b are ALARM, both independently cause ALARM",
		},
		{
			name:           "Three-way OR - all ALARM",
			awsRule:        "ALARM(a) OR ALARM(b) OR ALARM(c)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"a": "ALARM",
				"b": "ALARM",
				"c": "ALARM",
			},
			changedAlarm:   "c",
			wantTriggering: []string{"a", "b", "c"},
			description:    "All three are ALARM and each independently causes ALARM (OR logic)",
		},
		{
			name:           "Three-way OR - only one ALARM",
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
			name:           "INSUFFICIENT_DATA as alarm state",
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
			name:           "Simple single alarm",
			awsRule:        "ALARM(cpu)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"cpu": "ALARM",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{"cpu"},
			description:    "Single alarm in ALARM state",
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
			name:           "OR - both m1 and m2 are OK",
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
			name:           "OR - only m1 is OK",
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
			name:           "AND - both must be OK",
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
			name:           "Complex - m1 and m2 keep OK",
			awsRule:        "ALARM(m3) OR OK(m1) AND OK(m2)",
			compositeState: "OK",
			childStates: map[string]string{
				"m3": "OK",
				"m1": "OK",
				"m2": "OK",
			},
			changedAlarm:   "m3",
			wantTriggering: []string{"m1", "m2"},
			description:    "m3 is not ALARM (OK), m1 AND m2 are OK (second condition satisfied)",
		},
		{
			name:           "Simple - all OK",
			awsRule:        "ALARM(cpu)",
			compositeState: "OK",
			childStates: map[string]string{
				"cpu": "OK",
			},
			changedAlarm:   "cpu",
			wantTriggering: []string{},
			description:    "cpu is not ALARM, so composite is OK (no specific triggering alarms for negation)",
		},
		{
			name:           "NOT ALARM - alarm is OK",
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
			name:           "Complex with NOT",
			awsRule:        "ALARM(cpu) AND NOT ALARM(maintenance)",
			compositeState: "OK",
			childStates: map[string]string{
				"cpu":         "ALARM",
				"maintenance": "ALARM",
			},
			changedAlarm:   "maintenance",
			wantTriggering: []string{},
			description:    "maintenance being ALARM prevents composite ALARM (but composite is OK for other reasons)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)
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
			name:           "One alarm has insufficient data",
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
			name:           "Multiple alarms have insufficient data",
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
			name:           "Mixed states with insufficient data",
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
			name:           "INSUFFICIENT_DATA OR - one has insufficient data",
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
			name:           "INSUFFICIENT_DATA OR - both have insufficient data",
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
			name:           "NOT INSUFFICIENT_DATA - alarm has data",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "OK",
			childStates: map[string]string{
				"metric": "OK",
			},
			wantTriggering: []string{"metric"},
			description:    "metric is NOT INSUFFICIENT_DATA (as expected)",
		},
		{
			name:           "NOT INSUFFICIENT_DATA - alarm lacks data (causes ALARM)",
			awsRule:        "NOT INSUFFICIENT_DATA(metric)",
			compositeState: "ALARM",
			childStates: map[string]string{
				"metric": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{"metric"},
			description:    "metric IS INSUFFICIENT_DATA, causing composite ALARM",
		},
		{
			name:           "Mixed - OK and INSUFFICIENT_DATA",
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
	t.Run("Invalid composite state", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{"m1": "ALARM"}

		_, err := GetAllTriggeringAlarms(rule, "INVALID_STATE", states, "m1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown composite state")
	})

	t.Run("Missing alarm in states", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{} // m1 missing

		_, err := GetAllTriggeringAlarms(rule, "ALARM", states, "m1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Empty changed alarm is OK", func(t *testing.T) {
		rule := "ALARM('m1')"
		states := map[string]string{"m1": "ALARM"}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", states, "")
		require.NoError(t, err)
		assert.Equal(t, []string{"m1"}, result)
	})

	t.Run("Empty rule", func(t *testing.T) {
		states := map[string]string{"m1": "ALARM"}

		_, err := GetAllTriggeringAlarms("", "ALARM", states, "")
		assert.Error(t, err)
	})
}

// ═══════════════════════════════════════════════════════════
// Real-World Scenarios
// ═══════════════════════════════════════════════════════════

func TestGetAllTriggeringAlarms_RealWorldScenarios(t *testing.T) {
	t.Run("High availability - failover scenario", func(t *testing.T) {
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

	t.Run("Maintenance window - suppression active", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(cpu) AND NOT ALARM(maintenance)")
		childStates := map[string]string{
			"cpu":         "ALARM",
			"maintenance": "ALARM",
		}

		result, err := GetAllTriggeringAlarms(rule, "OK", childStates, "maintenance")
		require.NoError(t, err)

		assert.Empty(t, result)
		t.Logf("In maintenance - alarms suppressed")
	})

	t.Run("Multi-region - partial outage", func(t *testing.T) {
		rule := TransformAlarmRule("ALARM(us-east) OR ALARM(eu-west) OR ALARM(ap-south)")
		childStates := map[string]string{
			"us-east":  "ALARM",
			"eu-west":  "OK",
			"ap-south": "ALARM",
		}

		result, err := GetAllTriggeringAlarms(rule, "ALARM", childStates, "ap-south")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"us-east", "ap-south"}, result)
		t.Logf("Two regions down - partial outage")
	})
	//  scccccccccccccccccccccccccccccccccccccc
	t.Run("Load balancer - at least one backend healthy", func(t *testing.T) {
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

	t.Run("Critical system - all components must be healthy", func(t *testing.T) {
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

	t.Run("Health check with data validation", func(t *testing.T) {
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
}

// ═══════════════════════════════════════════════════════════
// Tests for Cleaner
// ═══════════════════════════════════════════════════════════
