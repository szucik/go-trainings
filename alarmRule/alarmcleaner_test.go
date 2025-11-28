// alarmcleaner/cleaner_test.go
package alarmcleaner_test

import (
	alarmcleaner "go-training-repo/alarmRule"
	"testing"
)

func TestFormat_GodMode(t *testing.T) {
	cases := map[string]string{
		`NOT TRUE`:                         `!true`,
		`NOT FALSE`:                        `!false`,
		`NOT(TRUE)`:                        `!(true)`,
		`NOT ( FALSE )`:                    `!(false)`,
		`NOT OK(cpu)`:                      `!OK('cpu')`,
		`NOT ALARM(disk)`:                  `!ALARM('disk')`,
		`NOT ( OK( \"cpu\" ) )`:            `!OK('cpu')`,
		`NOT ( OK( \\\"cpu\\\" ) )`:        `!OK('cpu')`,
		`NOT ( OK( \\\" high-cpu \\\" ) )`: `!OK('high-cpu')`,
		`NOT ( ALARM( \"?weird-name\" ) )`: `!ALARM('?weird-name')`,
		`NOT ( ?something )`: `
		 ( ?something )`,
		`xNOT OK(cpu)`:    `xNOT OK('cpu')`,
		`SuperNOTOK(mem)`: `SuperNOTOK('mem')`,
		`ALARM( \"disk\" ) AND NOT(FALSE) OR OK(x)`:          `ALARM('disk') AND !(false) OR OK('x')`,
		`NOT ( OK( \" cpu-steady \" ) ) AND ALARM(high-cpu)`: `!OK('cpu-steady') AND ALARM('high-cpu')`,
	}

	for in, want := range cases {
		got := alarmcleaner.Format(in)
		if got != want {
			t.Errorf("\nFAIL\nin:   %q\ngot:  %q\nwant: %q\n", in, got, want)
		} else {
			t.Logf("PASS: %q → %q", in, got)
		}
	}
}
