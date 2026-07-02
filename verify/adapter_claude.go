package verify

import (
	"os"
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

func (Claude) ParseResponse(raw string) (VerifyResult, error) {
	return parseVerifyResponse(raw)
}

func (Claude) PostExecute(string) {}

func (Claude) ListModels() ([]string, error) {
	key := getEnv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, nil
	}
	return fetchModelIDs("https://api.anthropic.com/v1/models", "x-api-key", key, "2023-06-01")
}
