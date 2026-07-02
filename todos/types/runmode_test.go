package types

import "testing"

func TestParseRunMode(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    RunMode
		wantErr bool
	}{
		{name: "empty defaults to run", input: "", want: ModeRun},
		{name: "run", input: "run", want: ModeRun},
		{name: "plan", input: "plan", want: ModePlan},
		{name: "verify", input: "verify", want: ModeVerify},
		{name: "legacy cmux mechanism is rejected", input: "cmux", wantErr: true},
		{name: "legacy inline mechanism is rejected", input: "inline", wantErr: true},
		{name: "unknown value is rejected", input: "bogus", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRunMode(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseRunMode(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRunMode(%q): %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("ParseRunMode(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRunModeValid(t *testing.T) {
	for _, m := range []RunMode{ModeRun, ModePlan, ModeVerify} {
		if !m.Valid() {
			t.Errorf("RunMode(%q).Valid() = false, want true", m)
		}
	}
	for _, m := range []RunMode{"", "cmux", "bogus"} {
		if m.Valid() {
			t.Errorf("RunMode(%q).Valid() = true, want false", m)
		}
	}
}
