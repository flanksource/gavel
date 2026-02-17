package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *AgentMessage
		wantErr bool
	}{
		{
			name:  "system message",
			input: `{"type":"system","session_id":"abc-123","model":"opus","tools":["Bash","Read"]}`,
			want: &AgentMessage{
				Type:      "system",
				SessionID: "abc-123",
				Model:     "opus",
				Tools:     []string{"Bash", "Read"},
			},
		},
		{
			name:  "assistant message",
			input: `{"type":"assistant","text":"I'll fix that bug"}`,
			want:  &AgentMessage{Type: "assistant", Text: "I'll fix that bug"},
		},
		{
			name:  "thinking message",
			input: `{"type":"thinking","text":"Let me analyze the code"}`,
			want:  &AgentMessage{Type: "thinking", Text: "Let me analyze the code"},
		},
		{
			name:  "tool_use message",
			input: `{"type":"tool_use","tool":"Bash","input":{"command":"go test ./..."}}`,
			want: &AgentMessage{
				Type:  "tool_use",
				Tool:  "Bash",
				Input: map[string]any{"command": "go test ./..."},
			},
		},
		{
			name:  "result message",
			input: `{"type":"result","success":true,"session_id":"abc","cost_usd":0.05,"num_turns":3,"duration_ms":12000,"usage":{"input_tokens":1000,"output_tokens":500}}`,
			want: &AgentMessage{
				Type:       "result",
				Success:    true,
				SessionID:  "abc",
				CostUSD:    0.05,
				NumTurns:   3,
				DurationMs: 12000,
				Usage:      &AgentUsage{InputTokens: 1000, OutputTokens: 500},
			},
		},
		{
			name:  "error message",
			input: `{"type":"error","message":"API rate limit exceeded"}`,
			want:  &AgentMessage{Type: "error", Message: "API rate limit exceeded"},
		},
		{
			name:  "blank line returns nil",
			input: "   ",
			want:  nil,
		},
		{
			name:  "empty string returns nil",
			input: "",
			want:  nil,
		},
		{
			name:    "invalid JSON",
			input:   `{not json}`,
			wantErr: true,
		},
		{
			name:  "line with surrounding whitespace",
			input: `  {"type":"assistant","text":"hello"}  `,
			want:  &AgentMessage{Type: "assistant", Text: "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine([]byte(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
