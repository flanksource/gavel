package verify

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/claudehistory"
)

type RunOptions struct {
	Config      VerifyConfig
	RepoPath    string
	Args        []string
	CommitRange string
}

func RunVerify(opts RunOptions) (*VerifyResult, error) {
	scope := ResolveScope(opts.Args, opts.CommitRange)
	tool, model := ResolveCLI(opts.Config.Model)

	logger.Infof("Verifying %s using %s", scope, model)

	prompt, err := renderPrompt(scope, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	schemaFile, err := SchemaFile(opts.Config.Checks)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema file: %w", err)
	}
	defer os.Remove(schemaFile)

	raw, err := Execute(tool, prompt, model, schemaFile, opts.RepoPath, logger.V(2).Enabled())
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %w", err)
	}

	if tool.Binary == "codex" {
		printCodexEvents(raw)
	}

	result, err := parseVerifyResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Score = ComputeOverallScore(result)
	return &result, nil
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
			switch event.Payload.Type {
			case "reasoning":
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
			switch event.Payload.Type {
			case "agent_reasoning":
				if event.Payload.Text != "" {
					tu = &claudehistory.ToolUse{Tool: "CodexReasoning", Input: map[string]any{"text": event.Payload.Text}}
				}
			}
		}
		if tu != nil {
			os.Stderr.WriteString(clicky.MustFormat(tu.Pretty(), formatters.FormatOptions{Pretty: true}) + "\n")
		}
	}
}

func ComputeOverallScore(r VerifyResult) int {
	var total, passed int
	for _, cr := range r.Checks {
		total++
		if cr.Pass {
			passed++
		}
	}

	checkScore := 0.0
	if total > 0 {
		checkScore = float64(passed) / float64(total) * 100
	}

	ratingSum := 0.0
	ratingCount := 0
	for _, rr := range r.Ratings {
		ratingSum += float64(rr.Score)
		ratingCount++
	}
	ratingScore := 0.0
	if ratingCount > 0 {
		ratingScore = ratingSum / float64(ratingCount)
	}

	completenessScore := 0.0
	if r.Completeness.Pass {
		completenessScore = 100
	}

	// Weighted: checks 50%, ratings 35%, completeness 15%
	combined := checkScore*0.50 + ratingScore*0.35 + completenessScore*0.15
	return int(math.Round(combined))
}
