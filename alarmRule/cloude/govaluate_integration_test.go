package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCriticalCase - weryfikacja czy escape'owanie działa poprawnie
func Test_CriticalCase(t *testing.T) {
	// Arrange
	input := `OK("'test'")`

	t.Logf("AWS Input: %s", input)
	t.Logf("AWS alarm name będzie: 'test' (z apostrofami)")

	// Act
	transformed := TransformAlarmRule(input)
	t.Logf("Nasza transformacja: %s", transformed)

	// Govaluate verification
	var receivedArg string
	functions := map[string]govaluate.ExpressionFunction{
		"OK": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args, "OK should receive arguments")
			receivedArg = args[0].(string)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err, "Govaluate should parse successfully")

	_, err = expr.Evaluate(nil)
	require.NoError(t, err, "Govaluate should evaluate successfully")

	// Assert
	t.Logf("Govaluate otrzymał: %q", receivedArg)

	expectedFromAWS := "'test'" // To jest nazwa alarmu w AWS!
	assert.Equal(t, expectedFromAWS, receivedArg, "Govaluate should receive exactly what AWS has")
	t.Logf("✅ ZGODNE Z AWS! Govaluate dostał dokładnie to co AWS: %q", receivedArg)
}

// TestAllEdgeCasesWithAWS - porównanie z AWS dla wszystkich edge cases
func Test_AllEdgeCasesWithAWS(t *testing.T) {
	testCases := []struct {
		awsInput     string
		awsAlarmName string // Co AWS ma jako nazwę alarmu
		description  string
	}{
		{
			awsInput:     `ALARM("test")`,
			awsAlarmName: `test`,
			description:  "Simple double quotes",
		},
		{
			awsInput:     `OK('test')`,
			awsAlarmName: `test`,
			description:  "Simple single quotes",
		},
		{
			awsInput:     `OK("'test'")`,
			awsAlarmName: `'test'`,
			description:  "Double quotes with single quotes inside - AWS keeps inner",
		},
		{
			awsInput:     `OK(test'name)`,
			awsAlarmName: `test'name`,
			description:  "Apostrophe in middle - no surrounding quotes",
		},
		{
			awsInput:     `OK(a')`,
			awsAlarmName: `a'`,
			description:  "Apostrophe at end - no surrounding quotes",
		},
		{
			awsInput:     `OK("a)")`,
			awsAlarmName: `a)`,
			description:  "Parenthesis inside quotes",
		},
		{
			awsInput:     `ALARM('a)')`,
			awsAlarmName: `a)`,
			description:  "Parenthesis after apostrophe",
		},
		{
			awsInput:     `OK(" a ")`,
			awsAlarmName: ` a `,
			description:  "Spaces inside quotes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Act
			transformed := TransformAlarmRule(tc.awsInput)

			// Govaluate verification
			var receivedArg string
			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK should receive arguments")
					receivedArg = args[0].(string)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					receivedArg = args[0].(string)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Govaluate should parse expression")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Govaluate should evaluate successfully")

			// Assert
			t.Logf("AWS Input:          %s", tc.awsInput)
			t.Logf("AWS alarm name:     %q", tc.awsAlarmName)
			t.Logf("Transformed:        %s", transformed)
			t.Logf("Govaluate received: %q", receivedArg)

			assert.Equal(t, tc.awsAlarmName, receivedArg, "Govaluate should receive exactly what AWS has")
			t.Logf("✅ MATCH! Govaluate otrzymał dokładnie to co AWS ma w nazwie alarmu")
		})
	}
}

// TestFullPipeline - testuje pełny pipeline: AWS → Transform → Govaluate
func Test_FullPipeline(t *testing.T) {
	testCases := []struct {
		name          string
		awsInput      string
		expectedCalls []string
		description   string
	}{
		{
			name:          "Simple alarm",
			awsInput:      `ALARM("DobryAlarm")`,
			expectedCalls: []string{`DobryAlarm`},
			description:   "Single alarm with double quotes",
		},
		{
			name:          "Alarm with parenthesis in name",
			awsInput:      `OK("a)")`,
			expectedCalls: []string{`a)`},
			description:   "Alarm name contains closing parenthesis",
		},
		{
			name:          "Two alarms with AND",
			awsInput:      `ALARM("Alarm1") AND ALARM("Alarm2")`,
			expectedCalls: []string{`Alarm1`, `Alarm2`},
			description:   "Two alarms combined with AND",
		},
		{
			name:          "Complex expression",
			awsInput:      `(OK("Health") OR ALARM("Error")) AND NOT ALARM("Maintenance")`,
			expectedCalls: []string{`Health`, `Maintenance`},
			description:   "Complex expression with OR, AND, NOT",
		},
		{
			name:          "Alarm with inner quotes - AWS keeps them",
			awsInput:      `OK("'test'")`,
			expectedCalls: []string{`'test'`},
			description:   "AWS keeps inner single quotes when outer are double quotes",
		},
		{
			name:          "Alarm with apostrophe in middle",
			awsInput:      `OK(test'name)`,
			expectedCalls: []string{`test'name`},
			description:   "Apostrophe in the middle of alarm name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Logf("Description: %s", tc.description)
			t.Logf("AWS Input: %s", tc.awsInput)

			// Act
			transformed := TransformAlarmRule(tc.awsInput)
			t.Logf("Transformed: %s", transformed)

			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK should receive arguments")
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					t.Logf("OK() called with: %q", alarmName)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					t.Logf("ALARM() called with: %q", alarmName)
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "INSUFFICIENT_DATA should receive arguments")
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					t.Logf("INSUFFICIENT_DATA() called with: %q", alarmName)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should create govaluate expression")

			result, err := expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("Result: %v", result)
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected: %v", tc.expectedCalls)

			// Verify we got expected calls (accounting for short-circuit)
			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog), "Should have call at index %d", i) {
					assert.Equal(t, expected, callLog[i], "Call %d should match expected", i)
				}
			}
		})
	}
}

// TestRealAWSCompositeAlarm - test z rzeczywistym AWS composite alarm
func Test_RealAWSCompositeAlarm(t *testing.T) {
	// Arrange - Dokładnie to co zwraca AWS DescribeAlarms
	awsResponse := `ALARM("alarm1") AND ALARM("alarm2")`

	t.Logf("AWS Response: %s", awsResponse)

	// Act
	transformed := TransformAlarmRule(awsResponse)
	t.Logf("Transformed:  %s", transformed)

	// Assert transformation
	expected := `ALARM('alarm1') && ALARM('alarm2')`
	assert.Equal(t, expected, transformed, "Transformation should match expected")

	// Weryfikacja z govaluate
	callLog := []string{}

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args, "ALARM should receive arguments")
			alarmName := args[0].(string)
			callLog = append(callLog, alarmName)
			t.Logf("ALARM() called with: %q", alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err, "Govaluate should parse successfully")

	result, err := expr.Evaluate(nil)
	require.NoError(t, err, "Govaluate should evaluate successfully")

	// Assert
	assert.True(t, result.(bool), "Result should be true")
	t.Logf("Result: %v", result)
	t.Logf("Alarm names received: %v", callLog)

	// Weryfikacja że dostaliśmy prawidłowe nazwy alarmów
	expectedAlarms := []string{"alarm1", "alarm2"}
	assert.Equal(t, expectedAlarms, callLog, "Should receive expected alarm names")

	t.Logf("✅ SUCCESS! AWS composite alarm correctly transformed and evaluated")
}

// TestRealWorldFormattedAlarmRule - rzeczywisty przykład z ładnym formatowaniem
func Test_RealWorldFormattedAlarmRule(t *testing.T) {
	// Arrange
	cpuAlarm := "production-cpu-high"
	memoryAlarm := "production-memory-high"
	diskAlarm := "production-disk-full"

	alarmRule := "ALARM(\n" +
		cpuAlarm + "\n" +
		") AND ALARM(\n" +
		memoryAlarm + "\n" +
		") OR ALARM(\n" +
		diskAlarm + "\n" +
		")"

	t.Logf("Formatted AlarmRule:\n%s", alarmRule)
	t.Logf("Raw: %q", alarmRule)

	// Act
	transformed := TransformAlarmRule(alarmRule)
	t.Logf("Transformed: %s", transformed)

	var callLog []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args, "ALARM should receive arguments")
			name := args[0].(string)
			callLog = append(callLog, name)
			t.Logf("ALARM(%q)", name)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err, "Should parse formatted expression")

	result, err := expr.Evaluate(nil)
	require.NoError(t, err, "Should evaluate successfully")

	// Assert
	assert.True(t, result.(bool), "Result should be true")
	assert.NotEmpty(t, callLog, "Should call at least one alarm function")

	t.Logf("Result: %v", result)
	t.Logf("Alarms evaluated: %v", callLog)

	// Short-circuit: pierwszy ALARM true && drugi true = true, OR nie sprawdza trzeciego
	assert.GreaterOrEqual(t, len(callLog), 1, "Should evaluate at least first alarm")
	assert.Equal(t, cpuAlarm, callLog[0], "First alarm should be CPU alarm")
}

// TestAWSMultilineFormatting - testy dla różnych formatowań multiline
func TestAWSMultilineFormatting(t *testing.T) {
	testCases := []struct {
		name          string
		alarmRule     string
		expectedCalls []string
		description   string
	}{
		{
			name: "Newlines around alarm names",
			alarmRule: "ALARM(\n" +
				"\"cpu-alarm\"\n" +
				") AND ALARM(\n" +
				"\"mem-alarm\"\n" +
				")",
			expectedCalls: []string{"cpu-alarm", "mem-alarm"},
			description:   "Newlines before and after alarm names should be trimmed",
		},
		{
			name: "Tabs and newlines",
			alarmRule: "ALARM(\n\t" +
				"\"server-1\"\n\t" +
				") OR ALARM(\n\t" +
				"\"server-2\"\n\t" +
				")",
			expectedCalls: []string{"server-1"},
			description:   "Tabs and newlines should be handled correctly",
		},
		{
			name: "Complex multiline expression",
			alarmRule: "(\n" +
				"  ALARM(\"web-1\") OR\n" +
				"  ALARM(\"web-2\")\n" +
				") AND NOT ALARM(\n" +
				"  \"maintenance\"\n" +
				")",
			expectedCalls: []string{"web-1", "maintenance"},
			description:   "Complex multiline with indentation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Logf("Description: %s", tc.description)
			t.Logf("AlarmRule:\n%s", tc.alarmRule)

			// Act
			transformed := TransformAlarmRule(tc.alarmRule)
			t.Logf("Transformed: %s", transformed)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("ALARM(%q)", name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse expression")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected:  %v", tc.expectedCalls)

			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog), "Should have call at index %d", i) {
					assert.Equal(t, expected, callLog[i], "Call %d should match", i)
				}
			}
		})
	}
}

// TestAWSCompositeAlarmScenarios - różne scenariusze composite alarmów z AWS
func TestAWSCompositeAlarmScenarios(t *testing.T) {
	testCases := []struct {
		name          string
		scenario      string
		alarmRule     string
		expectedCalls []string
	}{
		{
			name:          "High availability scenario",
			scenario:      "Alert when both primary and backup servers are down",
			alarmRule:     `ALARM("primary-server-down") AND ALARM("backup-server-down")`,
			expectedCalls: []string{"primary-server-down", "backup-server-down"},
		},
		{
			name:          "Maintenance window scenario",
			scenario:      "Alert only when not in maintenance window",
			alarmRule:     `ALARM("cpu-high") AND NOT ALARM("maintenance-window")`,
			expectedCalls: []string{"cpu-high", "maintenance-window"},
		},
		{
			name:          "Multi-region scenario",
			scenario:      "Alert when any region has issues",
			alarmRule:     `ALARM("us-east-1-down") OR ALARM("eu-west-1-down") OR ALARM("ap-southeast-1-down")`,
			expectedCalls: []string{"us-east-1-down"},
		},
		{
			name:          "Service dependency scenario",
			scenario:      "Alert when service is down and database is healthy",
			alarmRule:     `ALARM("service-down") AND OK("database-healthy")`,
			expectedCalls: []string{"service-down", "database-healthy"},
		},
		{
			name:          "Complex production scenario",
			scenario:      "Alert for production issues excluding known problems",
			alarmRule:     `(ALARM("prod-cpu-high") OR ALARM("prod-memory-high")) AND NOT (ALARM("known-issue-123") OR ALARM("planned-deployment"))`,
			expectedCalls: []string{"prod-cpu-high", "known-issue-123"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Logf("Scenario: %s", tc.scenario)
			t.Logf("AlarmRule: %s", tc.alarmRule)

			// Act
			transformed := TransformAlarmRule(tc.alarmRule)
			t.Logf("Transformed: %s", transformed)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK should receive arguments")
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("OK(%q)", name)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("ALARM(%q)", name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse expression")

			result, err := expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("Result: %v", result)
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected: %v", tc.expectedCalls)

			// Verify expected calls (accounting for short-circuit)
			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog), "Should have call at index %d", i) {
					assert.Equal(t, expected, callLog[i], "Call %d should match", i)
				}
			}
		})
	}
}

// TestAWSAlarmNamingConventions - test różnych konwencji nazewnictwa alarmów w AWS
func TestAWSAlarmNamingConventions(t *testing.T) {
	testCases := []struct {
		name         string
		alarmName    string
		convention   string
		expectedName string
	}{
		{
			name:         "Kebab case",
			alarmName:    "my-production-cpu-alarm",
			convention:   "kebab-case",
			expectedName: "my-production-cpu-alarm",
		},
		{
			name:         "Snake case",
			alarmName:    "production_memory_high",
			convention:   "snake_case",
			expectedName: "production_memory_high",
		},
		{
			name:         "Camel case",
			alarmName:    "productionCpuHigh",
			convention:   "camelCase",
			expectedName: "productionCpuHigh",
		},
		{
			name:         "Dot notation",
			alarmName:    "prod.web.cpu.high",
			convention:   "dot.notation",
			expectedName: "prod.web.cpu.high",
		},
		{
			name:         "Slash notation (path-like)",
			alarmName:    "prod/us-east-1/web/cpu",
			convention:   "path/notation",
			expectedName: "prod/us-east-1/web/cpu",
		},
		{
			name:         "ARN format",
			alarmName:    "arn:aws:cloudwatch:region:account:alarm:name",
			convention:   "ARN",
			expectedName: "arn:aws:cloudwatch:region:account:alarm:name",
		},
		{
			name:         "Mixed conventions",
			alarmName:    "prod_web-server.cpu:high",
			convention:   "mixed",
			expectedName: "prod_web-server.cpu:high",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			alarmRule := `ALARM("` + tc.alarmName + `")`
			t.Logf("Convention: %s", tc.convention)
			t.Logf("AlarmRule: %s", alarmRule)

			// Act
			transformed := TransformAlarmRule(alarmRule)
			t.Logf("Transformed: %s", transformed)

			var receivedName string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					receivedName = args[0].(string)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse expression")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.Equal(t, tc.expectedName, receivedName, "Should preserve naming convention")
			t.Logf("✅ Naming convention preserved: %q", receivedName)
		})
	}
}
