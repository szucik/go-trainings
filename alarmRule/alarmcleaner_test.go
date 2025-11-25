package alarmcleaner

import "testing"

func TestPerfectLogic(t *testing.T) {
	cases := map[string]string{
		`NOT TRUE`:                   `!true`,
		`(NOT TRUE)`:                 `(!true)`,
		`NOT FALSE`:                  `!false`,
		`NOT(TRUE)`:                  `!(true)`,
		`NOT(FALSE)`:                 `!(false)`,
		`NOT ( TRUE )`:               `!(true)`,
		`NOT (FALSE)`:                `!(false)`,
		`NOT  TRUE`:                  `!true`,
		`NOT   FALSE  OR OK(cpu)`:    `!false OR OK('cpu')`,
		`ALARM(x) AND NOT(FALSE)`:    `ALARM('x') AND !(false)`,
		`NOT TRUE AND ALARM(y)`:      `!true AND ALARM('y')`,
		`NOT OK(cpu)`:                `!OK('cpu')`,
		`NOT ( OK("mem") )`:          `!OK('mem')`,
		`NOT ALARM(disk-high)`:       `!ALARM('disk-high')`,
		`NOT INSUFFICIENT_DATA(net)`: `!INSUFFICIENT_DATA('net')`,
		`NOT OK(a) AND ALARM(b)`:     `!OK('a') AND ALARM('b')`,
		`NOT (TRUE)`:                 `!(true)`,
		`ALARM(cpu) OR NOT ALARM(mem) AND NOT(FALSE)`:                         `ALARM('cpu') OR !ALARM('mem') AND !(false)`,
		`  ALARM(  cpu-1 )   OR   NOT   ALARM(  mem-2 )  AND  NOT( FALSE )  `: `ALARM('cpu-1') OR !ALARM('mem-2') AND !(false)`,
	}

	for in, want := range cases {
		got := Format(in)
		if got != want {
			t.Errorf("FAIL\nin:   %q\ngot:  %q\nwant: %q", in, got, want)
		} else {
			t.Logf("PASS: %q → %q", in, got)
		}
	}
}
