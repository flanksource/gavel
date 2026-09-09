package report

import (
	"encoding/json"
	"testing"
)

func intPtr(v int) *int { return &v }

func TestResultFileUnmarshal(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantTests    []string // test names, in order
		wantLint     int
		wantError    string
		wantExitCode *int
		wantLogTail  string
		wantCrash    bool
	}{
		{
			name:      "bare array is the plain `gavel test` shape",
			input:     `[{"name":"TestServe","passed":true}]`,
			wantTests: []string{"TestServe"},
		},
		{
			name:      "object with tests and lint is the `--lint` shape",
			input:     `{"tests":[{"name":"TestServe"}],"lint":[{"linter":"golangci"}]}`,
			wantTests: []string{"TestServe"},
			wantLint:  1,
		},
		{
			name:         "composite action crash stub",
			input:        `{"error":"gavel exited 1 before writing results","exit_code":1,"log_tail":"pre-build failed"}`,
			wantError:    "gavel exited 1 before writing results",
			wantExitCode: intPtr(1),
			wantLogTail:  "pre-build failed",
			wantCrash:    true,
		},
		{
			name:         "gavel-written snapshot carries exit code under metadata",
			input:        `{"metadata":{"exit_code":2},"error":"pre-build failed","tests":[]}`,
			wantError:    "pre-build failed",
			wantExitCode: intPtr(2),
			wantCrash:    true,
		},
		{
			name:      "top-level exit code wins over metadata",
			input:     `{"exit_code":3,"metadata":{"exit_code":9},"error":"boom"}`,
			wantError: "boom", wantExitCode: intPtr(3), wantCrash: true,
		},
		{
			name:      "a failing run that still produced results is not a crash",
			input:     `{"tests":[{"name":"TestServe","failed":true}],"error":"exit 1"}`,
			wantTests: []string{"TestServe"},
			wantError: "exit 1",
			wantCrash: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got ResultFile
			if err := json.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(got.Tests) != len(tc.wantTests) {
				t.Fatalf("got %d tests, want %d", len(got.Tests), len(tc.wantTests))
			}
			for i, name := range tc.wantTests {
				if got.Tests[i].Name != name {
					t.Errorf("Tests[%d].Name = %q, want %q", i, got.Tests[i].Name, name)
				}
			}
			if len(got.Lint) != tc.wantLint {
				t.Errorf("got %d linters, want %d", len(got.Lint), tc.wantLint)
			}
			if got.Error != tc.wantError {
				t.Errorf("Error = %q, want %q", got.Error, tc.wantError)
			}
			if got.LogTail != tc.wantLogTail {
				t.Errorf("LogTail = %q, want %q", got.LogTail, tc.wantLogTail)
			}
			switch code := got.ExitCodeValue(); {
			case tc.wantExitCode == nil && code != nil:
				t.Errorf("ExitCodeValue = %d, want nil", *code)
			case tc.wantExitCode != nil && code == nil:
				t.Errorf("ExitCodeValue = nil, want %d", *tc.wantExitCode)
			case tc.wantExitCode != nil && *code != *tc.wantExitCode:
				t.Errorf("ExitCodeValue = %d, want %d", *code, *tc.wantExitCode)
			}
			if got.IsCrash() != tc.wantCrash {
				t.Errorf("IsCrash() = %t, want %t", got.IsCrash(), tc.wantCrash)
			}
		})
	}
}

// A crash stub whose log tail carries a raw control byte is not valid JSON.
// Decoding must fail loudly rather than yielding a silently empty ResultFile —
// this is the exact artifact shape that made PR pages render "no data".
func TestResultFileRejectsRawControlCharacters(t *testing.T) {
	malformed := []byte("{\"error\":\"gavel exited 1\",\"log_tail\":\"to update it: \tgo mod tidy\"}")
	var got ResultFile
	if err := json.Unmarshal(malformed, &got); err == nil {
		t.Fatalf("unmarshal succeeded, want a syntax error; got %+v", got)
	}
}
