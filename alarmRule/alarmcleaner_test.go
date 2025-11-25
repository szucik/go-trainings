package alarmcleaner_test

import (
	alarmcleaner "go-training-repo/alarmRule"
	"testing"
)

func TestGodTierFormatter(t *testing.T) {
	cases := map[string]string{
		// === Podstawowe stany (ALARM, OK, INSUFFICIENT_DATA) ===
		`ALARM(cpu-high)`:                 `ALARM('cpu-high')`,
		`OK("mem-ok")`:                    `OK('mem-ok')`,
		`INSUFFICIENT_DATA(disk)`:         `INSUFFICIENT_DATA('disk')`,
		`ALARM(  "  spaces  galore  "  )`: `ALARM('spaces  galore')`,

		// === NOT + stan (poprawne przypadki) ===
		`NOT OK(cpu)`:                `!OK('cpu')`,
		`NOT ( OK("mem") )`:          `!OK('mem')`,
		`NOT ALARM(disk-high)`:       `!ALARM('disk-high')`,
		`NOT INSUFFICIENT_DATA(net)`: `!INSUFFICIENT_DATA('net')`,
		`NOT ( ALARM("bad") )`:       `!ALARM('bad')`,

		// === NOT + true/false (z nawiasami i bez) ===
		`NOT TRUE`:     `!true`,
		`NOT FALSE`:    `!false`,
		`NOT(TRUE)`:    `!(true)`,
		`NOT(FALSE)`:   `!(false)`,
		`NOT ( TRUE )`: `!(true)`,
		`NOT (FALSE)`:  `!(false)`,
		`NOT  TRUE`:    `!true`,
		`NOT   FALSE`:  `!false`,

		// === Błędy ludzkie – "przyklejone" NOT (nie ruszamy!) ===
		`xNOT OK(a)`:            `xNOT OK('a')`,
		`NOTNOT ALARM(x)`:       `NOTNOT ALARM('x')`,
		`xyzNOT OK(cpu)`:        `xyzNOT OK('cpu')`,
		`SuperMegaNOT OK(mem)`:  `SuperMegaNOT OK('mem')`,
		`prefixNOT ALARM(disk)`: `prefixNOT ALARM('disk')`,

		// === Mieszane błędy ludzkie z true/false ===
		`xNOT TRUE`:    `xNOT true`,
		`NOTNOT FALSE`: `NOTNOT false`,
		`xyzNOTTRUE`:   `xyzNOTTRUE`, // nie ruszamy, bo nie ma spacji

		// === Złożone wyrażenia ===
		`NOT OK(cpu) AND ALARM(mem)`:         `!OK('cpu') AND ALARM('mem')`,
		`ALARM(a) OR NOT(FALSE)`:             `ALARM('a') OR !(false)`,
		`NOT TRUE AND OK("good")`:            `!true AND OK('good')`,
		`NOT (TRUE) OR INSUFFICIENT_DATA(x)`: `!(true) OR INSUFFICIENT_DATA('x')`,
		`NOT OK(a) OR NOT OK(b)`:             `!OK('a') OR !OK('b')`,
		`NOT ( OK(a) AND ALARM(b) )`:         `! ( OK('a') AND ALARM('b') )`,

		// === Legacy konkatenacja stringów (najgorsze przypadki) ===
		`ALARM(""+name+"")`:      `ALARM('value')`, // po podstawieniu name="value"
		`OK( "" + alarm2 + "" )`: `OK('alarm2')`,

		// === Prawdziwe potwory z produkcji ===
		`xNOT OK("cpu") AND NOT(FALSE)`:     `xNOT OK('cpu') AND !(false)`,
		`NOTNOT ALARM(x) OR NOT TRUE`:       `NOTNOT ALARM('x') OR !true`,
		`prefixNOT OK(mem) AND ALARM(disk)`: `prefixNOT OK('mem') AND ALARM('disk')`,

		// === TRUE/FALSE same ===
		`TRUE`:           `true`,
		`FALSE`:          `false`,
		`TRUE AND OK(x)`: `true AND OK('x')`,
		`OK(y) OR FALSE`: `OK('y') OR false`,
	}

	for input, expected := range cases {
		t.Run(input, func(t *testing.T) {
			got := alarmcleaner.Format(input)
			if got != expected {
				t.Errorf("\nINPUT:   %q\nGOT:     %q\nEXPECTED:%q", input, got, expected)
			}
		})
	}
}

// Dodatkowy test na bardzo długie, złożone wyrażenie (real world)
func TestRealWorldHell(t *testing.T) {
	input := `NOT ( OK("cpu-steady") ) AND ALARM(high-cpu) OR NOT(FALSE) AND xNOT OK(low-mem)`
	expected := `!OK('cpu-steady') AND ALARM('high-cpu') OR !(false) AND xNOT OK('low-mem')`

	if got := alarmcleaner.Format(input); got != expected {
		t.Errorf("\nINPUT:   %q\nGOT:     %q\nEXPECTED:%q", input, got, expected)
	}
}
