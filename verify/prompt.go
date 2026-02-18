package verify

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gomplate/v3"
	"github.com/ghodss/yaml"
)

//go:embed verify-prompt.md
var verifyPromptTemplate string

func renderPrompt(scope ReviewScope, cfg VerifyConfig) (string, error) {
	checks := EnabledChecks(cfg.Checks)
	byCategory := ChecksByCategory(checks)

	data := map[string]any{
		"scope":        scope,
		"extra_prompt": cfg.Prompt,
		"categories":   byCategory,
		"catOrder":     AllCategories,
		"ratings":      RatingDimensions,
	}
	return gomplate.RunTemplate(data, gomplate.Template{
		Template: verifyPromptTemplate,
	})
}

func parseVerifyResponse(raw string) (VerifyResult, error) {
	var result VerifyResult

	// Try direct JSON parse
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &result); err == nil && len(result.Checks) > 0 {
		return result, nil
	}

	// Try extracting from JSONL (codex output format)
	if text := extractFromJSONL(raw); text != "" {
		if err := json.Unmarshal([]byte(text), &result); err == nil && len(result.Checks) > 0 {
			return result, nil
		}
	}

	text := extractTextFromJSON(raw)
	text = stripMarkdownFences(text)
	text = strings.TrimSpace(text)

	if err := json.Unmarshal([]byte(text), &result); err == nil && len(result.Checks) > 0 {
		return result, nil
	}

	if err := yaml.Unmarshal([]byte(text), &result); err != nil {
		if block := extractYAMLBlock(text); block != "" {
			if err2 := yaml.Unmarshal([]byte(block), &result); err2 != nil {
				return result, fmt.Errorf("failed to parse response: %w (original: %v)", err2, err)
			}
			return result, nil
		}
		return result, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
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

func extractTextFromJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		return raw
	}
	var wrapper map[string]any
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return raw
	}
	for _, key := range []string{"result", "text", "response", "content"} {
		if v, ok := wrapper[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return raw
}

func stripMarkdownFences(s string) string {
	for _, prefix := range []string{"```json\n", "```json", "```yaml\n", "```yaml", "```yml\n", "```yml", "```\n", "```"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return s
}

func extractYAMLBlock(s string) string {
	parts := strings.Split(s, "---")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
