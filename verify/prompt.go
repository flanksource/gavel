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
	data := map[string]any{
		"scope":        scope,
		"sections":     cfg.Sections,
		"extra_prompt": cfg.Prompt,
	}
	return gomplate.RunTemplate(data, gomplate.Template{
		Template: verifyPromptTemplate,
	})
}

func parseVerifyResponse(raw string) (VerifyResult, error) {
	text := extractTextFromJSON(raw)
	text = stripYAMLFences(text)
	text = strings.TrimSpace(text)

	var result VerifyResult
	if err := yaml.Unmarshal([]byte(text), &result); err != nil {
		if block := extractYAMLBlock(text); block != "" {
			if err2 := yaml.Unmarshal([]byte(block), &result); err2 != nil {
				return result, fmt.Errorf("failed to parse YAML response: %w (original: %v)", err2, err)
			}
			return result, nil
		}
		return result, fmt.Errorf("failed to parse YAML response: %w", err)
	}
	return result, nil
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

func stripYAMLFences(s string) string {
	s = strings.TrimPrefix(s, "```yaml\n")
	s = strings.TrimPrefix(s, "```yaml")
	s = strings.TrimPrefix(s, "```yml\n")
	s = strings.TrimPrefix(s, "```yml")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimPrefix(s, "```")
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
