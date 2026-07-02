package ai

// Structured output for agentic backends. Captain's agentic providers
// (claude-code, claude-agent, cmux, codex) reject Prompt.Schema, so the schema
// travels IN the prompt (SchemaInstruction) and the final reply is parsed back
// out tolerantly (ParseStructured). This is deliberately the only schema
// plumbing gavel owns; if captain grows native schema-in-prompt support this
// file collapses onto it.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/ghodss/yaml"
)

// SchemaInstruction is the trailing prompt block demanding a bare-JSON reply
// conforming to schemaJSON.
func SchemaInstruction(schemaJSON string) string {
	var b strings.Builder
	b.WriteString("When you are done, respond with ONLY a single JSON object that")
	b.WriteString(" conforms to this JSON Schema — no prose, no markdown fences:\n")
	b.WriteString(schemaJSON)
	return b.String()
}

// StructuredRequest describes one schema-in-prompt agent execution.
type StructuredRequest struct {
	Config     AgentConfig
	Prompt     string // task prompt; SchemaInstruction(SchemaJSON) is appended
	SchemaJSON string
	RepoPath   string // agent working directory
	Source     string // request attribution for logging (e.g. "verify")
}

// ExecuteStructured runs req through a captain agentic backend and parses the
// reply into a T. validate decides whether a decoded candidate is acceptable —
// tolerant YAML decoding accepts nearly anything, so every caller must check
// its required fields there.
func ExecuteStructured[T any](ctx context.Context, req StructuredRequest, validate func(*T) error) (*T, error) {
	provider, err := NewProvider(req.Config)
	if err != nil {
		return nil, fmt.Errorf("build %s agent: %w", req.Source, err)
	}

	logger.V(1).Infof("%s: structured agent run with %s in %s", req.Source, req.Config.Model, req.RepoPath)
	creq := captainai.Request{
		Prompt: api.Prompt{
			User:   req.Prompt + "\n\n" + SchemaInstruction(req.SchemaJSON),
			Source: req.Source,
		},
	}
	creq.SetCwd(req.RepoPath)
	resp, err := provider.Execute(ctx, creq)
	if err != nil {
		return nil, fmt.Errorf("%s agent execution failed: %w", req.Source, err)
	}
	return ParseStructured(resp.Text, validate)
}

// ParseStructured extracts a T from an agent's raw reply: a bare JSON/YAML
// body, a JSON envelope carrying the text (result/text/response/content keys),
// fenced JSON, a `---` YAML document, or JSON embedded in surrounding prose.
// A candidate only counts when validate accepts it; the last validation
// failure is surfaced so a decoded-but-wrong payload is diagnosable.
func ParseStructured[T any](raw string, validate func(*T) error) (*T, error) {
	p := structuredParser[T]{validate: validate}
	if v := p.try(raw); v != nil {
		return v, nil
	}

	text := extractTextFromJSON(raw)
	text = stripMarkdownFences(text)
	text = strings.TrimSpace(text)
	if v := p.try(text); v != nil {
		return v, nil
	}

	if embedded := extractJSONFromText(text); embedded != "" {
		if v := p.try(embedded); v != nil {
			return v, nil
		}
	}

	preview := raw
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	if p.lastValidateErr != nil {
		return nil, fmt.Errorf("reply decoded but failed validation: %w (preview: %s)", p.lastValidateErr, preview)
	}
	return nil, fmt.Errorf("failed to parse structured reply (preview: %s)", preview)
}

type structuredParser[T any] struct {
	validate        func(*T) error
	lastValidateErr error
}

// try decodes text as JSON, then YAML, then a `---` YAML document, returning
// the first candidate validate accepts.
func (p *structuredParser[T]) try(text string) *T {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if v := p.decode(text, json.Unmarshal); v != nil {
		return v
	}
	if v := p.decode(text, yaml.Unmarshal); v != nil {
		return v
	}
	if block := extractYAMLBlock(text); block != "" {
		if v := p.decode(block, yaml.Unmarshal); v != nil {
			return v
		}
	}
	return nil
}

func (p *structuredParser[T]) decode(text string, unmarshal func([]byte, any) error) *T {
	var v T
	if err := unmarshal([]byte(text), &v); err != nil {
		return nil
	}
	if err := p.validate(&v); err != nil {
		p.lastValidateErr = err
		return nil
	}
	return &v
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

// extractTextFromJSON unwraps a JSON envelope ({"result": "..."} and friends)
// down to the text it carries; anything else passes through unchanged.
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

// extractJSONFromText returns the first brace-balanced JSON object embedded in
// prose, or "" when none closes.
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
