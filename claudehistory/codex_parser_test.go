package claudehistory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testCodexSession = strings.Join([]string{
	`{"timestamp":"2026-02-17T08:00:00.000Z","type":"session_meta","payload":{"id":"test-session-1","cwd":"/tmp/project","cli_version":"0.101.0","source":"exec","model_provider":"openai"}}`,
	`{"timestamp":"2026-02-17T08:00:01.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
	`{"timestamp":"2026-02-17T08:00:02.000Z","type":"event_msg","payload":{"type":"agent_reasoning","text":"**Analyzing the codebase**\n\nLooking at the directory structure."}}`,
	`{"timestamp":"2026-02-17T08:00:03.000Z","type":"response_item","payload":{"type":"reasoning","summary":[{"type":"summary_text","text":"**Reading files**"}]}}`,
	`{"timestamp":"2026-02-17T08:00:04.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"git diff HEAD\",\"workdir\":\"/tmp/project\"}","call_id":"call_abc123"}}`,
	`{"timestamp":"2026-02-17T08:00:05.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call_abc123","output":"Chunk ID: abc\nWall time: 0.05 seconds\nProcess exited with code 0\nOriginal token count: 100\nOutput:\ndiff --git a/main.go b/main.go\n+added line"}}`,
	`{"timestamp":"2026-02-17T08:00:06.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"go test ./...\"}","call_id":"call_def456"}}`,
	`{"timestamp":"2026-02-17T08:00:07.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call_def456","output":"Chunk ID: def\nWall time: 1.2 seconds\nProcess exited with code 0\nOriginal token count: 50\nOutput:\nok  \tgithub.com/example/pkg\t0.5s"}}`,
	`{"timestamp":"2026-02-17T08:00:08.000Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Here is my analysis of the code."}]}}`,
	`{"timestamp":"2026-02-17T08:00:09.000Z","type":"event_msg","payload":{"type":"agent_message","message":"The code looks good overall."}}`,
	`{"timestamp":"2026-02-17T08:00:10.000Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":5000,"cached_input_tokens":3000,"output_tokens":500,"total_tokens":5500},"last_token_usage":{"input_tokens":5000,"cached_input_tokens":3000,"output_tokens":500,"total_tokens":5500}}}}`,
	`{"timestamp":"2026-02-17T08:00:11.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
}, "\n")

func TestParseCodexLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantType    string
		wantPayload string
	}{
		{
			name:        "session_meta",
			line:        `{"timestamp":"2026-02-17T08:00:00.000Z","type":"session_meta","payload":{"id":"abc","cwd":"/tmp"}}`,
			wantType:    "session_meta",
			wantPayload: "abc",
		},
		{
			name:        "function_call",
			line:        `{"timestamp":"2026-02-17T08:00:00.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"ls\"}","call_id":"call_1"}}`,
			wantType:    "response_item",
			wantPayload: "function_call",
		},
		{
			name:        "agent_reasoning",
			line:        `{"timestamp":"2026-02-17T08:00:00.000Z","type":"event_msg","payload":{"type":"agent_reasoning","text":"thinking..."}}`,
			wantType:    "event_msg",
			wantPayload: "agent_reasoning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseCodexLine(tt.line)
			if err != nil {
				t.Fatalf("ParseCodexLine() error: %v", err)
			}
			if event.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", event.Type, tt.wantType)
			}
			if tt.wantType == "session_meta" && event.Payload.ID != tt.wantPayload {
				t.Errorf("Payload.ID = %q, want %q", event.Payload.ID, tt.wantPayload)
			}
			if tt.wantType != "session_meta" && event.Payload.Type != tt.wantPayload {
				t.Errorf("Payload.Type = %q, want %q", event.Payload.Type, tt.wantPayload)
			}
		})
	}
}

func TestExtractCodexToolUses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-session.jsonl")
	if err := os.WriteFile(path, []byte(testCodexSession), 0o644); err != nil {
		t.Fatal(err)
	}

	toolUses, err := ExtractCodexToolUses(path)
	if err != nil {
		t.Fatalf("ExtractCodexToolUses() error: %v", err)
	}

	counts := map[string]int{}
	for _, tu := range toolUses {
		counts[tu.Tool]++
	}

	if counts["CodexCommand"] != 2 {
		t.Errorf("CodexCommand count = %d, want 2", counts["CodexCommand"])
	}
	// agent_reasoning + reasoning summary = 2 reasoning events
	if counts["CodexReasoning"] != 2 {
		t.Errorf("CodexReasoning count = %d, want 2", counts["CodexReasoning"])
	}
	// assistant message + agent_message = 2 message events
	if counts["CodexMessage"] != 2 {
		t.Errorf("CodexMessage count = %d, want 2", counts["CodexMessage"])
	}

	// Verify command extraction
	for _, tu := range toolUses {
		if tu.Tool != "CodexCommand" {
			continue
		}
		cmd, _ := tu.Input["command"].(string)
		if cmd != "git diff HEAD" && cmd != "go test ./..." {
			t.Errorf("unexpected command: %q", cmd)
		}
		if tu.Source != "codex" {
			t.Errorf("Source = %q, want codex", tu.Source)
		}
		if tu.SessionID != "test-session-1" {
			t.Errorf("SessionID = %q, want test-session-1", tu.SessionID)
		}
		if tu.CWD != "/tmp/project" {
			t.Errorf("CWD = %q, want /tmp/project", tu.CWD)
		}
	}

	// Verify output parsing (strips metadata prefix)
	for _, tu := range toolUses {
		if tu.Tool != "CodexCommand" {
			continue
		}
		cmd, _ := tu.Input["command"].(string)
		if cmd != "git diff HEAD" {
			continue
		}
		output, _ := tu.Input["output"].(string)
		if !strings.HasPrefix(output, "diff --git") {
			t.Errorf("output should start with 'diff --git', got: %q", output[:min(50, len(output))])
		}
	}
}

func TestExtractCommand(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{`{"cmd":"git status"}`, "git status"},
		{`{"cmd":"ls -la","workdir":"/tmp"}`, "ls -la"},
		{"", ""},
		{`invalid`, `invalid`},
	}

	for _, tt := range tests {
		got := extractCommand(tt.args)
		if got != tt.want {
			t.Errorf("extractCommand(%q) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

func TestExtractCommandOutput(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{
			raw:  "Chunk ID: abc\nWall time: 0.05\nOutput:\nhello world",
			want: "hello world",
		},
		{
			raw:  "no output marker here",
			want: "no output marker here",
		},
	}

	for _, tt := range tests {
		got := extractCommandOutput(tt.raw)
		if got != tt.want {
			t.Errorf("extractCommandOutput() = %q, want %q", got, tt.want)
		}
	}
}

func TestFilterBySource(t *testing.T) {
	toolUses := []ToolUse{
		{Tool: "Bash", Source: "claude"},
		{Tool: "CodexCommand", Source: "codex"},
		{Tool: "Read", Source: "claude"},
	}

	claudeOnly := FilterToolUses(toolUses, Filter{Source: "claude"})
	if len(claudeOnly) != 2 {
		t.Errorf("claude filter: got %d, want 2", len(claudeOnly))
	}

	codexOnly := FilterToolUses(toolUses, Filter{Source: "codex"})
	if len(codexOnly) != 1 {
		t.Errorf("codex filter: got %d, want 1", len(codexOnly))
	}

	all := FilterToolUses(toolUses, Filter{})
	if len(all) != 3 {
		t.Errorf("no filter: got %d, want 3", len(all))
	}
}

func TestMalformedCodexLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad-session.jsonl")
	content := strings.Join([]string{
		`{"timestamp":"2026-02-17T08:00:00.000Z","type":"session_meta","payload":{"id":"s1","cwd":"/tmp"}}`,
		`not valid json`,
		`{"timestamp":"2026-02-17T08:00:01.000Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"echo hi\"}","call_id":"c1"}}`,
		`{"timestamp":"2026-02-17T08:00:02.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"c1","output":"Output:\nhi"}}`,
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	toolUses, err := ExtractCodexToolUses(path)
	if err != nil {
		t.Fatalf("should not error on malformed lines: %v", err)
	}
	if len(toolUses) != 1 {
		t.Errorf("got %d tool uses, want 1 (the valid command)", len(toolUses))
	}
}
