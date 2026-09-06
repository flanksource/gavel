package outline

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	captaincli "github.com/flanksource/captain/pkg/cli"
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
	Close() error
}

var newSummaryAgent = func(runtime captaincli.AIRuntimeResolved) (SummaryAgent, error) {
	return clickyai.NewAgent(runtime.Config)
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

	prompt, err := resolveSummaryPrompt(workDir)
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
			if err := summarizeFileTests(ctx, prompt, file, leaves); err != nil {
				logger.Warnf("ai-summary for %s failed (keeping static descriptions): %v", file, err)
			}
		}(file, leaves)
	}
	wg.Wait()
	return nil
}

func summarizeFileTests(ctx context.Context, prompt *summaryPrompt, file string, leaves []*Entry) (err error) {
	source, err := os.ReadFile(filepath.Join(prompt.dir, file))
	if err != nil {
		return fmt.Errorf("read test source: %w", err)
	}
	byID := map[string]*Entry{}
	ids := make([]string, 0, len(leaves))
	for _, leaf := range leaves {
		id := summaryID(leaf)
		byID[id] = leaf
		ids = append(ids, id)
	}

	runtime, err := prompt.resolve(map[string]any{
		"ids":    strings.Join(ids, "\n"),
		"file":   file,
		"source": truncateSource(string(source)),
	})
	if err != nil {
		return fmt.Errorf("render summary prompt: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, warning := range runtime.Resolution.Warnings {
		logger.Warnf("test outline summary %s preflight: %s", file, warning)
	}
	agent, err := newSummaryAgent(runtime)
	if err != nil {
		return fmt.Errorf("create AI agent for --ai-summary: %w", err)
	}
	defer func() {
		if closeErr := agent.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close AI summary agent: %w", closeErr))
		}
	}()

	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name: fmt.Sprintf("test outline summary: %s", file),
		Spec: runtime.Request,
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
