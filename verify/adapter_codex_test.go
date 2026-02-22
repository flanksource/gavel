package verify

import (
	"strings"
	"testing"
)

func TestCodexParseResponse(t *testing.T) {
	adapter := Codex{}

	jsonPayload := `{"checks":{"tests-added":{"pass":true},"null-safety":{"pass":false,"evidence":[{"file":"main.go","line":5,"message":"nil ptr"}]}},"ratings":{"security":{"score":80}},"completeness":{"pass":true,"summary":"ok"}}`

	tests := []struct {
		name       string
		input      string
		wantChecks int
		wantErr    bool
	}{
		{
			name: "JSONL with agent_message containing JSON",
			input: `{"type":"thread.started","thread_id":"abc"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"git diff"}}
{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"` + strings.ReplaceAll(jsonPayload, `"`, `\"`) + `"}}
{"type":"turn.completed","usage":{"input_tokens":100,"output_tokens":50}}`,
			wantChecks: 2,
		},
		{
			name:       "JSONL with markdown-fenced JSON in agent_message",
			input:      `{"type":"item.completed","item":{"type":"agent_message","text":"` + "```json\\n" + strings.ReplaceAll(jsonPayload, `"`, `\"`) + "\\n```" + `"}}`,
			wantChecks: 2,
		},
		{
			name:    "invalid JSONL",
			input:   `{"type":"thread.started"}`,
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

func TestCodexBuildFixArgs(t *testing.T) {
	adapter := Codex{}

	t.Run("full-auto", func(t *testing.T) {
		args := adapter.BuildFixArgs("codex-mini", "fix this", false)
		joined := strings.Join(args, " ")
		for _, want := range []string{"exec", "--full-auto", "-m", "codex-mini"} {
			if !strings.Contains(joined, want) {
				t.Errorf("args %q should contain %q", joined, want)
			}
		}
	})

	t.Run("patch-only", func(t *testing.T) {
		args := adapter.BuildFixArgs("", "fix this", true)
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "--full-auto") {
			t.Error("patch-only should not include --full-auto")
		}
		if !strings.Contains(joined, "exec") {
			t.Error("should include exec subcommand")
		}
	})
}
