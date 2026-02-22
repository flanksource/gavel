package verify

import "testing"

func TestClaudeParseResponse(t *testing.T) {
	adapter := Claude{}

	tests := []struct {
		name       string
		input      string
		wantChecks int
		wantErr    bool
	}{
		{
			name:       "direct JSON",
			input:      `{"checks":{"tests-added":{"pass":true},"null-safety":{"pass":false,"evidence":[{"file":"main.go","line":10,"message":"nil dereference"}]}},"ratings":{"security":{"score":85}},"completeness":{"pass":true,"summary":"looks good"}}`,
			wantChecks: 2,
		},
		{
			name:       "JSON wrapper with result field",
			input:      `{"result": "{\"checks\":{\"tests-added\":{\"pass\":true}},\"ratings\":{\"security\":{\"score\":90}},\"completeness\":{\"pass\":true,\"summary\":\"ok\"}}"}`,
			wantChecks: 1,
		},
		{
			name:       "prose with embedded JSON",
			input:      `Here is my analysis of the code:` + "\n\n" + `{"checks":{"tests-added":{"pass":true}},"ratings":{"security":{"score":90}},"completeness":{"pass":true,"summary":"ok"}}` + "\n\nOverall the code looks good.",
			wantChecks: 1,
		},
		{
			name:       "markdown fences wrapping JSON",
			input:      "```json\n" + `{"checks":{"tests-added":{"pass":true}},"ratings":{"security":{"score":90}},"completeness":{"pass":true,"summary":"ok"}}` + "\n```",
			wantChecks: 1,
		},
		{
			name:    "invalid input",
			input:   "not json at all {{{",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParseResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Checks) != tt.wantChecks {
				t.Errorf("got %d checks, want %d", len(result.Checks), tt.wantChecks)
			}
		})
	}
}

func TestClaudeBuildVerifyArgs(t *testing.T) {
	tests := []struct {
		model    string
		wantFlag bool
	}{
		{"claude-4-5-haiku", true},
		{"claude-sonnet-4", true},
		{"claude", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.model, func(t *testing.T) {
			args := Claude{}.BuildVerifyArgs("prompt", tc.model, "", false)
			hasModel := false
			for i, a := range args {
				if a == "--model" && i+1 < len(args) && args[i+1] == tc.model {
					hasModel = true
				}
			}
			if hasModel != tc.wantFlag {
				t.Errorf("BuildVerifyArgs(%q) --model present=%v, want %v; args=%v", tc.model, hasModel, tc.wantFlag, args)
			}
		})
	}
}

func TestExtractJSONFromText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"embedded object", `text {"a":1} more`, `{"a":1}`},
		{"nested braces", `prefix {"a":{"b":2}} suffix`, `{"a":{"b":2}}`},
		{"no JSON", "just text", ""},
		{"unmatched brace", "prefix { no close", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractJSONFromText(tt.input); got != tt.want {
				t.Errorf("extractJSONFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}
