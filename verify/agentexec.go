package verify

import (
	"context"
	"fmt"
	"os"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
)

// mockVerifyResult is returned by executeAgentic when MOCK is active and no
// MOCK_VERIFY_JSON override is set. It parses to a passing VerifyResult (one
// check, full ratings, complete, implemented) so MOCK runs exercise the full
// orchestration without calling a real agent.
const mockVerifyResult = `{"checks":{"definition-of-done":{"pass":true}},` +
	`"ratings":{"duplication":{"score":100},"consistency":{"score":100},"security":{"score":100},"coverage":{"score":100}},` +
	`"completeness":{"pass":true,"summary":"mock"},"implemented":true}`

// executeAgentic runs the verification prompt through a captain agentic backend
// (claude-code / claude-agent). The agent fetches the diff itself via its tools
// and is instructed to emit JSON matching the output schema, which the caller
// parses. The MOCK env short-circuits to a deterministic reply (override with
// MOCK_VERIFY_JSON; set MOCK=false to hit the real agent).
func executeAgentic(cfg ai.AgentConfig, prompt, schema, repoPath string) (string, error) {
	if os.Getenv("MOCK") != "false" {
		if override := os.Getenv("MOCK_VERIFY_JSON"); override != "" {
			return override, nil
		}
		return mockVerifyResult, nil
	}

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		return "", fmt.Errorf("build verify agent: %w", err)
	}

	logger.V(1).Infof("verify: agentic review with %s in %s", cfg.Model, repoPath)
	resp, err := provider.Execute(context.Background(), captainai.Request{
		Prompt: api.Prompt{
			User:   prompt + "\n\n" + schemaInstruction(schema),
			Source: "verify",
		},
		Context: api.Context{Dir: repoPath},
	})
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
