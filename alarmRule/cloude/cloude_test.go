package alarmrule

import (
	"testing"
)

func TestTransformAlarmRule(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// Boolean literals
		{
			name:  "TRUE",
			input: "TRUE",
			want:  "true",
		},
		{
			name:  "FALSE",
			input: "FALSE",
			want:  "false",
		},
		{
			name:  "(TRUE)",
			input: "(TRUE)",
			want:  "(true)",
		},
		{
			name:  "(FALSE)",
			input: "(FALSE)",
			want:  "(false)",
		},
		// Dodaj do TestTransformAlarmRule:

		{
			name:  "(TRUE)AND TRUE",
			input: "(TRUE)AND TRUE",
			want:  "(true) && true",
		},
		{
			name:  "TRUE AND(TRUE)",
			input: "TRUE AND(TRUE)",
			want:  "true && (true)",
		},
		{
			name:  "(TRUE)AND(FALSE)",
			input: "(TRUE)AND(FALSE)",
			want:  "(true) && (false)",
		},
		{
			name:  "NOT(TRUE)",
			input: "NOT(TRUE)",
			want:  "!(true)",
		},
		{
			name:  "OK(a)AND ALARM(b)",
			input: "OK(a)AND ALARM(b)",
			want:  "OK('a') && ALARM('b')",
		},
		{
			name:  "OK(a) OR(ALARM(b))",
			input: "OK(a) OR(ALARM(b))",
			want:  "OK('a') || (ALARM('b'))",
		},
		// NOT operator
		{
			name:  "NOT TRUE",
			input: "NOT TRUE",
			want:  "!true",
		},
		{
			name:  "NOT (TRUE)",
			input: "NOT (TRUE)",
			want:  "!(true)",
		},
		{
			name:  "(NOT TRUE)",
			input: "(NOT TRUE)",
			want:  "(!true)",
		},
		{
			name:  "NOT FALSE",
			input: "NOT FALSE",
			want:  "!false",
		},
		{
			name:  "NOT (FALSE)",
			input: "NOT (FALSE)",
			want:  "!(false)",
		},
		{
			name:  "NOT(FALSE)",
			input: "NOT(FALSE)",
			want:  "!(false)",
		},
		{
			name:  "NOT(TRUE)",
			input: "NOT(TRUE)",
			want:  "!(true)",
		},
		{
			name:  "(NOT FALSE)",
			input: "(NOT FALSE)",
			want:  "(!false)",
		},
		{
			name:  "NOT OK(a)",
			input: "NOT OK(a)",
			want:  "!OK('a')",
		},
		{
			name:  "NOT ALARM(a)",
			input: "NOT ALARM(a)",
			want:  "!ALARM('a')",
		},
		{
			name:  "NOT INSUFFICIENT_DATA(a)",
			input: "NOT INSUFFICIENT_DATA(a)",
			want:  "!INSUFFICIENT_DATA('a')",
		},

		// Alarm states - basic
		{
			name:  "OK(a)",
			input: "OK(a)",
			want:  "OK('a')",
		},
		{
			name:  "ALARM(a)",
			input: "ALARM(a)",
			want:  "ALARM('a')",
		},
		{
			name:  "INSUFFICIENT_DATA(a)",
			input: "INSUFFICIENT_DATA(a)",
			want:  "INSUFFICIENT_DATA('a')",
		},

		// Cudzysłowy i apostrofy
		{
			name:  `OK("a")`,
			input: `OK("a")`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  "OK('a')",
			input: "OK('a')",
			want:  "OK('a')",
		},
		{
			name:  "OK('a)",
			input: "OK('a)",
			want:  `OK('\'a')`,
		},
		{
			name:  "OK(a')",
			input: "OK(a')",
			want:  `OK('a\'')`,
		},
		{
			name:  `OK("'a'")`,
			input: `OK("'a'")`,
			want:  `OK('\"\'a\'\"')`,
		},
		{
			name:  `OK("a)")`,
			input: `OK("a)")`,
			want:  `OK('\"a)\"')`,
		},
		{
			name:  "OK('a)')",
			input: "OK('a)')",
			want:  "OK('a)')",
		},

		// Whitespace handling
		{
			name:  "OK( a )",
			input: "OK( a )",
			want:  "OK('a')",
		},
		{
			name:  `OK( "a" )`,
			input: `OK( "a" )`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  `OK( "a")`,
			input: `OK( "a")`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  `OK("a" )`,
			input: `OK("a" )`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  "OK( 'a')",
			input: "OK( 'a')",
			want:  "OK('a')",
		},
		{
			name:  `OK(" a ")`,
			input: `OK(" a ")`,
			want:  `OK('\" a \"')`,
		},

		// AND operator
		{
			name:  "OK(a) AND ALARM(b)",
			input: "OK(a) AND ALARM(b)",
			want:  "OK('a') && ALARM('b')",
		},
		{
			name:  "NOT OK(a) AND ALARM(b)",
			input: "NOT OK(a) AND ALARM(b)",
			want:  "!OK('a') && ALARM('b')",
		},
		{
			name:  "(NOT OK(a)) AND ALARM(b)",
			input: "(NOT OK(a)) AND ALARM(b)",
			want:  "(!OK('a')) && ALARM('b')",
		},

		// OR operator
		{
			name:  "OK(a) OR ALARM(b)",
			input: "OK(a) OR ALARM(b)",
			want:  "OK('a') || ALARM('b')",
		},
		{
			name:  "OK(a) AND ALARM(b) OR TRUE",
			input: "OK(a) AND ALARM(b) OR TRUE",
			want:  "OK('a') && ALARM('b') || true",
		},

		// Złożone wyrażenia
		{
			name:  "NOT (OK(a) AND ALARM(b))",
			input: "NOT (OK(a) AND ALARM(b))",
			want:  "!(OK('a') && ALARM('b'))",
		},
		{
			name:  "OK(a) AND (ALARM(b) OR INSUFFICIENT_DATA(c))",
			input: "OK(a) AND (ALARM(b) OR INSUFFICIENT_DATA(c))",
			want:  "OK('a') && (ALARM('b') || INSUFFICIENT_DATA('c'))",
		},
		{
			name:  "NOT OK(a) AND NOT ALARM(b)",
			input: "NOT OK(a) AND NOT ALARM(b)",
			want:  "!OK('a') && !ALARM('b')",
		},
		{
			name:  "(OK(a) OR ALARM(b)) AND INSUFFICIENT_DATA(c)",
			input: "(OK(a) OR ALARM(b)) AND INSUFFICIENT_DATA(c)",
			want:  "(OK('a') || ALARM('b')) && INSUFFICIENT_DATA('c')",
		},

		// Edge cases z nazwami alarmów
		{
			name:  "OK(my-alarm-123)",
			input: "OK(my-alarm-123)",
			want:  "OK('my-alarm-123')",
		},
		{
			name:  "ALARM(prod/web/cpu)",
			input: "ALARM(prod/web/cpu)",
			want:  "ALARM('prod/web/cpu')",
		},
		{
			name:  `OK('"a"')`,
			input: `OK('"a"')`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  `OK("'a"')`,
			input: `OK("'a"')`,
			want:  `OK('\"\'a\"\'')`,
		},
		{
			name:  `OK("'a)`,
			input: `OK("'a)`,
			want:  `OK('\"\'a')`,
		},
		{
			name:  `OK('"a)`,
			input: `OK('"a)`,
			want:  `OK('\'\"a')`,
		},

		// Przypadki AWS z dokumentacji
		{
			name:  "AWS example 1",
			input: "ALARM(CPUUtilizationTooHigh) AND ALARM(DiskReadOpsTooHigh)",
			want:  "ALARM('CPUUtilizationTooHigh') && ALARM('DiskReadOpsTooHigh')",
		},
		{
			name:  "AWS example 2",
			input: "ALARM(CPUUtilizationTooHigh) AND NOT ALARM(DeploymentInProgress)",
			want:  "ALARM('CPUUtilizationTooHigh') && !ALARM('DeploymentInProgress')",
		},
		{
			name:  "Complex AWS example",
			input: "(ALARM(WebServer1CPU) OR ALARM(WebServer2CPU)) AND NOT ALARM(MaintenanceWindow)",
			want:  "(ALARM('WebServer1CPU') || ALARM('WebServer2CPU')) && !ALARM('MaintenanceWindow')",
		},

		// Real AWS scenarios
		{
			name:  "Real AWS - ALARM with double quotes and trailing space",
			input: `ALARM("DobryAlarm") `,
			want:  `ALARM('\"DobryAlarm\"')`,
		},
		{
			name:  "Real AWS - two alarms with AND",
			input: `ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`,
			want:  `ALARM('\"DobryAlarm\"') && ALARM('\"dobryAlarm2\"')`,
		},
		{
			name:  "Real AWS - with newlines",
			input: "ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)",
			want:  `ALARM('\"DobryAlarm\"') || ALARM('\"dobryAlarm2\"')`,
		},
		{
			name:  "Real AWS - with tabs",
			input: "ALARM(\t\"CPUHigh\"\t) AND ALARM(\t\"MemHigh\"\t)",
			want:  `ALARM('\"CPUHigh\"') && ALARM('\"MemHigh\"')`,
		},
		{
			name:  "Real AWS - mixed whitespace",
			input: "ALARM( \n\t \"Test\" \n\t )",
			want:  `ALARM('\"Test\"')`,
		},
		{
			name:  "Real AWS - newlines in complex expression",
			input: "ALARM(\n\"A\"\n) AND\nNOT ALARM(\n\"B\"\n)",
			want:  `ALARM('\"A\"') && !ALARM('\"B\"')`,
		},
		{
			name: "Real AWS - multiline formatted",
			input: `ALARM(
"Production-CPU"
) AND ALARM(
"Production-Memory"
)`,
			want: `ALARM('\"Production-CPU\"') && ALARM('\"Production-Memory\"')`,
		},
		{
			name:  "Real AWS - variable interpolation pattern",
			input: `ALARM("my-prod-alarm")`,
			want:  `ALARM('\"my-prod-alarm\"')`,
		},
		{
			name:  "Real AWS - ARN-like name",
			input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm")`,
			want:  `ALARM('\"arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm\"')`,
		},
		{
			name:  "Real AWS - two alarms with OR",
			input: `ALARM("HighCPU") OR ALARM("HighMemory")`,
			want:  `ALARM('\"HighCPU\"') || ALARM('\"HighMemory\"')`,
		},
		{
			name:  "Real AWS - three alarms combined",
			input: `ALARM("Alarm1") AND ALARM("Alarm2") OR ALARM("Alarm3")`,
			want:  `ALARM('\"Alarm1\"') && ALARM('\"Alarm2\"') || ALARM('\"Alarm3\"')`,
		},
		{
			name:  "Real AWS - with NOT",
			input: `ALARM("Production") AND NOT ALARM("Maintenance")`,
			want:  `ALARM('\"Production\"') && !ALARM('\"Maintenance\"')`,
		},
		{
			name:  "Real AWS - complex with parentheses",
			input: `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`,
			want:  `(ALARM('\"CPU1\"') || ALARM('\"CPU2\"')) && !ALARM('\"Deploying\"')`,
		},
		{
			name:  "Real AWS - OK and ALARM mixed",
			input: `OK("HealthCheck") AND ALARM("ErrorRate")`,
			want:  `OK('\"HealthCheck\"') && ALARM('\"ErrorRate\"')`,
		},
		{
			name:  "Real AWS - INSUFFICIENT_DATA in mix",
			input: `ALARM("CPUHigh") AND NOT INSUFFICIENT_DATA("MetricMissing")`,
			want:  `ALARM('\"CPUHigh\"') && !INSUFFICIENT_DATA('\"MetricMissing\"')`,
		},

		// Error cases
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformAlarmRule(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("TransformAlarmRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("TransformAlarmRule()\ninput: %q\ngot:   %q\nwant:  %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestAllTransformedWithGovaluate testuje czy WSZYSTKIE nasze transformacje działają z govaluate
// func TestAllTransformedWithGovaluate(t *testing.T) {
// 	// Wszystkie test case'y które zawierają wywołania funkcji
// 	tests := []struct {
// 		name  string
// 		input string
// 		want  string
// 	}{
// 		// Alarm states
// 		{"OK(a)", "OK(a)", "OK('a')"},
// 		{"ALARM(a)", "ALARM(a)", "ALARM('a')"},
// 		{"INSUFFICIENT_DATA(a)", "INSUFFICIENT_DATA(a)", "INSUFFICIENT_DATA('a')"},
// 		{"NOT OK(a)", "NOT OK(a)", "!OK('a')"},
// 		{"NOT ALARM(a)", "NOT ALARM(a)", "!ALARM('a')"},
// 		{"NOT INSUFFICIENT_DATA(a)", "NOT INSUFFICIENT_DATA(a)", "!INSUFFICIENT_DATA('a')"},

// 		// Cudzysłowy i apostrofy
// 		{`OK("a")`, `OK("a")`, `OK('\"a\"')`},
// 		{"OK('a')", "OK('a')", "OK('a')"},
// 		{"OK('a)", "OK('a)", `OK('\'a')`},
// 		{"OK(a')", "OK(a')", `OK('a\'')`},
// 		{`OK("'a'")`, `OK("'a'")`, `OK('\"\'a\'\"')`},
// 		{`OK("a)")`, `OK("a)")`, `OK('\"a)\"')`},
// 		{"OK('a)')", "OK('a)')", "OK('a)')"},

// 		// Whitespace
// 		{"OK( a )", "OK( a )", "OK('a')"},
// 		{`OK( "a" )`, `OK( "a" )`, `OK('\"a\"')`},
// 		{`OK(" a ")`, `OK(" a ")`, `OK('\" a \"')`},

// 		// AND/OR
// 		{"OK(a) AND ALARM(b)", "OK(a) AND ALARM(b)", "OK('a') && ALARM('b')"},
// 		{"NOT OK(a) AND ALARM(b)", "NOT OK(a) AND ALARM(b)", "!OK('a') && ALARM('b')"},
// 		{"OK(a) OR ALARM(b)", "OK(a) OR ALARM(b)", "OK('a') || ALARM('b')"},
// 		{"NOT (OK(a) AND ALARM(b))", "NOT (OK(a) AND ALARM(b))", "!(OK('a') && ALARM('b'))"},

// 		// Real AWS
// 		{`ALARM("DobryAlarm") `, `ALARM("DobryAlarm") `, `ALARM('\"DobryAlarm\"')`},
// 		{`ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`, `ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`, `ALARM('\"DobryAlarm\"') && ALARM('\"dobryAlarm2\"')`},
// 		{"ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)", "ALARM(\n\"DobryAlarm\"\n) OR ALARM(\n\"dobryAlarm2\"\n)", `ALARM('\"DobryAlarm\"') || ALARM('\"dobryAlarm2\"')`},
// 		{`(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`, `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`, `(ALARM('\"CPU1\"') || ALARM('\"CPU2\"')) && !ALARM('\"Deploying\"')`},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Step 1: Verify our transformation is correct
// 			got, err := TransformAlarmRule(tt.input)
// 			if err != nil {
// 				t.Fatalf("TransformAlarmRule failed: %v", err)
// 			}

// 			if got != tt.want {
// 				t.Fatalf("Transform mismatch:\ngot:  %q\nwant: %q", got, tt.want)
// 			}

// 			// Step 2: Try to parse with govaluate
// 			callLog := []string{}

// 			functions := map[string]govaluate.ExpressionFunction{
// 				"OK": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						callLog = append(callLog, "OK("+args[0].(string)+")")
// 					}
// 					return true, nil
// 				},
// 				"ALARM": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						callLog = append(callLog, "ALARM("+args[0].(string)+")")
// 					}
// 					return true, nil
// 				},
// 				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						callLog = append(callLog, "INSUFFICIENT_DATA("+args[0].(string)+")")
// 					}
// 					return true, nil
// 				},
// 			}

// 			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tt.want, functions)
// 			if err != nil {
// 				t.Fatalf("Govaluate failed to parse: %v\nExpression: %s", err, tt.want)
// 			}

// 			result, err := expr.Evaluate(nil)
// 			if err != nil {
// 				t.Fatalf("Govaluate failed to evaluate: %v\nExpression: %s", err, tt.want)
// 			}

// 			t.Logf("✅ Input: %s", tt.input)
// 			t.Logf("   Transformed: %s", tt.want)
// 			t.Logf("   Govaluate result: %v", result)
// 			t.Logf("   Function calls: %v", callLog)
// 		})
// 	}
// }

// TestGovaluateReceivesCorrectArguments sprawdza dokładnie co govaluate otrzymuje jako argumenty
// func TestGovaluateReceivesCorrectArguments(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		transformed      string
// 		expectedFunction string
// 		expectedArg      string // czego oczekujemy jako argument
// 	}{
// 		{
// 			name:             "Simple name",
// 			transformed:      `OK('test')`,
// 			expectedFunction: "OK",
// 			expectedArg:      "test",
// 		},
// 		{
// 			name:             "Name with escaped quotes",
// 			transformed:      `ALARM('\"DobryAlarm\"')`,
// 			expectedFunction: "ALARM",
// 			expectedArg:      `"DobryAlarm"`, // oczekujemy że govaluate usunie \ i zostawi "
// 		},
// 		{
// 			name:             "Name with parenthesis",
// 			transformed:      `OK('\"a)\"')`,
// 			expectedFunction: "OK",
// 			expectedArg:      `"a)"`,
// 		},
// 		{
// 			name:             "Name with escaped apostrophe at start",
// 			transformed:      `OK('\'a')`,
// 			expectedFunction: "OK",
// 			expectedArg:      `'a`, // oczekujemy 'a (apostrof + a)
// 		},
// 		{
// 			name:             "Name ending with escaped apostrophe",
// 			transformed:      `OK('a\'')`,
// 			expectedFunction: "OK",
// 			expectedArg:      `a'`, // oczekujemy a'
// 		},
// 		{
// 			name:             "Name with multiple escaped quotes",
// 			transformed:      `OK('\"\'a\'\"')`,
// 			expectedFunction: "OK",
// 			expectedArg:      `"'a'"`,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			var receivedArg interface{}

// 			functions := map[string]govaluate.ExpressionFunction{
// 				"OK": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						receivedArg = args[0]
// 					}
// 					return true, nil
// 				},
// 				"ALARM": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						receivedArg = args[0]
// 					}
// 					return true, nil
// 				},
// 				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						receivedArg = args[0]
// 					}
// 					return true, nil
// 				},
// 			}

// 			expr, err := govaluate.NewEvaluableExpressionWithFunctions(tt.transformed, functions)
// 			if err != nil {
// 				t.Fatalf("Govaluate parse failed: %v", err)
// 			}

// 			_, err = expr.Evaluate(nil)
// 			if err != nil {
// 				t.Fatalf("Govaluate evaluate failed: %v", err)
// 			}

// 			receivedStr, ok := receivedArg.(string)
// 			if !ok {
// 				t.Fatalf("Received arg is not string: %T %v", receivedArg, receivedArg)
// 			}

// 			if receivedStr != tt.expectedArg {
// 				t.Errorf("%s received wrong argument:\ngot:  %q\nwant: %q",
// 					tt.expectedFunction, receivedStr, tt.expectedArg)
// 			} else {
// 				t.Logf("✅ %s received correct arg: %q", tt.expectedFunction, receivedStr)
// 			}
// 		})
// 	}
// }

// // TestFullPipeline testuje pełny pipeline: AWS → Transform → Govaluate
// func TestFullPipeline(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		awsInput      string
// 		expectedCalls []string // oczekiwane wywołania funkcji
// 	}{
// 		{
// 			name:          "Simple alarm",
// 			awsInput:      `ALARM("DobryAlarm")`,
// 			expectedCalls: []string{`ALARM("DobryAlarm")`},
// 		},
// 		{
// 			name:          "Alarm with parenthesis in name",
// 			awsInput:      `OK("a)")`,
// 			expectedCalls: []string{`OK("a)")`},
// 		},
// 		{
// 			name:          "Two alarms with AND",
// 			awsInput:      `ALARM("Alarm1") AND ALARM("Alarm2")`,
// 			expectedCalls: []string{`ALARM("Alarm1")`, `ALARM("Alarm2")`},
// 		},
// 		{
// 			name:          "Complex expression",
// 			awsInput:      `(OK("Health") OR ALARM("Error")) AND NOT ALARM("Maintenance")`,
// 			expectedCalls: []string{`OK("Health")`, `ALARM("Error")`, `ALARM("Maintenance")`},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Step 1: Transform AWS format to govaluate format
// 			transformed, err := TransformAlarmRule(tt.awsInput)
// 			if err != nil {
// 				t.Fatalf("TransformAlarmRule failed: %v", err)
// 			}

// 			t.Logf("AWS Input:    %s", tt.awsInput)
// 			t.Logf("Transformed:  %s", transformed)

// 			// Step 2: Create govaluate expression with mock functions
// 			callLog := []string{}

// 			functions := map[string]govaluate.ExpressionFunction{
// 				"OK": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						call := `OK("` + args[0].(string) + `")`
// 						callLog = append(callLog, call)
// 						t.Logf("OK() called with: %q", args[0])
// 					}
// 					return true, nil
// 				},
// 				"ALARM": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						call := `ALARM("` + args[0].(string) + `")`
// 						callLog = append(callLog, call)
// 						t.Logf("ALARM() called with: %q", args[0])
// 					}
// 					return true, nil
// 				},
// 				"INSUFFICIENT_DATA": func(args ...interface{}) (interface{}, error) {
// 					if len(args) > 0 {
// 						call := `INSUFFICIENT_DATA("` + args[0].(string) + `")`
// 						callLog = append(callLog, call)
// 						t.Logf("INSUFFICIENT_DATA() called with: %q", args[0])
// 					}
// 					return true, nil
// 				},
// 			}

// 			// Step 3: Evaluate
// 			expr, err := govaluate.NewEvaluableExpressionWithFunctions(transformed, functions)
// 			if err != nil {
// 				t.Fatalf("Failed to create govaluate expression: %v", err)
// 			}

// 			result, err := expr.Evaluate(nil)
// 			if err != nil {
// 				t.Fatalf("Failed to evaluate: %v", err)
// 			}

// 			t.Logf("Result: %v", result)
// 			t.Logf("Call log: %v", callLog)

// 		})
// 	}
// }

// // Benchmark dla performance
// func BenchmarkTransformAlarmRule(b *testing.B) {
// 	input := `(ALARM("CPU1") OR ALARM("CPU2")) AND NOT ALARM("Deploying")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, _ = TransformAlarmRule(input)
// 	}
// }

// func BenchmarkTransformAlarmRuleSimple(b *testing.B) {
// 	input := `ALARM("SimpleAlarm")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, _ = TransformAlarmRule(input)
// 	}
// }

// func BenchmarkTransformAlarmRuleComplex(b *testing.B) {
// 	input := `((ALARM("A") AND ALARM("B")) OR (OK("C") AND OK("D"))) AND NOT INSUFFICIENT_DATA("E")`

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, _ = TransformAlarmRule(input)
// 	}
// }
