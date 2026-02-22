package verify

import (
	"encoding/json"
	"os"
	"strings"
)

type Claude struct{}

func (Claude) Name() string { return "claude" }

func (Claude) BuildVerifyArgs(prompt, model, schemaFile string, debug bool) []string {
	args := []string{"-p", prompt, "--output-format", "json"}
	if model != "" && model != "claude" {
		args = append(args, "--model", model)
	}
	if schemaFile != "" {
		if data, err := os.ReadFile(schemaFile); err == nil {
			args = append(args, "--json-schema", string(data))
		}
	}
	if debug {
		args = append(args, "--verbose")
	}
	return args
}

func (Claude) BuildFixArgs(model, prompt string, patchOnly bool) []string {
	if patchOnly {
		args := []string{"-p", prompt, "--output-format", "json"}
		if model != "" && model != "claude" {
			args = append(args, "--model", model)
		}
		return args
	}
	args := []string{"-p", prompt, "--allowedTools", "Edit,Write,Bash,Read,Glob,Grep"}
	if model != "" && model != "claude" {
		args = append(args, "--model", model)
	}
	return args
}

func (Claude) ParseResponse(raw string) (VerifyResult, error) {
	if result, ok := tryUnmarshalResult(raw); ok {
		return result, nil
	}

	text := extractTextFromJSON(raw)
	text = stripMarkdownFences(text)
	text = strings.TrimSpace(text)

	if result, ok := tryUnmarshalResult(text); ok {
		return result, nil
	}

	if embedded := extractJSONFromText(text); embedded != "" {
		if result, ok := tryUnmarshalResult(embedded); ok {
			return result, nil
		}
	}

	return VerifyResult{}, parseError(raw)
}

func (Claude) PostExecute(string) {}

func (Claude) ListModels() ([]string, error) {
	key := getEnv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, nil
	}
	return fetchModelIDs("https://api.anthropic.com/v1/models", "x-api-key", key, "2023-06-01")
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

func extractJSONFromText(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}
