package verify

import (
	"context"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
)

// executeAgentic runs the verification prompt through a captain agentic backend
// (claude-code / claude-agent). The agent fetches the diff itself via its tools
// and is instructed to emit JSON matching the output schema, which the caller
// parses.
func executeAgentic(cfg ai.AgentConfig, prompt, schema, repoPath string) (string, error) {
	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return "", fmt.Errorf("build verify agent: %w", err)
	}

	logger.V(1).Infof("verify: agentic review with %s in %s", cfg.Model, repoPath)
	req := captainai.Request{
		Prompt: api.Prompt{
			User:   prompt + "\n\n" + schemaInstruction(schema),
			Source: "verify",
		},
	}
	req.SetCwd(repoPath)
	resp, err := provider.Execute(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("verify agent execution failed: %w", err)
	}
	return resp.Text, nil
}

// schemaInstruction appends the JSON output-schema contract to the prompt. The
// agentic backends have no native structured-output mode, so the schema is
// carried in the prompt and the reply is parsed by parseVerifyResponse.
func schemaInstruction(schema string) string {
	var b strings.Builder
	b.WriteString("When you are done reviewing, respond with ONLY a single JSON object that")
	b.WriteString(" conforms to this JSON Schema — no prose, no markdown fences:\n")
	b.WriteString(schema)
	return b.String()
}

// parseVerifyResponse extracts a VerifyResult from an agent's raw reply: a bare
// JSON/YAML body, a JSON envelope carrying the text, fenced JSON, or JSON
// embedded in surrounding prose.
func parseVerifyResponse(raw string) (VerifyResult, error) {
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
