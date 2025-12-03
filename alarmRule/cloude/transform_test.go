package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTransformAlarmRule - wszystkie testy transformacji
func Test_TransformAlarmRule(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := TransformAlarmRule(tc.input)

			// Assert
			assert.Equal(t, tc.want, got, "Transformation should match expected output")
		})
	}
}
