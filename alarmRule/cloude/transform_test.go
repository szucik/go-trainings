package alarmrule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformAlarmRule(t *testing.T) {
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

		// Normalization
		{name: "(TRUE)AND TRUE", input: "(TRUE)AND TRUE", want: "(true) && true"},
		{name: "TRUE AND(TRUE)", input: "TRUE AND(TRUE)", want: "true && (true)"},
		{name: "(TRUE)AND(FALSE)", input: "(TRUE)AND(FALSE)", want: "(true) && (false)"},

		// NOT with alarm functions
		{name: "NOT OK(a)", input: "NOT OK(a)", want: "!OK('a')"},
		{name: "NOT ALARM(a)", input: "NOT ALARM(a)", want: "!ALARM('a')"},
		{name: "NOT INSUFFICIENT_DATA(a)", input: "NOT INSUFFICIENT_DATA(a)", want: "!INSUFFICIENT_DATA('a')"},

		// Basic alarm states
		{name: "OK(a)", input: "OK(a)", want: "OK('a')"},
		{name: "ALARM(a)", input: "ALARM(a)", want: "ALARM('a')"},
		{name: "INSUFFICIENT_DATA(a)", input: "INSUFFICIENT_DATA(a)", want: "INSUFFICIENT_DATA('a')"},

		// Quotes handling - AWS removes only outer pair
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

		// Special characters
		{name: "Hyphens in name", input: "OK(my-alarm-123)", want: "OK('my-alarm-123')"},
		{name: "Slashes in name", input: "ALARM(prod/web/cpu)", want: "ALARM('prod/web/cpu')"},

		// Real AWS examples
		{name: "Two alarms with AND", input: "ALARM(CPUUtilizationTooHigh) AND ALARM(DiskReadOpsTooHigh)", want: "ALARM('CPUUtilizationTooHigh') && ALARM('DiskReadOpsTooHigh')"},
		{name: "AND with NOT", input: "ALARM(CPUUtilizationTooHigh) AND NOT ALARM(DeploymentInProgress)", want: "ALARM('CPUUtilizationTooHigh') && !ALARM('DeploymentInProgress')"},
		{name: "Nested OR and NOT", input: "(ALARM(WebServer1CPU) OR ALARM(WebServer2CPU)) AND NOT ALARM(MaintenanceWindow)", want: "(ALARM('WebServer1CPU') || ALARM('WebServer2CPU')) && !ALARM('MaintenanceWindow')"},

		// With quotes
		{name: "Double quotes", input: `ALARM("testAlarm1")`, want: `ALARM('testAlarm1')`},
		{name: "Two alarms with quotes", input: `ALARM("testAlarm1") AND ALARM("testAlarm2")`, want: `ALARM('testAlarm1') && ALARM('testAlarm2')`},
		{name: "With newlines", input: "ALARM(\n\"testAlarm1\"\n) OR ALARM(\n\"testAlarm2\"\n)", want: `ALARM('testAlarm1') || ALARM('testAlarm2')`},
		{name: "With tabs", input: "ALARM(\t\"CPUHigh\"\t) AND ALARM(\t\"MemHigh\"\t)", want: `ALARM('CPUHigh') && ALARM('MemHigh')`},
		{name: "Mixed whitespace", input: "ALARM( \n\t \"testAlarm\" \n\t )", want: `ALARM('testAlarm')`},
		{name: "Newlines in expression", input: "ALARM(\n\"alarm1\"\n) AND\nNOT ALARM(\n\"alarm2\"\n)", want: `ALARM('alarm1') && !ALARM('alarm2')`},
		{name: "Multiline formatted", input: "ALARM(\n\"prod-cpu\"\n) AND ALARM(\n\"prod-memory\"\n)", want: `ALARM('prod-cpu') && ALARM('prod-memory')`},
		{name: "ARN format", input: `ALARM("arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm")`, want: `ALARM('arn:aws:cloudwatch:us-east-1:123456789012:alarm:MyAlarm')`},
		{name: "Three alarms", input: `ALARM("alarm1") AND ALARM("alarm2") OR ALARM("alarm3")`, want: `ALARM('alarm1') && ALARM('alarm2') || ALARM('alarm3')`},
		{name: "Production scenario", input: `ALARM("prod-alert") AND NOT ALARM("maintenance")`, want: `ALARM('prod-alert') && !ALARM('maintenance')`},
		{name: "Complex nested", input: `(ALARM("cpu1") OR ALARM("cpu2")) AND NOT ALARM("deploying")`, want: `(ALARM('cpu1') || ALARM('cpu2')) && !ALARM('deploying')`},
		{name: "Mixed functions", input: `OK("health-check") AND ALARM("error-rate")`, want: `OK('health-check') && ALARM('error-rate')`},
		{name: "With INSUFFICIENT_DATA", input: `ALARM("cpu-high") AND NOT INSUFFICIENT_DATA("metric-missing")`, want: `ALARM('cpu-high') && !INSUFFICIENT_DATA('metric-missing')`},

		// Edge cases
		{name: "Empty input", input: "", want: ""},
		{name: "Whitespace only", input: "   ", want: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := TransformAlarmRule(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
