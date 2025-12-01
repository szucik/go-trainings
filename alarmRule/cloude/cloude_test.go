package alarmrule

import "testing"

func TestFixAlarmRule(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// Podstawowe - bez cudzysłowów
		{
			name:  "ALARM - prosty tekst",
			input: "ALARM(a)",
			want:  "ALARM('a')",
		},
		{
			name:  "INSUFFICIENT_DATA - prosty tekst",
			input: "INSUFFICIENT_DATA(a)",
			want:  "INSUFFICIENT_DATA('a')",
		},

		// Z podwójnymi cudzysłowami
		{
			name:  "OK - z cudzysłowami",
			input: `OK("a")`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  "OK - cudzysłów na końcu",
			input: `OK(a")`,
			want:  `OK('a\"')`,
		},
		{
			name:  "OK - cudzysłów na początku",
			input: `OK("a)`,
			want:  `OK('\"a')`,
		},

		// Z apostrofami
		{
			name:  "OK - w apostrofach (już poprawne)",
			input: "OK('a')",
			want:  "OK('a')",
		},
		{
			name:  "OK - apostrof na początku",
			input: "OK('a)",
			want:  `OK('\'a')`,
		},
		{
			name:  "OK - apostrof na końcu",
			input: "OK(a')",
			want:  `OK('a\'')`,
		},

		// Mieszane - apostrofy i cudzysłowy
		{
			name:  "OK - cudzysłowy w apostrofach",
			input: `OK("'a'")`,
			want:  `OK('\"\'a\'\"')`,
		},
		{
			name:  "OK - apostrofy wokół cudzysłowów (usuwamy apostrofy)",
			input: `OK('"a"')`,
			want:  `OK('\"a\"')`,
		},
		{
			name:  "OK - mieszane znaki 1",
			input: `OK("'a"')`,
			want:  `OK('\"\'a\"\'')`,
		},
		{
			name:  "OK - mieszane znaki 2",
			input: `OK("'a)`,
			want:  `OK('\"\'a')`,
		},
		{
			name:  "OK - mieszane znaki 3",
			input: `OK('"a)`,
			want:  `OK('\'\"a')`,
		},

		// Edge cases
		{
			name:  "OK - długi tekst",
			input: "OK(my-alarm-name)",
			want:  "OK('my-alarm-name')",
		},
		{
			name:  "OK - ze spacjami",
			input: "OK(my alarm)",
			want:  "OK('my alarm')",
		},

		// Error cases
		{
			name:    "brak nawiasów",
			input:   "OK",
			wantErr: true,
		},
		{
			name:    "brak zamykającego nawiasu",
			input:   "OK(a",
			wantErr: true,
		},
		{
			name:    "brak otwierającego nawiasu",
			input:   "OKa)",
			wantErr: true,
		},
		{
			name:    "małe litery w stanie",
			input:   "ok(a)",
			wantErr: true,
		},
		{
			name:    "nieprawidłowy stan",
			input:   "INVALID(a)",
			wantErr: true,
		},
		{
			name:    "pusty content",
			input:   "OK()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FixAlarmRule(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("FixAlarmRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.want {
				t.Errorf("FixAlarmRule()\ninput: %q\ngot:   %q\nwant:  %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateAlarmState(t *testing.T) {
	tests := []struct {
		state string
		want  bool
	}{
		{"OK", true},
		{"ALARM", true},
		{"INSUFFICIENT_DATA", true},
		{"ok", false},
		{"alarm", false},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := ValidateAlarmState(tt.state); got != tt.want {
				t.Errorf("ValidateAlarmState(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}
