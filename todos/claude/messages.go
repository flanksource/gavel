package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/claudehistory"
	"github.com/flanksource/gavel/todos"
)

type AgentMessage struct {
	Type       string         `json:"type"`
	SessionID  string         `json:"session_id,omitempty"`
	Model      string         `json:"model,omitempty"`
	Tools      []string       `json:"tools,omitempty"`
	Text       string         `json:"text,omitempty"`
	Tool       string         `json:"tool,omitempty"`
	Input      map[string]any `json:"input,omitempty"`
	Success    bool           `json:"success,omitempty"`
	Subtype    string         `json:"subtype,omitempty"`
	CostUSD    float64        `json:"cost_usd,omitempty"`
	NumTurns   int            `json:"num_turns,omitempty"`
	DurationMs int            `json:"duration_ms,omitempty"`
	Usage      *AgentUsage    `json:"usage,omitempty"`
	Errors     []string       `json:"errors,omitempty"`
	ResultText string         `json:"result_text,omitempty"`
	Message    string         `json:"message,omitempty"`
}

type AgentUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func ParseLine(line []byte) (*AgentMessage, error) {
	line = trimWhitespace(line)
	if len(line) == 0 {
		return nil, nil
	}
	var msg AgentMessage
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &msg, nil
}

func trimWhitespace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && (b[start] == ' ' || b[start] == '\t' || b[start] == '\n' || b[start] == '\r') {
		start++
	}
	for end > start && (b[end-1] == ' ' || b[end-1] == '\t' || b[end-1] == '\n' || b[end-1] == '\r') {
		end--
	}
	return b[start:end]
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "..."
}

func toolUseSummary(tool string, input map[string]any) string {
	str := func(key string) string {
		if v, ok := input[key].(string); ok {
			return v
		}
		return ""
	}

	switch tool {
	case "Bash":
		return truncate(str("command"), 80)
	case "Edit", "Read", "Write":
		return str("file_path")
	case "Grep":
		return fmt.Sprintf("%s %s", str("pattern"), str("path"))
	case "Glob":
		return str("pattern")
	case "Task":
		if desc := str("description"); desc != "" {
			return truncate(desc, 80)
		}
		return truncate(str("prompt"), 80)
	case "TodoWrite":
		return todoWriteSummary(input)
	default:
		return tool
	}
}

func todoWriteSummary(input map[string]any) string {
	todosRaw, ok := input["todos"].([]any)
	if !ok || len(todosRaw) == 0 {
		return "TodoWrite (empty)"
	}
	var subjects []string
	for _, t := range todosRaw {
		m, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if subject, ok := m["subject"].(string); ok && subject != "" {
			subjects = append(subjects, truncate(subject, 60))
		} else if content, ok := m["content"].(string); ok && content != "" {
			subjects = append(subjects, truncate(content, 60))
		}
	}
	if len(subjects) == 0 {
		return fmt.Sprintf("TodoWrite (%d todos)", len(todosRaw))
	}
	return fmt.Sprintf("TodoWrite: %s", strings.Join(subjects, ", "))
}

func ProcessMessage(ctx *todos.ExecutorContext, msg *AgentMessage, result *todos.ExecutionResult) {
	transcript := ctx.GetTranscript()

	switch msg.Type {
	case "system":
		ctx.Notify(todos.Notification{
			Type:    todos.NotifyProgress,
			Message: fmt.Sprintf("Session started: %s", msg.SessionID),
			Data:    map[string]any{"session_id": msg.SessionID},
		})

	case "assistant":
		transcript.AddExecutorMessage(truncate(msg.Text, 200), todos.EntryText, nil)

	case "thinking":
		transcript.AddExecutorMessage(msg.Text, todos.EntryThinking, nil)
		ctx.Notify(todos.Notification{
			Type:    todos.NotifyThinking,
			Message: truncate(msg.Text, 100),
		})

	case "tool_use":
		toolUse := claudehistory.ToolUse{Tool: msg.Tool, Input: msg.Input}
		fmt.Println(toolUse.Pretty().ANSI())

		action := toolUseSummary(msg.Tool, msg.Input)
		transcript.AddExecutorMessage(action, todos.EntryAction, map[string]any{
			"tool":   msg.Tool,
			"action": action,
		})

	case "result":
		result.Success = msg.Success
		result.CostUSD = msg.CostUSD
		result.NumTurns = msg.NumTurns
		if msg.Usage != nil {
			result.TokensUsed = msg.Usage.InputTokens + msg.Usage.OutputTokens
		}
		if len(msg.Errors) > 0 {
			result.ErrorMessage = fmt.Sprintf("SDK errors: %v", msg.Errors)
		}
		ctx.Notify(todos.Notification{
			Type:    todos.NotifyCompletion,
			Message: fmt.Sprintf("Completed (subtype=%s, turns=%d, cost=$%.4f)", msg.Subtype, msg.NumTurns, msg.CostUSD),
			Data: map[string]any{
				"session_id": msg.SessionID,
				"cost":       msg.CostUSD,
				"turns":      msg.NumTurns,
				"tokens":     result.TokensUsed,
				"errors":     msg.Errors,
			},
		})

	case "error":
		ctx.Notify(todos.Notification{
			Type:    todos.NotifyError,
			Message: msg.Message,
		})
	}
}
