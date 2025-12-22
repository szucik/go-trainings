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
			name:           "NOT INSUFFICIENT_DATA - alarm lacks data (ALARM state)",
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
		{
			name:           "Complex - three conditions",
			awsRule:        "OK(a) AND NOT ALARM(b) AND NOT INSUFFICIENT_DATA(c)",
			compositeState: "OK",
			childStates: map[string]string{
				"a": "OK",
				"b": "OK",
				"c": "ALARM",
			},
			wantTriggering: []string{"a", "b", "c"},
			description:    "All in expected states (a=OK, b not ALARM, c not INSUFFICIENT_DATA)",
		},
		{
			name:           "INSUFFICIENT_DATA as composite state",
			awsRule:        "ALARM(m1)",
			compositeState: "INSUFFICIENT_DATA",
			childStates: map[string]string{
				"m1": "INSUFFICIENT_DATA",
			},
			wantTriggering: []string{"m1"},
			description:    "Composite is INSUFFICIENT_DATA, m1 has insufficient data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsRule)

			t.Logf("Description: %s", tt.description)
			t.Logf("AWS Rule: %s", tt.awsRule)
			t.Logf("Transformed: %s", transformed)
			t.Logf("Composite State: %s", tt.compositeState)
			t.Logf("Child States: %v", tt.childStates)

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

func TestGetAllTriggeringAlarms_ComplexMixedStates(t *testing.T) {
	t.Run("All three state types in rule", func(t *testing.T) {
		rule := TransformAlarmRule(
			"(OK(health) AND NOT ALARM(cpu)) OR INSUFFICIENT_DATA(network)",
		)

		// Scenario 1: health and cpu satisfy first condition
		states1 := map[string]string{
			"health":  "OK",
			"cpu":     "OK",
			"network": "OK",
		}

		result1, err := GetAllTriggeringAlarms(rule, "OK", states1, "")
		require.NoError(t, err)

		t.Logf("Scenario 1 - First condition satisfied:")
		t.Logf("  States: %v", states1)
		t.Logf("  Triggering: %v", result1)
		assert.ElementsMatch(t, []string{"health", "cpu"}, result1)

		// Scenario 2: network has insufficient data (second condition)
		states2 := map[string]string{
			"health":  "ALARM",
			"cpu":     "ALARM",
			"network": "INSUFFICIENT_DATA",
		}

		result2, err := GetAllTriggeringAlarms(rule, "OK", states2, "")
		require.NoError(t, err)

		t.Logf("Scenario 2 - Second condition satisfied:")
		t.Logf("  States: %v", states2)
		t.Logf("  Triggering: %v", result2)
		assert.ElementsMatch(t, []string{"network"}, result2)
	})

	t.Run("Real world - health check with data validation", func(t *testing.T) {
		rule := TransformAlarmRule(
			"OK(api) AND OK(database) AND NOT INSUFFICIENT_DATA(metrics)",
		)

		states := map[string]string{
			"api":      "OK",
			"database": "OK",
			"metrics":  "OK",
		}

		result, err := GetAllTriggeringAlarms(rule, "OK", states, "")
		require.NoError(t, err)

		assert.ElementsMatch(t, []string{"api", "database", "metrics"}, result)
		t.Logf("All systems healthy and have sufficient data")
	})
}
```

---

## 📊 Tabela Prawdy - Wszystkie Funkcje
```
┌──────────────────────────────────────────────────────────────────────────┐
│ Funkcja w AlarmRule          │ Composite OK gdy │ Triggering gdy       │
├──────────────────────────────┼──────────────────┼──────────────────────┤
│ OK(alarm)                    │ alarm = OK       │ alarm = OK           │
│ ALARM(alarm)                 │ alarm = ALARM    │ N/A (dla ALARM comp) │
│ INSUFFICIENT_DATA(alarm)     │ alarm = INSUF    │ alarm = INSUF        │
│ NOT OK(alarm)                │ alarm ≠ OK       │ alarm ≠ OK           │
│ NOT ALARM(alarm)             │ alarm ≠ ALARM    │ alarm ≠ ALARM        │
│ NOT INSUFFICIENT_DATA(alarm) │ alarm ≠ INSUF    │ alarm ≠ INSUF        │
└──────────────────────────────┴──────────────────┴──────────────────────┘
```

---

## ✅ Podsumowanie
```
✅ Kod obsługuje WSZYSTKIE stany:
   - OK
   - ALARM
   - INSUFFICIENT_DATA

✅ Kod obsługuje WSZYSTKIE operatory:
   - OK(x)
   - ALARM(x)
   - INSUFFICIENT_DATA(x)
   - NOT OK(x)
   - NOT ALARM(x)
   - NOT INSUFFICIENT_DATA(x)

✅ Kod obsługuje WSZYSTKIE composite states:
   - Composite = ALARM
   - Composite = OK
   - Composite = INSUFFICIENT_DATA

✅ Logika jest poprawna dla:
   - OR logic (OK(m1) OR OK(m2))
   - AND logic (OK(m1) AND OK(m2))
   - Complex logic (mix wszystkiego)