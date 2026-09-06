package status

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
)

const maxAISummaryInputChars = 12000

//go:embed ai-file-summary.prompt
var fileSummaryPromptTemplate string

type fileSummarySchema struct {
	Summary string `json:"summary" description:"One-line plain-English summary of the file change, ideally under 12 words"`
}

var summarizeFileChangeWithAIFunc = summarizeFileChangeWithAI
var diffForStatusFileFunc = diffForStatusFile
var readUntrackedStatusFileFunc = readUntrackedStatusFile

func summarizeFileChangeWithAI(ctx context.Context, options SummaryOptions, file FileStatus) (summary string, err error) {
	if options.Prompt == nil {
		return "", fmt.Errorf("status summary prompt is required")
	}
	details, err := buildFileSummaryDetails(options.WorkDir, file)
	if err != nil {
		return "", err
	}

	runtime, err := options.Prompt.resolve(details)
	if err != nil {
		return "", fmt.Errorf("render AI file-summary prompt: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	for _, warning := range runtime.Resolution.Warnings {
		logger.Warnf("status summary %s preflight: %s", file.Path, warning)
	}
	agent, err := options.NewAgent(runtime)
	if err != nil {
		return "", fmt.Errorf("create AI file-summary agent: %w", err)
	}
	defer func() {
		if closeErr := agent.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close AI file-summary agent: %w", closeErr))
		}
	}()

	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name: fmt.Sprintf("status summary: %s", file.Path),
		Spec: runtime.Request,
	})
	if err != nil {
		return "", fmt.Errorf("execute AI file-summary prompt: %w", err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("AI file-summary prompt returned error: %s", resp.Error)
	}

	var schema fileSummarySchema
	if err := clickyai.DecodeStructured(resp, &schema); err != nil {
		return "", fmt.Errorf("decode AI file-summary response: %w", err)
	}
	summary = normalizeAISummary(schema.Summary)
	if summary == "" {
		return "", fmt.Errorf("AI file-summary prompt returned empty summary")
	}
	return summary, nil
}

func buildFileSummaryDetails(workDir string, file FileStatus) (string, error) {
	var sections []string

	stagedDiff, err := diffForStatusFileFunc(workDir, file.Path, true)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(stagedDiff) != "" {
		sections = append(sections, "Staged diff:\n"+truncateAISummaryInput(stagedDiff))
	}

	unstagedDiff, err := diffForStatusFileFunc(workDir, file.Path, false)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(unstagedDiff) != "" {
		sections = append(sections, "Unstaged diff:\n"+truncateAISummaryInput(unstagedDiff))
	}

	if file.State == StateUntracked {
		content, err := readUntrackedStatusFileFunc(workDir, file.Path)
		if err != nil {
			return "", err
		}
		if content != "" {
			sections = append(sections, "Untracked file contents:\n"+truncateAISummaryInput(content))
		}
	}

	if len(sections) == 0 {
		switch file.State {
		case StateConflict:
			sections = append(sections, "The file is currently in a merge conflict.")
		case StateUntracked:
			sections = append(sections, "New untracked file with no readable text content.")
		default:
			sections = append(sections, "No textual diff was available for this file.")
		}
	}

	return strings.Join(sections, "\n\n"), nil
}

func diffForStatusFile(workDir, path string, cached bool) (string, error) {
	args := []string{"diff", "--find-renames"}
	if cached {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff for %s (cached=%v): %w: %s", path, cached, err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func readUntrackedStatusFile(workDir, path string) (string, error) {
	data, err := os.ReadFile(filepath.Join(workDir, path))
	if err != nil {
		return "", fmt.Errorf("read untracked file %s: %w", path, err)
	}
	if !utf8.Valid(data) {
		return "Binary or non-UTF-8 file content omitted.", nil
	}
	return string(data), nil
}

func truncateAISummaryInput(s string) string {
	if len(s) <= maxAISummaryInputChars {
		return s
	}
	return s[:maxAISummaryInputChars] + "\n... (truncated)"
}

func normalizeAISummary(summary string) string {
	summary = strings.TrimSpace(summary)
	summary = strings.ReplaceAll(summary, "\n", " ")
	summary = strings.Join(strings.Fields(summary), " ")
	return summary
}
