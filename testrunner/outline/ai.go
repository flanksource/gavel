package outline

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
)

const (
	maxAISourceChars = 12000
	aiSummaryWorkers = 4
)

//go:embed test-summary.prompt
var testSummaryPromptTemplate string

type testSummaryItem struct {
	ID      string `json:"id" description:"the test id exactly as given"`
	Summary string `json:"summary" description:"one line, <=15 words, stating what behavior the test verifies"`
}

type fileSummariesSchema struct {
	Tests []testSummaryItem `json:"tests"`
}

// SummaryAgent is the slice of clickyai.Agent the outline needs; narrowed so
// tests can stub it.
type SummaryAgent interface {
	ExecutePrompt(ctx context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error)
}

// newSummaryAgent is swapped in tests.
var newSummaryAgent = func() (SummaryAgent, error) {
	return clickyai.NewAgent(clickyai.DefaultConfig())
}

// applyAISummaries generates one-line AI summaries for every leaf, batched
// per file to bound cost. Per-file failures are warned about and leave the
// static description in place; the outline itself still succeeds.
func applyAISummaries(ctx context.Context, report *Report, workDir string) error {
	leavesByFile := map[string][]*Entry{}
	for _, leaf := range report.Leaves() {
		leavesByFile[leaf.File] = append(leavesByFile[leaf.File], leaf)
	}
	if len(leavesByFile) == 0 {
		return nil
	}

	agent, err := newSummaryAgent()
	if err != nil {
		return fmt.Errorf("create AI agent for --ai-summary: %w", err)
	}

	template, err := resolveSummaryPrompt(workDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, aiSummaryWorkers)
	for file, leaves := range leavesByFile {
		wg.Add(1)
		go func(file string, leaves []*Entry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := summarizeFileTests(ctx, agent, workDir, file, leaves, template); err != nil {
				logger.Warnf("ai-summary for %s failed (keeping static descriptions): %v", file, err)
			}
		}(file, leaves)
	}
	wg.Wait()
	return nil
}

func summarizeFileTests(ctx context.Context, agent SummaryAgent, workDir, file string, leaves []*Entry, template string) error {
	source, err := os.ReadFile(filepath.Join(workDir, file))
	if err != nil {
		return fmt.Errorf("read test source: %w", err)
	}
	if strings.TrimSpace(template) == "" {
		template = testSummaryPromptTemplate
	}

	byID := map[string]*Entry{}
	ids := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		id := summaryID(leaf)
		byID[id] = leaf
		ids = append(ids, id)
	}

	req, _, err := prompt.Load(template).Render(map[string]any{
		"ids":    strings.Join(ids, "\n"),
		"file":   file,
		"source": truncateSource(string(source)),
	}, nil)
	if err != nil {
		return fmt.Errorf("render summary prompt: %w", err)
	}

	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:       fmt.Sprintf("test outline summary: %s", file),
		Prompt:     req.Prompt.User,
		SchemaJSON: req.Prompt.SchemaJSON,
		Source:     "test-summary.prompt",
	})
	if err != nil {
		return fmt.Errorf("execute summary prompt: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("summary prompt returned error: %s", resp.Error)
	}

	var schema fileSummariesSchema
	if err := clickyai.DecodeStructured(resp, &schema); err != nil {
		return fmt.Errorf("decode summary response: %w", err)
	}

	matched := 0
	for _, item := range schema.Tests {
		leaf := byID[strings.TrimSpace(item.ID)]
		summary := strings.Join(strings.Fields(item.Summary), " ")
		if leaf == nil || summary == "" {
			continue
		}
		leaf.AISummary = summary
		matched++
	}
	if matched == 0 {
		return fmt.Errorf("no returned summaries matched the %d requested test ids", len(ids))
	}
	if matched < len(ids) {
		logger.Warnf("ai-summary for %s matched %d of %d tests", file, matched, len(ids))
	}
	return nil
}

func summaryID(leaf *Entry) string {
	name := strings.Join(append(append([]string(nil), leaf.Suite...), leaf.Name), " > ")
	return fmt.Sprintf("%s:%d %s", leaf.File, leaf.Line, name)
}

func truncateSource(s string) string {
	if len(s) <= maxAISourceChars {
		return s
	}
	return s[:maxAISourceChars] + "\n... (truncated)"
}
