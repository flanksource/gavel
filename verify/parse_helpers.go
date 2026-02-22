package verify

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
)

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

func tryUnmarshalResult(text string) (VerifyResult, bool) {
	text = strings.TrimSpace(text)
	var result VerifyResult
	if err := json.Unmarshal([]byte(text), &result); err == nil && len(result.Checks) > 0 {
		return result, true
	}
	if err := yaml.Unmarshal([]byte(text), &result); err == nil && len(result.Checks) > 0 {
		return result, true
	}
	if block := extractYAMLBlock(text); block != "" {
		if err := yaml.Unmarshal([]byte(block), &result); err == nil && len(result.Checks) > 0 {
			return result, true
		}
	}
	return VerifyResult{}, false
}

func parseError(raw string) error {
	preview := raw
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	return fmt.Errorf("failed to parse response (preview: %s)", preview)
}
