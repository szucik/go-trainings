package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGovaluateEscaping - test czy govaluate poprawnie interpretuje nasze escape'y
func Test_GovaluateEscaping(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		expected   string
	}{
		{
			name:       "Simple string",
			expression: "OK('test')",
			expected:   "test",
		},
		{
			name:       "Escaped apostrophe",
			expression: `OK('test\'s')`,
			expected:   "test's",
		},
		{
			name:       "Escaped quotes",
			expression: `OK('\"test\"')`,
			expected:   `"test"`,
		},
		{
			name:       "Multiple escaped apostrophes",
			expression: `OK('\'test\'')`,
			expected:   "'test'",
		},
		{
			name:       "Apostrophe at end",
			expression: `OK('test\'')`,
			expected:   "test'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var received string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK function should receive arguments")
					received = args[0].(string)
					return true, nil
				},
			}

			// Act
			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.expression, functions)
			require.NoError(t, err, "Should parse expression successfully")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.Equal(t, tc.expected, received, "Received value should match expected")
			t.Logf("✅ Expression: %s → %q", tc.expression, received)
		})
	}
}

// TestAllTransformedWithGovaluate - testuje czy WSZYSTKIE nasze transformacje działają z govaluate
func Test_AllTransformedWithGovaluate(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{"OK(a)", "OK(a)", "OK('a')"},
		{"ALARM(a)", "ALARM(a)", "ALARM('a')"},
		{"INSUFFICIENT_DATA(a)", "INSUFFICIENT_DATA(a)", "INSUFFICIENT_DATA('a')"},
		{"NOT OK(a)", "NOT OK(a)", "!OK('a')"},
		{"NOT ALARM(a)", "NOT ALARM(a)", "!ALARM('a')"},
		{"NOT INSUFFICIENT_DATA(a)", "NOT INSUFFICIENT_DATA(a)", "!INSUFFICIENT_DATA('a')"},

		{`OK("a")`, `OK("a")`, `OK('a')`},
		{"OK('a')", "OK('a')", "OK('a')"},
		{"OK(a')", "OK(a')", `OK('a\'')`},
		{`OK("'a'")`, `OK("'a'")`, `OK('\'a\'')`},
		{`OK("a)")`, `OK("a)")`, `OK('a)')`},
		{"OK('a)')", "OK('a)')", "OK('a)')"},
		{"OK(test'name)", "OK(test'name)", `OK('test\'name')`},

		{"OK( a )", "OK( a )", "OK('a')"},
		{`OK( "a" )`, `OK( "a" )`, `OK('a')`},
		{`OK(" a ")`, `OK(" a ")`, `OK(' a ')`},

		{"OK(a) AND ALARM(b)", "OK(a) AND ALARM(b)", "OK('a') && ALARM('b')"},
		{"NOT OK(a) AND ALARM(b)", "NOT OK(a) AND ALARM(b)", "!OK('a') && ALARM('b')"},
		{"OK(a) OR ALARM(b)", "OK(a) OR ALARM(b)", "OK('a') || ALARM('b')"},
		{"NOT (OK(a) AND ALARM(b))", "NOT (OK(a) AND ALARM(b))", "!(OK('a') && ALARM('b'))"},

		{`ALARM("DobryAlarm") `, `ALARM("DobryAlarm") `, `ALARM('DobryAlarm')`},
		{`ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`, `ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`, `ALARM('DobryAlarm') && ALARM('dobryAlarm2')`},
		{"ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)", "ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)", `ALARM('DobryAlarm') || ALARM('dobryAlarm2')`},
		{`(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`, `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`, `(ALARM('CPU1') || ALARM('CPU2')) && !ALARM('Deploying')`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := TransformAlarmRule(tc.input)

			// Assert
			assert.Equal(t, tc.want, got, "Transformation should match expected")

			// Verify with govaluate
			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK should receive arguments")
					callLog = append(callLog, "OK("+args[0].(string)+")")
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					callLog = append(callLog, "ALARM("+args[0].(string)+")")
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "INSUFFICIENT_DATA should receive arguments")
					callLog = append(callLog, "INSUFFICIENT_DATA("+args[0].(string)+")")
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.want, functions)
			require.NoError(t, err, "Govaluate should parse expression")

			result, err := expr.Evaluate(nil)
			require.NoError(t, err, "Govaluate should evaluate successfully")

			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("✅ Input: %s → Transformed: %s → Result: %v, Calls: %v", tc.input, tc.want, result, callLog)
		})
	}
}

// TestGovaluateReceivesCorrectArguments - sprawdza dokładnie co govaluate otrzymuje jako argumenty
func Test_GovaluateReceivesCorrectArguments(t *testing.T) {
	testCases := []struct {
		name             string
		transformed      string
		expectedFunction string
		expectedArg      string
	}{
		{
			name:             "Simple name",
			transformed:      `OK('test')`,
			expectedFunction: "OK",
			expectedArg:      "test",
		},
		{
			name:             "Name without quotes - like AWS",
			transformed:      `ALARM('DobryAlarm')`,
			expectedFunction: "ALARM",
			expectedArg:      `DobryAlarm`,
		},
		{
			name:             "Name with parenthesis",
			transformed:      `OK('a)')`,
			expectedFunction: "OK",
			expectedArg:      `a)`,
		},
		{
			name:             "Name with escaped apostrophe",
			transformed:      `OK('a\'')`,
			expectedFunction: "OK",
			expectedArg:      `a'`,
		},
		{
			name:             "Name with apostrophes inside - like AWS",
			transformed:      `OK('\'test\'')`,
			expectedFunction: "OK",
			expectedArg:      `'test'`,
		},
		{
			name:             "Name with apostrophe in middle",
			transformed:      `OK('test\'name')`,
			expectedFunction: "OK",
			expectedArg:      `test'name`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var receivedArg interface{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK should receive arguments")
					receivedArg = args[0]
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					receivedArg = args[0]
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "INSUFFICIENT_DATA should receive arguments")
					receivedArg = args[0]
					return true, nil
				},
			}

			// Act
			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.transformed, functions)
			require.NoError(t, err, "Should parse expression")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			receivedStr, ok := receivedArg.(string)
			require.True(t, ok, "Received arg should be string, got %T", receivedArg)

			assert.Equal(t, tc.expectedArg, receivedStr, "%s should receive correct argument", tc.expectedFunction)
			t.Logf("✅ %s received correct arg: %q", tc.expectedFunction, receivedStr)
		})
	}
}

// TestAWSStringConcatenation - test dla dokładnie takiej konstrukcji jak w AWS SDK
func Test_AWSStringConcatenation(t *testing.T) {
	// Arrange
	alarm1 := "cpu-high-alarm"
	alarm2 := "memory-high-alarm"
	alarmRule := "ALARM(\"" + alarm1 + "\") AND ALARM(\"" + alarm2 + "\")"

	t.Logf("Constructed AlarmRule: %s", alarmRule)

	// Act
	transformed := TransformAlarmRule(alarmRule)

	// Assert
	expectedTransformed := "ALARM('cpu-high-alarm') && ALARM('memory-high-alarm')"
	assert.Equal(t, expectedTransformed, transformed, "Transformation should match expected format")

	// Govaluate verification
	var receivedAlarms []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args, "ALARM function should receive arguments")
			alarmName := args[0].(string)
			receivedAlarms = append(receivedAlarms, alarmName)
			t.Logf("ALARM() called with: %q", alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err, "Govaluate should parse the expression without error")

	result, err := expr.Evaluate(nil)
	require.NoError(t, err, "Govaluate should evaluate without error")

	// Assertions
	assert.True(t, result.(bool), "Result should be true (both alarms return true)")
	assert.Len(t, receivedAlarms, 2, "Should call ALARM function twice")
	assert.Equal(t, []string{"cpu-high-alarm", "memory-high-alarm"}, receivedAlarms, "Should receive correct alarm names in order")
}

// TestAWSStringConcatenationWithNewlines - test dla AlarmRule z \n
func Test_AWSStringConcatenationWithNewlines(t *testing.T) {
	// Arrange
	alarm1 := "cpu-high"
	alarm2 := "memory-high"
	alarmRule := "ALARM(\n" + alarm1 + " \n) AND ALARM(\n" + alarm2 + "\n)"

	t.Logf("Constructed AlarmRule:\n%s", alarmRule)
	t.Logf("Raw: %q", alarmRule)

	// Act
	transformed := TransformAlarmRule(alarmRule)
	t.Logf("Transformed: %s", transformed)

	// Assert
	expectedTransformed := "ALARM('cpu-high') && ALARM('memory-high')"
	assert.Equal(t, expectedTransformed, transformed, "Newlines should be trimmed correctly")

	// Govaluate verification
	var receivedAlarms []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args, "ALARM function should receive arguments")
			alarmName := args[0].(string)
			receivedAlarms = append(receivedAlarms, alarmName)
			t.Logf("ALARM() called with: %q", alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err, "Govaluate should parse the expression")

	result, err := expr.Evaluate(nil)
	require.NoError(t, err, "Govaluate should evaluate successfully")

	// Assertions
	assert.True(t, result.(bool), "Result should be true")
	assert.Equal(t, []string{"cpu-high", "memory-high"}, receivedAlarms, "Should receive trimmed alarm names")
}

// TestMultipleAWSStringConcatenationPatterns - różne warianty konstrukcji AlarmRule
func Test_MultipleAWSStringConcatenationPatterns(t *testing.T) {
	testCases := []struct {
		name          string
		buildRule     func() string
		expectedCalls []string
		description   string
	}{
		{
			name: "Two alarms with AND",
			buildRule: func() string {
				alarm1 := "alarm-1"
				alarm2 := "alarm-2"
				return "ALARM(\"" + alarm1 + "\") AND ALARM(\"" + alarm2 + "\")"
			},
			expectedCalls: []string{"alarm-1", "alarm-2"},
			description:   "Both alarms should be evaluated with AND operator",
		},
		{
			name: "Three alarms with OR - short circuit",
			buildRule: func() string {
				a1 := "web-1"
				a2 := "web-2"
				a3 := "web-3"
				return "ALARM(\"" + a1 + "\") OR ALARM(\"" + a2 + "\") OR ALARM(\"" + a3 + "\")"
			},
			expectedCalls: []string{"web-1"},
			description:   "Only first alarm should be called due to short-circuit evaluation",
		},
		{
			name: "Complex with NOT",
			buildRule: func() string {
				prod := "production-alarm"
				maint := "maintenance-mode"
				return "ALARM(\"" + prod + "\") AND NOT ALARM(\"" + maint + "\")"
			},
			expectedCalls: []string{"production-alarm", "maintenance-mode"},
			description:   "Both alarms should be evaluated for AND NOT expression",
		},
		{
			name: "Nested parentheses",
			buildRule: func() string {
				cpu1 := "cpu-server-1"
				cpu2 := "cpu-server-2"
				deploy := "deployment-in-progress"
				return "(ALARM(\"" + cpu1 + "\") OR ALARM(\"" + cpu2 + "\")) AND NOT ALARM(\"" + deploy + "\")"
			},
			expectedCalls: []string{"cpu-server-1", "deployment-in-progress"},
			description:   "Should evaluate first alarm in OR, then the NOT alarm",
		},
		{
			name: "Mixed OK and ALARM",
			buildRule: func() string {
				health := "health-check"
				errorRate := "error-rate"
				return "OK(\"" + health + "\") AND ALARM(\"" + errorRate + "\")"
			},
			expectedCalls: []string{"health-check", "error-rate"},
			description:   "Should handle both OK and ALARM functions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			alarmRule := tc.buildRule()
			t.Logf("Rule: %s", alarmRule)
			t.Logf("Description: %s", tc.description)

			// Act
			transformed := TransformAlarmRule(alarmRule)
			t.Logf("Transformed: %s", transformed)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "OK function should receive arguments")
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("OK(%q) called", name)
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM function should receive arguments")
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("ALARM(%q) called", name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse expression successfully")

			result, err := expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("Result: %v", result)
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected: %v", tc.expectedCalls)

			// Verify calls (accounting for short-circuit evaluation)
			for i, expected := range tc.expectedCalls {
				if assert.Less(t, i, len(callLog), "Should have call at index %d", i) {
					assert.Equal(t, expected, callLog[i], "Call %d should match expected alarm name", i)
				}
			}
		})
	}
}

// TestVariousNewlinePatterns - różne wzorce z newlines
func Test_VariousNewlinePatterns(t *testing.T) {
	testCases := []struct {
		name          string
		buildRule     func() string
		expectedCalls []string
		description   string
	}{
		{
			name: "Newline after opening paren",
			buildRule: func() string {
				a := "alarm1"
				return "ALARM(\n" + a + ")"
			},
			expectedCalls: []string{"alarm1"},
			description:   "Should trim newline after opening parenthesis",
		},
		{
			name: "Newline before closing paren",
			buildRule: func() string {
				a := "alarm2"
				return "ALARM(" + a + "\n)"
			},
			expectedCalls: []string{"alarm2"},
			description:   "Should trim newline before closing parenthesis",
		},
		{
			name: "Newlines on both sides",
			buildRule: func() string {
				a := "alarm3"
				return "ALARM(\n" + a + "\n)"
			},
			expectedCalls: []string{"alarm3"},
			description:   "Should trim newlines on both sides",
		},
		{
			name: "Newlines with spaces",
			buildRule: func() string {
				a := "alarm4"
				return "ALARM(\n " + a + " \n)"
			},
			expectedCalls: []string{"alarm4"},
			description:   "Should trim newlines and spaces",
		},
		{
			name: "Multiple newlines",
			buildRule: func() string {
				a := "alarm5"
				return "ALARM(\n\n" + a + "\n\n)"
			},
			expectedCalls: []string{"alarm5"},
			description:   "Should trim multiple consecutive newlines",
		},
		{
			name: "Tabs and newlines",
			buildRule: func() string {
				a := "alarm6"
				return "ALARM(\n\t" + a + "\t\n)"
			},
			expectedCalls: []string{"alarm6"},
			description:   "Should trim tabs and newlines",
		},
		{
			name: "Complex with newlines in AND",
			buildRule: func() string {
				a1 := "web-1"
				a2 := "web-2"
				return "ALARM(\n" + a1 + "\n) AND\nALARM(\n" + a2 + "\n)"
			},
			expectedCalls: []string{"web-1", "web-2"},
			description:   "Should handle newlines in complex expressions",
		},
		{
			name: "Quoted alarm names with newlines",
			buildRule: func() string {
				a1 := "prod-cpu"
				a2 := "prod-mem"
				return "ALARM(\n\"" + a1 + "\"\n) AND ALARM(\n\"" + a2 + "\"\n)"
			},
			expectedCalls: []string{"prod-cpu", "prod-mem"},
			description:   "Should handle quoted names with newlines",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			alarmRule := tc.buildRule()
			t.Logf("Input: %q", alarmRule)
			t.Logf("Description: %s", tc.description)

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
			require.NoError(t, err, "Should parse expression")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected:  %v", tc.expectedCalls)

			assert.Equal(t, len(tc.expectedCalls), len(callLog), "Should call expected number of functions")
			assert.Equal(t, tc.expectedCalls, callLog, "Should call functions with expected names")
		})
	}
}

// TestAWSStringConcatenationWithSpecialCharacters - alarmy ze specjalnymi znakami
func Test_AWSStringConcatenationWithSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name         string
		alarm1       string
		alarm2       string
		expectedName string
	}{
		{
			name:         "Hyphens in name",
			alarm1:       "my-prod-cpu-alarm",
			alarm2:       "my-prod-mem-alarm",
			expectedName: "my-prod-cpu-alarm",
		},
		{
			name:         "Underscores in name",
			alarm1:       "web_server_1_cpu",
			alarm2:       "web_server_2_cpu",
			expectedName: "web_server_1_cpu",
		},
		{
			name:         "Dots in name",
			alarm1:       "api.prod.errors",
			alarm2:       "api.prod.latency",
			expectedName: "api.prod.errors",
		},
		{
			name:         "Slashes in name (ARN-like)",
			alarm1:       "prod/web/cpu",
			alarm2:       "prod/web/memory",
			expectedName: "prod/web/cpu",
		},
		{
			name:         "Colons in name (ARN)",
			alarm1:       "arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm",
			alarm2:       "arn:aws:cloudwatch:us-east-1:123456:alarm:OtherAlarm",
			expectedName: "arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm",
		},
		{
			name:         "Numbers in name",
			alarm1:       "server-123-cpu",
			alarm2:       "server-456-cpu",
			expectedName: "server-123-cpu",
		},
		{
			name:         "Mixed special characters",
			alarm1:       "prod_web-server.cpu:high",
			alarm2:       "prod_web-server.mem:high",
			expectedName: "prod_web-server.cpu:high",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			alarmRule := "ALARM(\"" + tc.alarm1 + "\") AND ALARM(\"" + tc.alarm2 + "\")"
			t.Logf("AlarmRule: %s", alarmRule)

			// Act
			transformed := TransformAlarmRule(alarmRule)
			t.Logf("Transformed: %s", transformed)

			var firstAlarmName string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args, "ALARM should receive arguments")
					if firstAlarmName == "" {
						firstAlarmName = args[0].(string)
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse expression with special characters")

			_, err = expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.Equal(t, tc.expectedName, firstAlarmName, "First alarm name should match expected")
			t.Logf("✅ Correctly received: %q", firstAlarmName)
		})
	}
}

// TestComplexNestedExpressions - złożone zagnieżdżone wyrażenia
func TestComplexNestedExpressions(t *testing.T) {
	testCases := []struct {
		name        string
		alarmRule   string
		description string
		minCalls    int
		firstCall   string
	}{
		{
			name:        "Triple nested OR",
			alarmRule:   "((ALARM(\"a\") OR ALARM(\"b\")) OR ALARM(\"c\")) OR ALARM(\"d\")",
			description: "Should short-circuit after first true",
			minCalls:    1,
			firstCall:   "a",
		},
		{
			name:        "Mixed AND OR with NOT",
			alarmRule:   "(ALARM(\"x\") AND NOT ALARM(\"y\")) OR (ALARM(\"z\") AND ALARM(\"w\"))",
			description: "Should evaluate first branch completely",
			minCalls:    2,
			firstCall:   "x",
		},
		{
			name:        "Deep nesting with parentheses",
			alarmRule:   "((ALARM(\"p1\") AND ALARM(\"p2\")) AND (ALARM(\"p3\") OR ALARM(\"p4\")))",
			description: "Should evaluate nested expressions correctly",
			minCalls:    3,
			firstCall:   "p1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Logf("AlarmRule: %s", tc.alarmRule)
			t.Logf("Description: %s", tc.description)

			// Act
			transformed := TransformAlarmRule(tc.alarmRule)
			t.Logf("Transformed: %s", transformed)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					name := args[0].(string)
					callLog = append(callLog, name)
					t.Logf("ALARM(%q)", name)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err, "Should parse complex expression")

			result, err := expr.Evaluate(nil)
			require.NoError(t, err, "Should evaluate successfully")

			// Assert
			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, len(callLog), tc.minCalls, "Should make minimum expected calls")
			assert.Equal(t, tc.firstCall, callLog[0], "First call should match expected")

			t.Logf("Calls made: %v", callLog)
		})
	}
}
