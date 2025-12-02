package alarmrule

import "testing"

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

		// Error cases
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
		// Dodaj do TestTransformAlarmRule:

		// Brakujące przypadki z wymagań
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
		{
			name:  `OK(" a ")`,
			input: `OK(" a ")`,
			want:  `OK('\" a \"')`,
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
		// Real AWS scenarios
		{
			name:  "Real AWS - ALARM with double quotes and trailing space",
			input: `ALARM("DobryAlarm") `,
			want:  `ALARM('\"DobryAlarm\"')`,
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
			name:  "Real AWS - two alarms with AND",
			input: `ALARM("DobryAlarm") AND ALARM("dobryAlarm2")`,
			want:  `ALARM('\"DobryAlarm\"') && ALARM('\"dobryAlarm2\"')`,
		},
		// Real AWS scenarios - whitespace handling
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
