package alarmrule

import (
	"testing"

	"github.com/Knetic/govaluate"
)

// TestTransformAlarmRule - wszystkie testy transformacji
func TestTransformAlarmRule(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Boolean literals
		{name: "TRUE", input: "TRUE", want: "true"},
		{name: "FALSE", input: "FALSE", want: "false"},
		{name: "(TRUE)", input: "(TRUE)", want: "(true)"},
		{name: "(FALSE)", input: "(FALSE)", want: "(false)"},

		// NOT operator with booleans
		{name: "NOT TRUE", input: "NOT TRUE", want: "!true"},
		{name: "NOT (TRUE)", input: "NOT (TRUE)", want: "!(true)"},
		{name: "(NOT TRUE)", input: "(NOT TRUE)", want: "(!true)"},
		{name: "NOT FALSE", input: "NOT FALSE", want: "!false"},
		{name: "NOT (FALSE)", input: "NOT (FALSE)", want: "!(false)"},
		{name: "(NOT FALSE)", input: "(NOT FALSE)", want: "(!false)"},
		{name: "NOT(TRUE)", input: "NOT(TRUE)", want: "!(true)"},
		{name: "NOT(FALSE)", input: "NOT(FALSE)", want: "!(false)"},

		// Normalization cases
		{name: "(TRUE)AND TRUE", input: "(TRUE)AND TRUE", want: "(true) && true"},
		{name: "TRUE AND(TRUE)", input: "TRUE AND(TRUE)", want: "true && (true)"},
		{name: "(TRUE)AND(FALSE)", input: "(TRUE)AND(FALSE)", want: "(true) && (false)"},

		// NOT with alarm functions
		{name: "NOT OK(a)", input: "NOT OK(a)", want: "!OK('a')"},
		{name: "NOT ALARM(a)", input: "NOT ALARM(a)", want: "!ALARM('a')"},
		{name: "NOT INSUFFICIENT_DATA(a)", input: "NOT INSUFFICIENT_DATA(a)", want: "!INSUFFICIENT_DATA('a')"},

		// Basic alarm states - NO QUOTES
		{name: "OK(a)", input: "OK(a)", want: "OK('a')"},
		{name: "ALARM(a)", input: "ALARM(a)", want: "ALARM('a')"},
		{name: "INSUFFICIENT_DATA(a)", input: "INSUFFICIENT_DATA(a)", want: "INSUFFICIENT_DATA('a')"},

		// Alarm states WITH quotes - AWS removes ONLY outer pair!
		{name: `OK("a")`, input: `OK("a")`, want: `OK('a')`},
		{name: "OK('a')", input: "OK('a')", want: "OK('a')"},
		{name: "OK(a')", input: "OK(a')", want: `OK('a\'')`},
		{name: `OK("'a'")`, input: `OK("'a'")`, want: `OK('\'a\'')`},
		{name: `OK("a)")`, input: `OK("a)")`, want: `OK('a)')`},
		{name: "OK('a)')", input: "OK('a)')", want: "OK('a)')"},
		{name: "OK(test'name)", input: "OK(test'name)", want: `OK('test\'name')`},

		// Whitespace handling
		{name: "OK( a )", input: "OK( a )", want: "OK('a')"},
		{name: `OK( "a" )`, input: `OK( "a" )`, want: `OK('a')`},
		{name: `OK( "a")`, input: `OK( "a")`, want: `OK('a')`},
		{name: `OK("a" )`, input: `OK("a" )`, want: `OK('a')`},
		{name: "OK( 'a')", input: "OK( 'a')", want: "OK('a')"},
		{name: `OK(" a ")`, input: `OK(" a ")`, want: `OK(' a ')`},

		// AND operator
		{name: "OK(a) AND ALARM(b)", input: "OK(a) AND ALARM(b)", want: "OK('a') && ALARM('b')"},
		{name: "NOT OK(a) AND ALARM(b)", input: "NOT OK(a) AND ALARM(b)", want: "!OK('a') && ALARM('b')"},
		{name: "(NOT OK(a)) AND ALARM(b)", input: "(NOT OK(a)) AND ALARM(b)", want: "(!OK('a')) && ALARM('b')"},
		{name: "OK(a)AND ALARM(b)", input: "OK(a)AND ALARM(b)", want: "OK('a') && ALARM('b')"},

		// OR operator
		{name: "OK(a) OR ALARM(b)", input: "OK(a) OR ALARM(b)", want: "OK('a') || ALARM('b')"},
		{name: "OK(a) OR(ALARM(b))", input: "OK(a) OR(ALARM(b))", want: "OK('a') || (ALARM('b'))"},

		// Complex expressions
		{name: "OK(a) AND ALARM(b) OR TRUE", input: "OK(a) AND ALARM(b) OR TRUE", want: "OK('a') && ALARM('b') || true"},
		{name: "NOT (OK(a) AND ALARM(b))", input: "NOT (OK(a) AND ALARM(b))", want: "!(OK('a') && ALARM('b'))"},
		{name: "OK(a) AND (ALARM(b) OR INSUFFICIENT_DATA(c))", input: "OK(a) AND (ALARM(b) OR INSUFFICIENT_DATA(c))", want: "OK('a') && (ALARM('b') || INSUFFICIENT_DATA('c'))"},
		{name: "NOT OK(a) AND NOT ALARM(b)", input: "NOT OK(a) AND NOT ALARM(b)", want: "!OK('a') && !ALARM('b')"},
		{name: "(OK(a) OR ALARM(b)) AND INSUFFICIENT_DATA(c)", input: "(OK(a) OR ALARM(b)) AND INSUFFICIENT_DATA(c)", want: "(OK('a') || ALARM('b')) && INSUFFICIENT_DATA('c')"},

		// Special characters in alarm names
		{name: "OK(my-alarm-123)", input: "OK(my-alarm-123)", want: "OK('my-alarm-123')"},
		{name: "ALARM(prod/web/cpu)", input: "ALARM(prod/web/cpu)", want: "ALARM('prod/web/cpu')"},

		// Real AWS examples
		{name: "AWS example 1", input: "ALARM(CPUUtilizationTooHigh) AND ALARM(DiskReadOpsTooHigh)", want: "ALARM('CPUUtilizationTooHigh') && ALARM('DiskReadOpsTooHigh')"},
		{name: "AWS example 2", input: "ALARM(CPUUtilizationTooHigh) AND NOT ALARM(DeploymentInProgress)", want: "ALARM('CPUUtilizationTooHigh') && !ALARM('DeploymentInProgress')"},
		{name: "Complex AWS example", input: "(ALARM(WebServer1CPU) OR ALARM(WebServer2CPU)) AND NOT ALARM(MaintenanceWindow)", want: "(ALARM('WebServer1CPU') || ALARM('WebServer2CPU')) && !ALARM('MaintenanceWindow')"},

		// Real AWS with quotes
		{name: "Real AWS - ALARM with double quotes", input: `ALARM("DobryAlarm")`, want: `ALARM('DobryAlarm')`},
		{name: "Real AWS - two alarms with AND", input: `ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`, want: `ALARM('DobryAlarm') && ALARM('dobryAlarm2')`},
		{name: "Real AWS - with newlines", input: "ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)", want: `ALARM('DobryAlarm') || ALARM('dobryAlarm2')`},
		{name: "Real AWS - with tabs", input: "ALARM(\t\"CPUHigh\"\t) AND ALARM(\t\"MemHigh\"\t)", want: `ALARM('CPUHigh') && ALARM('MemHigh')`},
		{name: "Real AWS - mixed whitespace", input: "ALARM( \n\t \"Test\" \n\t )", want: `ALARM('Test')`},
		{name: "Real AWS - newlines in complex expression", input: "ALARM(\n\"A\"\n) AND\nNOT ALARM(\n\"B\"\n)", want: `ALARM('A') && !ALARM('B')`},
		{name: "Real AWS - multiline formatted", input: "ALARM(\n\"Production-CPU\"\n) AND ALARM(\n\"Production-Memory\"\n)", want: `ALARM('Production-CPU') && ALARM('Production-Memory')`},
		{name: "Real AWS - ARN-like name", input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm")`, want: `ALARM('arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm')`},
		{name: "Real AWS - three alarms combined", input: `ALARM("Alarm1") AND ALARM("Alarm2") OR ALARM("Alarm3")`, want: `ALARM('Alarm1') && ALARM('Alarm2') || ALARM('Alarm3')`},
		{name: "Real AWS - with NOT", input: `ALARM("Production") AND NOT ALARM("Maintenance")`, want: `ALARM('Production') && !ALARM('Maintenance')`},
		{name: "Real AWS - complex with parentheses", input: `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`, want: `(ALARM('CPU1') || ALARM('CPU2')) && !ALARM('Deploying')`},
		{name: "Real AWS - OK and ALARM mixed", input: `OK("HealthCheck") AND ALARM("ErrorRate")`, want: `OK('HealthCheck') && ALARM('ErrorRate')`},
		{name: "Real AWS - INSUFFICIENT_DATA in mix", input: `ALARM("CPUHigh") AND NOT INSUFFICIENT_DATA("MetricMissing")`, want: `ALARM('CPUHigh') && !INSUFFICIENT_DATA('MetricMissing')`},

		// Edge cases
		{name: "Empty input", input: "", want: ""},
		{name: "Just whitespace", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformAlarmRule(tt.input)

			if got != tt.want {
				t.Errorf("TransformAlarmRule()\ninput: %q\ngot:   %q\nwant:  %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestCriticalCase - weryfikacja czy escape'owanie działa poprawnie
func TestCriticalCase(t *testing.T) {
	// AWS: OK("'test'") → alarm name = 'test'
	input := `OK("'test'")`

	t.Logf("AWS Input: %s", input)
	t.Logf("AWS alarm name będzie: 'test' (z apostrofami)")

	// Nasza transformacja
	transformed := TransformAlarmRule(input)
	t.Logf("Nasza transformacja: %s", transformed)

	// Co dostanie govaluate?
	var receivedArg string
	functions := map[string]govaluate.ExpressionFunction{
		"OK": func(args ...interface{}) (interface{}, error) {
			if len(args) > 0 {
				receivedArg = args[0].(string)
			}
			return true, nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
	if err != nil {
		t.Fatalf("Govaluate parse failed: %v", err)
	}

	_, err = expr.Evaluate(nil)
	if err != nil {
		t.Fatalf("Govaluate evaluate failed: %v", err)
	}

	t.Logf("Govaluate otrzymał: %q", receivedArg)

	expectedFromAWS := "'test'" // To jest nazwa alarmu w AWS!

	if receivedArg != expectedFromAWS {
		t.Errorf("❌ NIEZGODNE Z AWS!\nGovaluate dostał: %q\nAWS ma:           %q", receivedArg, expectedFromAWS)
	} else {
		t.Logf("✅ ZGODNE Z AWS! Govaluate dostał dokładnie to co AWS: %q", receivedArg)
	}
}

// TestAllEdgeCasesWithAWS - porównanie z AWS dla wszystkich edge cases
func TestAllEdgeCasesWithAWS(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Nasza transformacja
			transformed := TransformAlarmRule(tt.awsInput)

			// Co dostanie govaluate?
			var receivedArg string
			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						receivedArg = args[0].(string)
					}
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						receivedArg = args[0].(string)
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			if err != nil {
				t.Fatalf("Govaluate parse failed: %v\nTransformed: %s", err, transformed)
			}

			_, err = expr.Evaluate(nil)
			if err != nil {
				t.Fatalf("Govaluate evaluate failed: %v\nTransformed: %s", err, transformed)
			}

			t.Logf("AWS Input:          %s", tt.awsInput)
			t.Logf("AWS alarm name:     %q", tt.awsAlarmName)
			t.Logf("Transformed:        %s", transformed)
			t.Logf("Govaluate received: %q", receivedArg)

			if receivedArg != tt.awsAlarmName {
				t.Errorf("❌ MISMATCH!\n  AWS alarm name:     %q\n  Govaluate received: %q",
					tt.awsAlarmName, receivedArg)
			} else {
				t.Logf("✅ MATCH! Govaluate otrzymał dokładnie to co AWS ma w nazwie alarmu")
			}
		})
	}
}

// TestGovaluateEscaping - test czy govaluate poprawnie interpretuje nasze escape'y
func TestGovaluateEscaping(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var received string

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						received = args[0].(string)
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tt.expression, functions)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			_, err = expr.Evaluate(nil)
			if err != nil {
				t.Fatalf("Failed to evaluate: %v", err)
			}

			if received != tt.expected {
				t.Errorf("Mismatch!\n  Expression: %s\n  Got:  %q\n  Want: %q",
					tt.expression, received, tt.expected)
			} else {
				t.Logf("✅ Expression: %s → %q", tt.expression, received)
			}
		})
	}
}

// TestAllTransformedWithGovaluate testuje czy WSZYSTKIE nasze transformacje działają z govaluate
func TestAllTransformedWithGovaluate(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformAlarmRule(tt.input)

			if got != tt.want {
				t.Fatalf("Transform mismatch:\ngot:  %q\nwant: %q", got, tt.want)
			}

			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						callLog = append(callLog, "OK("+args[0].(string)+")")
					}
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						callLog = append(callLog, "ALARM("+args[0].(string)+")")
					}
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						callLog = append(callLog, "INSUFFICIENT_DATA("+args[0].(string)+")")
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tt.want, functions)
			if err != nil {
				t.Fatalf("Govaluate failed to parse: %v\nExpression: %s", err, tt.want)
			}

			result, err := expr.Evaluate(nil)
			if err != nil {
				t.Fatalf("Govaluate failed to evaluate: %v\nExpression: %s", err, tt.want)
			}

			t.Logf("✅ Input: %s", tt.input)
			t.Logf("   Transformed: %s", tt.want)
			t.Logf("   Govaluate result: %v", result)
			t.Logf("   Function calls: %v", callLog)
		})
	}
}

// TestGovaluateReceivesCorrectArguments sprawdza dokładnie co govaluate otrzymuje jako argumenty
func TestGovaluateReceivesCorrectArguments(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedArg interface{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						receivedArg = args[0]
					}
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						receivedArg = args[0]
					}
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						receivedArg = args[0]
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tt.transformed, functions)
			if err != nil {
				t.Fatalf("Govaluate parse failed: %v", err)
			}

			_, err = expr.Evaluate(nil)
			if err != nil {
				t.Fatalf("Govaluate evaluate failed: %v", err)
			}

			receivedStr, ok := receivedArg.(string)
			if !ok {
				t.Fatalf("Received arg is not string: %T %v", receivedArg, receivedArg)
			}

			if receivedStr != tt.expectedArg {
				t.Errorf("%s received wrong argument:\ngot:  %q\nwant: %q",
					tt.expectedFunction, receivedStr, tt.expectedArg)
			} else {
				t.Logf("✅ %s received correct arg: %q", tt.expectedFunction, receivedStr)
			}
		})
	}
}

// TestFullPipeline testuje pełny pipeline: AWS → Transform → Govaluate
func TestFullPipeline(t *testing.T) {
	tests := []struct {
		name          string
		awsInput      string
		expectedCalls []string
	}{
		{
			name:          "Simple alarm",
			awsInput:      `ALARM("DobryAlarm")`,
			expectedCalls: []string{`DobryAlarm`},
		},
		{
			name:          "Alarm with parenthesis in name",
			awsInput:      `OK("a)")`,
			expectedCalls: []string{`a)`},
		},
		{
			name:          "Two alarms with AND",
			awsInput:      `ALARM("Alarm1") AND ALARM("Alarm2")`,
			expectedCalls: []string{`Alarm1`, `Alarm2`},
		},
		{
			name:          "Complex expression",
			awsInput:      `(OK("Health") OR ALARM("Error")) AND NOT ALARM("Maintenance")`,
			expectedCalls: []string{`Health`, `Maintenance`},
		},
		{
			name:          "Alarm with inner quotes - AWS keeps them",
			awsInput:      `OK("'test'")`,
			expectedCalls: []string{`'test'`},
		},
		{
			name:          "Alarm with apostrophe in middle",
			awsInput:      `OK(test'name)`,
			expectedCalls: []string{`test'name`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := TransformAlarmRule(tt.awsInput)

			t.Logf("AWS Input:    %s", tt.awsInput)
			t.Logf("Transformed:  %s", transformed)

			callLog := []string{}

			functions := map[string]govaluate.ExpressionFunction{
				"OK": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						alarmName := args[0].(string)
						callLog = append(callLog, alarmName)
						t.Logf("OK() called with: %q", alarmName)
					}
					return true, nil
				},
				"ALARM": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						alarmName := args[0].(string)
						callLog = append(callLog, alarmName)
						t.Logf("ALARM() called with: %q", alarmName)
					}
					return true, nil
				},
				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
					if len(args) > 0 {
						alarmName := args[0].(string)
						callLog = append(callLog, alarmName)
						t.Logf("INSUFFICIENT_DATA() called with: %q", alarmName)
					}
					return true, nil
				},
			}

			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
			if err != nil {
				t.Fatalf("Failed to create govaluate expression: %v", err)
			}

			result, err := expr.Evaluate(nil)
			if err != nil {
				t.Fatalf("Failed to evaluate: %v", err)
			}

			t.Logf("Result: %v", result)
			t.Logf("Call log: %v", callLog)
			t.Logf("Expected: %v", tt.expectedCalls)
			// Verify we got expected calls
			for i, expected := range tt.expectedCalls {
				if i < len(callLog) {
					if callLog[i] != expected {
						t.Errorf("Call %d: got %q, want %q", i, callLog[i], expected)
					}
				}
			}
		})
	}
}

// // Benchmarks
// func BenchmarkTransformAlarmRule(b *testing.B) {
// 	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }
// func BenchmarkTransformAlarmRuleSimple(b *testing.B) {
// 	input := ALARM("SimpleAlarm")
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }
// func BenchmarkTransformAlarmRuleComplex(b *testing.B) {
// 	input := `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = TransformAlarmRule(input)
// 	}
// }
