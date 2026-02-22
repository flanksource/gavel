package verify

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/gavel/claudehistory"
)

type Codex struct{}

func (Codex) Name() string { return "codex" }

func (Codex) BuildVerifyArgs(prompt, model, schemaFile string, _ bool) []string {
	args := []string{"exec", "--json"}
	if model != "" && model != "codex" {
		args = append(args, "-m", model)
	}
	if schemaFile != "" {
		args = append(args, "--output-schema", schemaFile)
	}
	args = append(args, "--", prompt)
	return args
}

func (Codex) BuildFixArgs(model, prompt string, patchOnly bool) []string {
	args := []string{"exec"}
	if !patchOnly {
		args = append(args, "--full-auto")
	}
	if model != "" && model != "codex" {
		args = append(args, "-m", model)
	}
	args = append(args, "--", prompt)
	return args
}

func (Codex) ParseResponse(raw string) (VerifyResult, error) {
	if text := extractFromJSONL(raw); text != "" {
		cleaned := strings.TrimSpace(stripMarkdownFences(text))
		if result, ok := tryUnmarshalResult(cleaned); ok {
			return result, nil
		}
	}
	if result, ok := tryUnmarshalResult(raw); ok {
		return result, nil
	}
	return VerifyResult{}, parseError(raw)
}

func (Codex) PostExecute(raw string) {
	printCodexEvents(raw)
}

func extractFromJSONL(raw string) string {
	var lastMessage string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if json.Unmarshal([]byte(line), &event) != nil {
			continue
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" && event.Item.Text != "" {
			lastMessage = event.Item.Text
		}
	}
	return lastMessage
}

func (Codex) ListModels() ([]string, error) {
	key := getEnv("OPENAI_API_KEY")
	if key == "" {
		return nil, nil
	}
	return fetchModelIDs("https://api.openai.com/v1/models", "Authorization", "Bearer "+key, "")
}

func printCodexEvents(raw string) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		event, err := claudehistory.ParseCodexLine(line)
		if err != nil {
			continue
		}
		var tu *claudehistory.ToolUse
		switch event.Type {
		case "response_item":
			if event.Payload.Type == "reasoning" {
				var text string
				for _, s := range event.Payload.Summary {
					if s.Text != "" {
						text = s.Text
					}
				}
				if text != "" {
					tu = &claudehistory.ToolUse{Tool: "CodexReasoning", Input: map[string]any{"text": text}}
				}
			}
		case "event_msg":
			if event.Payload.Type == "agent_reasoning" && event.Payload.Text != "" {
				tu = &claudehistory.ToolUse{Tool: "CodexReasoning", Input: map[string]any{"text": event.Payload.Text}}
			}
		}
		if tu != nil {
			os.Stderr.WriteString(clicky.MustFormat(tu.Pretty(), formatters.FormatOptions{Pretty: true}) + "\n")
		}
	}
}
