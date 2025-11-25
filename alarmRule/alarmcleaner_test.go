package alarmcleaner

import "testing"

func TestPerfectLogic(t *testing.T) {
	cases := map[string]string{
		`NOT TRUE`:                `!true`,
		`(NOT TRUE)`:              `(!true)`,
		`NOT FALSE`:               `!false`,
		`NOT(TRUE)`:               `!(true)`,
		`NOT(FALSE)`:              `!(false)`,
		`NOT ( TRUE )`:            `!(true)`,
		`NOT (FALSE)`:             `!(false)`,
		`NOT  TRUE`:               `!true`,
		`NOT   FALSE  OR OK(cpu)`: `!false OR OK('cpu')`,
		`ALARM(x) AND NOT(FALSE)`: `ALARM('x') AND !(false)`,
		`NOT TRUE AND ALARM(y)`:   `!true AND ALARM('y')`,
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
