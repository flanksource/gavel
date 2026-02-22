package verify

import "testing"

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"json fences", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"yaml fences", "```yaml\na: 1\n```", "a: 1"},
		{"plain fences", "```\nsome text\n```", "some text"},
		{"no fences", `{"a":1}`, `{"a":1}`},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripMarkdownFences(tt.input); got != tt.want {
				t.Errorf("stripMarkdownFences() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractYAMLBlock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with yaml block", "---\nchecks:\n  a: true\n---", "checks:\n  a: true"},
		{"no separators", "just text", ""},
		{"single separator", "---\nfoo", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractYAMLBlock(tt.input); got != tt.want {
				t.Errorf("extractYAMLBlock() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTryUnmarshalResult(t *testing.T) {
	validJSON := `{"checks":{"a":{"pass":true}},"ratings":{"sec":{"score":90}},"completeness":{"pass":true,"summary":"ok"}}`
	validYAML := "checks:\n  a:\n    pass: true\nratings:\n  sec:\n    score: 90\ncompleteness:\n  pass: true\n  summary: ok"

	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"valid JSON", validJSON, true},
		{"valid YAML", validYAML, true},
		{"empty checks", `{"checks":{},"ratings":{}}`, false},
		{"garbage", "not valid at all", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := tryUnmarshalResult(tt.input)
			if ok != tt.wantOK {
				t.Errorf("tryUnmarshalResult() ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && len(result.Checks) == 0 {
				t.Error("expected non-empty checks")
			}
		})
	}
}
