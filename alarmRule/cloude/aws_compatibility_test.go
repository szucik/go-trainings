package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCriticalCase verifies that escaping works correctly and matches AWS behavior exactly.
// This is the most important test - it proves our transformation produces the same alarm names as AWS.
func TestCriticalCase(t *testing.T) {
	input := `OK("'test'")`

	transformed := TransformAlarmRule(input)

	var receivedArg string
	functions := map[string]govaluate.ExpressionFunction{
		"OK": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args)
			receivedArg = args[0].(string)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err)

	_, err = expr.Evaluate(nil)
	require.NoError(t, err)

	expectedFromAWS := "'test'"
	assert.Equal(t, expectedFromAWS, receivedArg)
}

// TestAllEdgeCasesWithAWS compares our transformation with AWS behavior for all edge cases.
// Each test case represents actual AWS CloudWatch alarm name handling.
func TestAllEdgeCasesWithAWS(t *testing.T) {
	testCases := []struct {
		awsInput     string
		awsAlarmName string
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
			transformed := TransformAlarmRule(tc.awsInput)

			var receivedArg string
			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedArg = args[0].(string)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedArg = args[0].(string)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			assert.Equal(t, tc.awsAlarmName, receivedArg)
		})
	}
}

// TestFullPipeline tests the complete pipeline: AWS input → Transform → Govaluate evaluation.
// Verifies that complex expressions with multiple alarms work correctly end-to-end.
func TestFullPipeline(t *testing.T) {
	testCases := []struct {
		name          string
		awsInput      string
		expectedCalls []string
		description   string
	}{
		{
			name:          "Simple alarm",
			awsInput:      `ALARM("test-alarm")`,
			expectedCalls: []string{`test-alarm`},
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
			awsInput:      `ALARM("alarm1") AND ALARM("alarm2")`,
			expectedCalls: []string{`alarm1`, `alarm2`},
			description:   "Two alarms combined with AND",
		},
		{
			name:          "Complex expression",
			awsInput:      `(OK("health-check") OR ALARM("error-rate")) AND NOT ALARM("maintenance")`,
			expectedCalls: []string{`health-check`, `maintenance`},
			description:   "Complex expression with OR, AND, NOT",
		},
		{
			name:          "Alarm with inner quotes",
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
			transformed := TransformAlarmRule(tc.awsInput)

			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					alarmName := args[0].(string)
					callLog = append(callLog, alarmName)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			result, err := expr.Evaluate(nil)
			require.NoError(t, err)

			assert.NotNil(t, result)

			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog)) {
					assert.Equal(t, expected, callLog[i])
				}
			}
		})
	}
}

// TestRealAWSCompositeAlarm tests with exact AWS DescribeAlarms response format.
// This verifies compatibility with real AWS CloudWatch composite alarm responses.
func TestRealAWSCompositeAlarm(t *testing.T) {
	awsResponse := `ALARM("alarm1") AND ALARM("alarm2")`

	transformed := TransformAlarmRule(awsResponse)

	expected := `ALARM('alarm1') && ALARM('alarm2')`
	assert.Equal(t, expected, transformed)

	callLog := []string{}

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args)
			alarmName := args[0].(string)
			callLog = append(callLog, alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err)

	result, err := expr.Evaluate(nil)
	require.NoError(t, err)

	assert.True(t, result.(bool))

	expectedAlarms := []string{"alarm1", "alarm2"}
	assert.Equal(t, expectedAlarms, callLog)
}

// TestRealWorldFormattedAlarmRule tests AlarmRule with formatted output (newlines, indentation).
// AWS API responses sometimes include formatted alarm rules with whitespace.
func TestRealWorldFormattedAlarmRule(t *testing.T) {
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

	transformed := TransformAlarmRule(alarmRule)

	var callLog []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args)
			name := args[0].(string)
			callLog = append(callLog, name)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err)

	result, err := expr.Evaluate(nil)
	require.NoError(t, err)

	assert.True(t, result.(bool))
	assert.NotEmpty(t, callLog)

	assert.GreaterOrEqual(t, len(callLog), 1)
	assert.Equal(t, cpuAlarm, callLog[0])
}

// TestAWSMultilineFormatting tests various multiline formatting patterns from AWS API.
// AWS can return alarm rules with different whitespace formatting.
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
			transformed := TransformAlarmRule(tc.alarmRule)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					name := args[0].(string)
					callLog = append(callLog, name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog)) {
					assert.Equal(t, expected, callLog[i])
				}
			}
		})
	}
}

// TestAWSCompositeAlarmScenarios tests real-world AWS composite alarm scenarios.
// Each scenario represents a common production use case.
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
			transformed := TransformAlarmRule(tc.alarmRule)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					name := args[0].(string)
					callLog = append(callLog, name)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					name := args[0].(string)
					callLog = append(callLog, name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			result, err := expr.Evaluate(nil)
			require.NoError(t, err)

			assert.NotNil(t, result)

			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog)) {
					assert.Equal(t, expected, callLog[i])
				}
			}
		})
	}
}

// TestAWSAlarmNamingConventions tests various AWS alarm naming conventions.
// AWS alarms can use different naming patterns - all must be preserved exactly.
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
			alarmRule := `ALARM("` + tc.alarmName + `")`

			transformed := TransformAlarmRule(alarmRule)

			var receivedName string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedName = args[0].(string)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedName, receivedName)
		})
	}
}
