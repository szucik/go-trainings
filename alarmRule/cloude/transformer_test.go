package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════════════════════
// TESTS FOR TransformAlarmRule (Main Function)
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmRule(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Basic transformations
		{
			name:  "Simple alarm with double quotes",
			input: `ALARM("test")`,
			want:  `ALARM('test')`,
		},
		{
			name:  "Simple alarm with single quotes",
			input: `ALARM('test')`,
			want:  `ALARM('test')`,
		},
		{
			name:  "Alarm without quotes",
			input: `ALARM(test)`,
			want:  `ALARM('test')`,
		},

		// Whitespace handling (uses Cleaner)
		{
			name:  "With newlines",
			input: "ALARM(\n\"test1\"\n)",
			want:  `ALARM('test1')`,
		},
		{
			name:  "With tabs",
			input: "ALARM(\t\"test\"\t)",
			want:  `ALARM('test')`,
		},
		{
			name:  "With multiple spaces",
			input: `ALARM("cpu")  AND   ALARM("mem")`,
			want:  `ALARM('cpu') && ALARM('mem')`,
		},
		{
			name:  "Leading/trailing whitespace",
			input: "  ALARM(\"test\")  ",
			want:  `ALARM('test')`,
		},

		// Operators
		{
			name:  "AND operator",
			input: `ALARM("a") AND ALARM("b")`,
			want:  `ALARM('a') && ALARM('b')`,
		},
		{
			name:  "OR operator",
			input: `ALARM("a") OR ALARM("b")`,
			want:  `ALARM('a') || ALARM('b')`,
		},
		{
			name:  "NOT operator",
			input: `NOT ALARM("a")`,
			want:  `!ALARM('a')`,
		},
		{
			name:  "Combined operators",
			input: `ALARM("a") AND NOT ALARM("b")`,
			want:  `ALARM('a') && !ALARM('b')`,
		},

		// Boolean literals
		{
			name:  "TRUE literal",
			input: `TRUE`,
			want:  `true`,
		},
		{
			name:  "FALSE literal",
			input: `FALSE`,
			want:  `false`,
		},
		{
			name:  "Alarm with TRUE",
			input: `ALARM("a") AND TRUE`,
			want:  `ALARM('a') && true`,
		},

		// Quote handling - AWS behavior
		{
			name:  "Inner single quotes preserved",
			input: `OK("'test'")`,
			want:  `OK('\'test\'')`,
		},
		{
			name:  "Apostrophe in middle",
			input: `OK(test'name)`,
			want:  `OK('test\'name')`,
		},
		{
			name:  "Apostrophe at end",
			input: `OK(a')`,
			want:  `OK('a\'')`,
		},
		{
			name:  "Spaces inside quotes preserved",
			input: `OK(" a ")`,
			want:  `OK(' a ')`,
		},

		// Edge cases
		{
			name:  "Parenthesis in alarm name",
			input: `ALARM("a)")`,
			want:  `ALARM('a)')`,
		},
		{
			name:  "Empty alarm name",
			input: `ALARM("")`,
			want:  `ALARM('')`,
		},

		// Multiple functions
		{
			name:  "ALARM function",
			input: `ALARM("test")`,
			want:  `ALARM('test')`,
		},
		{
			name:  "OK function",
			input: `OK("test")`,
			want:  `OK('test')`,
		},
		{
			name:  "INSUFFICIENT_DATA function",
			input: `INSUFFICIENT_DATA("test")`,
			want:  `INSUFFICIENT_DATA('test')`,
		},

		// Complex expressions
		{
			name:  "Complex with parentheses",
			input: `(ALARM("a") OR ALARM("b")) AND NOT ALARM("c")`,
			want:  `(ALARM('a') || ALARM('b')) && !ALARM('c')`,
		},
		{
			name:  "Nested expression",
			input: `((ALARM("a") AND ALARM("b")) OR (OK("c") AND OK("d"))) AND NOT INSUFFICIENT_DATA("e")`,
			want:  `((ALARM('a') && ALARM('b')) || (OK('c') && OK('d'))) && !INSUFFICIENT_DATA('e')`,
		},

		// Real AWS examples
		{
			name:  "Two alarms with AND",
			input: `ALARM("CPUUtilizationTooHigh") AND ALARM("DiskReadOpsTooHigh")`,
			want:  `ALARM('CPUUtilizationTooHigh') && ALARM('DiskReadOpsTooHigh')`,
		},
		{
			name:  "AND with NOT",
			input: `ALARM("CPUUtilizationTooHigh") AND NOT ALARM("DeploymentInProgress")`,
			want:  `ALARM('CPUUtilizationTooHigh') && !ALARM('DeploymentInProgress')`,
		},
		{
			name:  "Nested OR and NOT",
			input: `(ALARM("WebServer1CPU") OR ALARM("WebServer2CPU")) AND NOT ALARM("MaintenanceWindow")`,
			want:  `(ALARM('WebServer1CPU') || ALARM('WebServer2CPU')) && !ALARM('MaintenanceWindow')`,
		},

		// With newlines (formatted AWS responses)
		{
			name:  "Multiline formatted",
			input: "ALARM(\n\"prod-cpu\"\n) AND ALARM(\n\"prod-memory\"\n)",
			want:  `ALARM('prod-cpu') && ALARM('prod-memory')`,
		},
		{
			name:  "Complex multiline",
			input: "(\n  ALARM(\"web-1\") OR\n  ALARM(\"web-2\")\n) AND NOT ALARM(\n  \"maintenance\"\n)",
			want:  `(ALARM('web-1') || ALARM('web-2')) && !ALARM('maintenance')`, // ← Bez spacji wewnątrz ()
		},

		// Normalization edge cases
		{
			name:  "Missing space after closing paren",
			input: `ALARM("a")AND ALARM("b")`,
			want:  `ALARM('a') && ALARM('b')`,
		},
		{
			name:  "Missing space before opening paren",
			input: `ALARM("a") AND(ALARM("b"))`,
			want:  `ALARM('a') && (ALARM('b'))`,
		},
		{
			name:  "NOT without space",
			input: `NOT(ALARM("a"))`,
			want:  `!(ALARM('a'))`,
		},

		// Special characters in names
		{
			name:  "Hyphens in name",
			input: `ALARM("my-alarm-123")`,
			want:  `ALARM('my-alarm-123')`,
		},
		{
			name:  "Underscores in name",
			input: `ALARM("test_alarm_name")`,
			want:  `ALARM('test_alarm_name')`,
		},
		{
			name:  "Dots in name",
			input: `ALARM("prod.web.cpu")`,
			want:  `ALARM('prod.web.cpu')`,
		},
		{
			name:  "Slashes in name",
			input: `ALARM("prod/web/cpu")`,
			want:  `ALARM('prod/web/cpu')`,
		},
		{
			name:  "ARN format",
			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm")`,
			want:  `ALARM('arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm')`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformAlarmRule(tt.input)
			assert.Equal(t, tt.want, got, "Input: %q", tt.input)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// CRITICAL TEST - AWS Compatibility
// ═══════════════════════════════════════════════════════════

func TestCriticalCase_AWSCompatibility(t *testing.T) {
	// This is THE most important test!
	// It proves our transformation produces the same alarm names as AWS.

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

	// AWS CloudWatch has: 'test' (with apostrophes!)
	// Govaluate MUST receive: 'test' (with apostrophes!)
	expectedFromAWS := "'test'"
	assert.Equal(t, expectedFromAWS, receivedArg)
}

// ═══════════════════════════════════════════════════════════
// INTEGRATION TESTS with Govaluate
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmRule_GovaluateIntegration(t *testing.T) {
	testCases := []struct {
		name           string
		awsInput       string
		expectedAlarms []string
		description    string
	}{
		{
			name:           "Simple alarm",
			awsInput:       `ALARM("test-alarm")`,
			expectedAlarms: []string{`test-alarm`},
			description:    "Single alarm with double quotes",
		},
		{
			name:           "Two alarms with AND",
			awsInput:       `ALARM("alarm1") AND ALARM("alarm2")`,
			expectedAlarms: []string{`alarm1`, `alarm2`},
			description:    "Two alarms combined with AND",
		},
		{
			name:           "Complex expression",
			awsInput:       `(OK("health-check") OR ALARM("error-rate")) AND NOT ALARM("maintenance")`,
			expectedAlarms: []string{`health-check`, `maintenance`},
			description:    "Complex expression with OR, AND, NOT",
		},
		{
			name:           "Alarm with inner quotes",
			awsInput:       `OK("'test'")`,
			expectedAlarms: []string{`'test'`},
			description:    "AWS keeps inner single quotes",
		},
		{
			name:           "Alarm with apostrophe in middle",
			awsInput:       `OK(test'name)`,
			expectedAlarms: []string{`test'name`},
			description:    "Apostrophe in the middle of alarm name",
		},
		{
			name:           "Alarm with newlines",
			awsInput:       "ALARM(\n\"cpu-high\"\n) AND ALARM(\n\"memory-high\"\n)",
			expectedAlarms: []string{`cpu-high`, `memory-high`},
			description:    "AWS formatted with newlines",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tc.awsInput)
			t.Logf("AWS Input:    %q", tc.awsInput)
			t.Logf("Transformed:  %q", transformed)

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
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected: %v", tc.expectedAlarms)

			// Verify expected alarms were called
			for i, expected := range tc.expectedAlarms {
				if assert.Less(t, i, len(callLog)) {
					assert.Equal(t, expected, callLog[i])
				}
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// EDGE CASES - AWS Compatibility
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmRule_AllAWSEdgeCases(t *testing.T) {
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

			t.Logf("AWS Input:          %s", tc.awsInput)
			t.Logf("AWS alarm name:     %q", tc.awsAlarmName)
			t.Logf("Transformed:        %s", transformed)
			t.Logf("Govaluate received: %q", receivedArg)

			assert.Equal(t, tc.awsAlarmName, receivedArg)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// TESTS FOR transformAlarmCalls (Internal Function)
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmCalls(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Double quotes to single",
			input: `ALARM("test")`,
			want:  `ALARM('test')`,
		},
		{
			name:  "Single quotes unchanged",
			input: `ALARM('test')`,
			want:  `ALARM('test')`,
		},
		{
			name:  "No quotes",
			input: `ALARM(test)`,
			want:  `ALARM('test')`,
		},
		{
			name:  "Inner quotes escaped",
			input: `OK("'test'")`,
			want:  `OK('\'test\'')`,
		},
		{
			name:  "Apostrophe escaped",
			input: `OK(test'name)`,
			want:  `OK('test\'name')`,
		},
		{
			name:  "Multiple alarms",
			input: `ALARM("a") AND OK("b")`,
			want:  `ALARM('a') AND OK('b')`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformAlarmCalls(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// TESTS FOR transformAlarmContent (Internal Function)
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "Simple string",
			input: `"test"`,
			want:  `'test'`,
		},
		{
			name:  "Remove outer double quotes",
			input: `"test"`,
			want:  `'test'`,
		},
		{
			name:  "Remove outer single quotes",
			input: `'test'`,
			want:  `'test'`,
		},
		{
			name:  "Inner quotes preserved and escaped",
			input: `"'test'"`,
			want:  `'\'test\''`,
		},
		{
			name:  "Apostrophe escaped",
			input: `test'name`,
			want:  `'test\'name'`,
		},
		{
			name:  "Empty string",
			input: ``,
			want:  `''`,
		},
		{
			name:  "Whitespace trimmed",
			input: ` test `,
			want:  `'test'`,
		},
		{
			name:  "Spaces inside quotes preserved",
			input: `" a "`,
			want:  `' a '`,
		},
		{
			name:  "Parenthesis preserved",
			input: `"a)"`,
			want:  `'a)'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformAlarmContent(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// transformer_test.go - Poprawiony test

func TestFindFunctionEnd(t *testing.T) {
	tests := []struct {
		name  string
		input string
		start int
		want  int
	}{
		{
			name:  "Simple case",
			input: `"test")`,
			start: 0,
			want:  6,
		},
		{
			name:  "With parenthesis in name",
			input: `"a)")`,
			start: 0,
			want:  4, // ← Poprawiona wartość (było 3)
		},
		{
			name:  "Multiple closing parens",
			input: `"test"))`,
			start: 0,
			want:  6,
		},
		{
			name:  "Before AND operator",
			input: `"test") AND`,
			start: 0,
			want:  6,
		},
		{
			name:  "Before OR operator",
			input: `"test") OR`,
			start: 0,
			want:  6,
		},
		{
			name:  "End of string",
			input: `"test")`,
			start: 0,
			want:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findFunctionEnd(tt.input, tt.start)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ═══════════════════════════════════════════════════════════
// REAL-WORLD AWS SCENARIOS
// ═══════════════════════════════════════════════════════════

func TestTransformAlarmRule_RealWorldScenarios(t *testing.T) {
	testCases := []struct {
		name     string
		scenario string
		awsInput string
		want     string
	}{
		{
			name:     "High availability",
			scenario: "Alert when both primary and backup servers are down",
			awsInput: `ALARM("primary-server-down") AND ALARM("backup-server-down")`,
			want:     `ALARM('primary-server-down') && ALARM('backup-server-down')`,
		},
		{
			name:     "Maintenance window",
			scenario: "Alert only when not in maintenance window",
			awsInput: `ALARM("cpu-high") AND NOT ALARM("maintenance-window")`,
			want:     `ALARM('cpu-high') && !ALARM('maintenance-window')`,
		},
		{
			name:     "Multi-region",
			scenario: "Alert when any region has issues",
			awsInput: `ALARM("us-east-1-down") OR ALARM("eu-west-1-down") OR ALARM("ap-southeast-1-down")`,
			want:     `ALARM('us-east-1-down') || ALARM('eu-west-1-down') || ALARM('ap-southeast-1-down')`,
		},
		{
			name:     "Service dependency",
			scenario: "Alert when service is down and database is healthy",
			awsInput: `ALARM("service-down") AND OK("database-healthy")`,
			want:     `ALARM('service-down') && OK('database-healthy')`,
		},
		{
			name:     "Complex production",
			scenario: "Alert for production issues excluding known problems",
			awsInput: `(ALARM("prod-cpu-high") OR ALARM("prod-memory-high")) AND NOT (ALARM("known-issue-123") OR ALARM("planned-deployment"))`,
			want:     `(ALARM('prod-cpu-high') || ALARM('prod-memory-high')) && !(ALARM('known-issue-123') || ALARM('planned-deployment'))`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Scenario: %s", tc.scenario)
			got := TransformAlarmRule(tc.awsInput)
			assert.Equal(t, tc.want, got)
		})
	}
}
