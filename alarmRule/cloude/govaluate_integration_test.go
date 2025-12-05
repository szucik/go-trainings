package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGovaluateEscaping verifies that govaluate correctly interprets our escape sequences.
func TestGovaluateEscaping(t *testing.T) {
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
			var received string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					received = args[0].(string)
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.expression, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, received)
		})
	}
}

// TestAllTransformedWithGovaluate verifies that all transformations work correctly with govaluate evaluation.
func TestAllTransformedWithGovaluate(t *testing.T) {
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

		{`ALARM("testAlarm1")`, `ALARM("testAlarm1")`, `ALARM('testAlarm1')`},
		{`ALARM("testAlarm1") AND ALARM("testAlarm2")`, `ALARM("testAlarm1") AND ALARM("testAlarm2")`, `ALARM('testAlarm1') && ALARM('testAlarm2')`},
		{"ALARM(\n\"testAlarm1\"\n) OR ALARM(\n\"testAlarm2\"\n)", "ALARM(\n\"testAlarm1\"\n) OR ALARM(\n\"testAlarm2\"\n)", `ALARM('testAlarm1') || ALARM('testAlarm2')`},
		{`(ALARM("cpu1") OR ALARM("cpu2")) AND NOT ALARM("deploying")`, `(ALARM("cpu1") OR ALARM("cpu2")) AND NOT ALARM("deploying")`, `(ALARM('cpu1') || ALARM('cpu2')) && !ALARM('deploying')`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformAlarmRule(tc.input)
			assert.Equal(t, tc.want, got)

			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, "OK("+args[0].(string)+")")
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, "ALARM("+args[0].(string)+")")
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, "INSUFFICIENT_DATA("+args[0].(string)+")")
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.want, functions)
			require.NoError(t, err)

			result, err := expr.Evaluate(nil)
			require.NoError(t, err)

			assert.NotNil(t, result)
		})
	}
}

// TestGovaluateReceivesCorrectArguments verifies that govaluate passes the correct alarm names to functions.
func TestGovaluateReceivesCorrectArguments(t *testing.T) {
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
			name:             "Name without quotes",
			transformed:      `ALARM('testAlarm')`,
			expectedFunction: "ALARM",
			expectedArg:      `testAlarm`,
		},
		{
			name:             "Name with parenthesis",
			transformed:      `OK('a)')`,
			expectedFunction: "OK",
			expectedArg:      `a)`,
		},
		{
			name:             "Escaped apostrophe",
			transformed:      `OK('a\'')`,
			expectedFunction: "OK",
			expectedArg:      `a'`,
		},
		{
			name:             "Inner quotes preserved",
			transformed:      `OK('\'test\'')`,
			expectedFunction: "OK",
			expectedArg:      `'test'`,
		},
		{
			name:             "Apostrophe in middle",
			transformed:      `OK('test\'name')`,
			expectedFunction: "OK",
			expectedArg:      `test'name`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var receivedArg interface{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedArg = args[0]
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedArg = args[0]
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					receivedArg = args[0]
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tc.transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			receivedStr, ok := receivedArg.(string)
			require.True(t, ok, "Expected string, got %T", receivedArg)

			assert.Equal(t, tc.expectedArg, receivedStr)
		})
	}
}

// TestAWSStringConcatenation tests AWS SDK string concatenation pattern: "ALARM(\"" + alarm1 + "\")".
func TestAWSStringConcatenation(t *testing.T) {
	alarm1 := "cpu-high-alarm"
	alarm2 := "memory-high-alarm"
	alarmRule := "ALARM(\"" + alarm1 + "\") AND ALARM(\"" + alarm2 + "\")"

	transformed := TransformAlarmRule(alarmRule)

	expectedTransformed := "ALARM('cpu-high-alarm') && ALARM('memory-high-alarm')"
	assert.Equal(t, expectedTransformed, transformed)

	var receivedAlarms []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args)
			alarmName := args[0].(string)
			receivedAlarms = append(receivedAlarms, alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err)

	result, err := expr.Evaluate(nil)
	require.NoError(t, err)

	assert.True(t, result.(bool))
	assert.Len(t, receivedAlarms, 2)
	assert.Equal(t, []string{"cpu-high-alarm", "memory-high-alarm"}, receivedAlarms)
}

// TestAWSStringConcatenationWithNewlines tests AlarmRule with newline characters (\n).
func TestAWSStringConcatenationWithNewlines(t *testing.T) {
	alarm1 := "cpu-high"
	alarm2 := "memory-high"
	alarmRule := "ALARM(\n" + alarm1 + " \n) AND ALARM(\n" + alarm2 + "\n)"

	transformed := TransformAlarmRule(alarmRule)

	expectedTransformed := "ALARM('cpu-high') && ALARM('memory-high')"
	assert.Equal(t, expectedTransformed, transformed)

	var receivedAlarms []string

	functions := map[string]govaluate.ExpressionFunction{
		"ALARM": func(args ...interface{}) (interface{}, error) {
			require.NotEmpty(t, args)
			alarmName := args[0].(string)
			receivedAlarms = append(receivedAlarms, alarmName)
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	require.NoError(t, err)

	result, err := expr.Evaluate(nil)
	require.NoError(t, err)

	assert.True(t, result.(bool))
	assert.Equal(t, []string{"cpu-high", "memory-high"}, receivedAlarms)
}

// TestMultipleAWSStringConcatenationPatterns tests various AWS AlarmRule construction patterns.
func TestMultipleAWSStringConcatenationPatterns(t *testing.T) {
	testCases := []struct {
		name          string
		buildRule     func() string
		expectedCalls []string
	}{
		{
			name: "Two alarms with AND",
			buildRule: func() string {
				return "ALARM(\"alarm-1\") AND ALARM(\"alarm-2\")"
			},
			expectedCalls: []string{"alarm-1", "alarm-2"},
		},
		{
			name: "OR with short-circuit",
			buildRule: func() string {
				return "ALARM(\"web-1\") OR ALARM(\"web-2\") OR ALARM(\"web-3\")"
			},
			expectedCalls: []string{"web-1"},
		},
		{
			name: "AND with NOT",
			buildRule: func() string {
				return "ALARM(\"prod-alarm\") AND NOT ALARM(\"maintenance\")"
			},
			expectedCalls: []string{"prod-alarm", "maintenance"},
		},
		{
			name: "Nested with NOT",
			buildRule: func() string {
				return "(ALARM(\"cpu-1\") OR ALARM(\"cpu-2\")) AND NOT ALARM(\"deploying\")"
			},
			expectedCalls: []string{"cpu-1", "deploying"},
		},
		{
			name: "Mixed functions",
			buildRule: func() string {
				return "OK(\"health-check\") AND ALARM(\"error-rate\")"
			},
			expectedCalls: []string{"health-check", "error-rate"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alarmRule := tc.buildRule()
			transformed := TransformAlarmRule(alarmRule)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, args[0].(string))
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, args[0].(string))
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

// TestVariousNewlinePatterns tests different newline formatting patterns in AlarmRules.
func TestVariousNewlinePatterns(t *testing.T) {
	testCases := []struct {
		name          string
		buildRule     func() string
		expectedCalls []string
	}{
		{
			name: "Newline after opening paren",
			buildRule: func() string {
				return "ALARM(\nalarm1)"
			},
			expectedCalls: []string{"alarm1"},
		},
		{
			name: "Newline before closing paren",
			buildRule: func() string {
				return "ALARM(alarm2\n)"
			},
			expectedCalls: []string{"alarm2"},
		},
		{
			name: "Newlines on both sides",
			buildRule: func() string {
				return "ALARM(\nalarm3\n)"
			},
			expectedCalls: []string{"alarm3"},
		},
		{
			name: "Newlines with spaces",
			buildRule: func() string {
				return "ALARM(\n alarm4 \n)"
			},
			expectedCalls: []string{"alarm4"},
		},
		{
			name: "Multiple newlines",
			buildRule: func() string {
				return "ALARM(\n\nalarm5\n\n)"
			},
			expectedCalls: []string{"alarm5"},
		},
		{
			name: "Tabs and newlines",
			buildRule: func() string {
				return "ALARM(\n\talarm6\t\n)"
			},
			expectedCalls: []string{"alarm6"},
		},
		{
			name: "Newlines in AND expression",
			buildRule: func() string {
				return "ALARM(\nweb-1\n) AND\nALARM(\nweb-2\n)"
			},
			expectedCalls: []string{"web-1", "web-2"},
		},
		{
			name: "Quoted with newlines",
			buildRule: func() string {
				return "ALARM(\n\"prod-cpu\"\n) AND ALARM(\n\"prod-mem\"\n)"
			},
			expectedCalls: []string{"prod-cpu", "prod-mem"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alarmRule := tc.buildRule()
			transformed := TransformAlarmRule(alarmRule)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, args[0].(string))
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedCalls, callLog)
		})
	}
}

// TestAWSStringConcatenationWithSpecialCharacters tests alarm names containing special characters.
func TestAWSStringConcatenationWithSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name         string
		alarm1       string
		alarm2       string
		expectedName string
	}{
		{
			name:         "Hyphens",
			alarm1:       "my-prod-cpu-alarm",
			alarm2:       "my-prod-mem-alarm",
			expectedName: "my-prod-cpu-alarm",
		},
		{
			name:         "Underscores",
			alarm1:       "web_server_1_cpu",
			alarm2:       "web_server_2_cpu",
			expectedName: "web_server_1_cpu",
		},
		{
			name:         "Dots",
			alarm1:       "api.prod.errors",
			alarm2:       "api.prod.latency",
			expectedName: "api.prod.errors",
		},
		{
			name:         "Slashes",
			alarm1:       "prod/web/cpu",
			alarm2:       "prod/web/memory",
			expectedName: "prod/web/cpu",
		},
		{
			name:         "ARN format",
			alarm1:       "arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm",
			alarm2:       "arn:aws:cloudwatch:us-east-1:123456:alarm:OtherAlarm",
			expectedName: "arn:aws:cloudwatch:us-east-1:123456:alarm:MyAlarm",
		},
		{
			name:         "Numbers",
			alarm1:       "server-123-cpu",
			alarm2:       "server-456-cpu",
			expectedName: "server-123-cpu",
		},
		{
			name:         "Mixed special chars",
			alarm1:       "prod_web-server.cpu:high",
			alarm2:       "prod_web-server.mem:high",
			expectedName: "prod_web-server.cpu:high",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alarmRule := "ALARM(\"" + tc.alarm1 + "\") AND ALARM(\"" + tc.alarm2 + "\")"
			transformed := TransformAlarmRule(alarmRule)

			var firstAlarmName string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					if firstAlarmName == "" {
						firstAlarmName = args[0].(string)
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			_, err = expr.Evaluate(nil)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedName, firstAlarmName)
		})
	}
}

func TestComplexNestedExpressions(t *testing.T) {
	testCases := []struct {
		name      string
		alarmRule string
		minCalls  int
		firstCall string
	}{
		{
			name:      "Triple nested OR",
			alarmRule: "((ALARM(\"a\") OR ALARM(\"b\")) OR ALARM(\"c\")) OR ALARM(\"d\")",
			minCalls:  1,
			firstCall: "a",
		},
		{
			name:      "Mixed AND OR with NOT",
			alarmRule: "(ALARM(\"x\") AND NOT ALARM(\"y\")) OR (ALARM(\"z\") AND ALARM(\"w\"))",
			minCalls:  2,
			firstCall: "x",
		},
		{
			name:      "Deep nesting",
			alarmRule: "((ALARM(\"p1\") AND ALARM(\"p2\")) AND (ALARM(\"p3\") OR ALARM(\"p4\")))",
			minCalls:  3,
			firstCall: "p1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tc.alarmRule)

			var callLog []string

			functions := map[string]govaluate.ExpressionFunction{
				"ALARM": func(args ...interface{}) (interface{}, error) {
					require.NotEmpty(t, args)
					callLog = append(callLog, args[0].(string))
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			require.NoError(t, err)

			result, err := expr.Evaluate(nil)
			require.NoError(t, err)

			assert.NotNil(t, result)
			assert.GreaterOrEqual(t, len(callLog), tc.minCalls)
			assert.Equal(t, tc.firstCall, callLog[0])
		})
	}
}
