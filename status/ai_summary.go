package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	clickyai "github.com/flanksource/clicky/ai"
)

const (
	maxAISummaryInputChars = 12000
	fileSummaryPrompt      = `You are summarizing changes to a single file from git diff context.

Return a JSON object with:
- summary: a single-line plain-English summary of what changed in this file

Rules:
- Keep it to one short sentence fragment, ideally under 12 words.
- Describe the substance of the change, not git mechanics.
- Mention the file path only when it adds necessary context.
- Do not mention line counts unless they are the point of the change.
- The input will be git diff text for tracked files, or file contents for new untracked files.

Change context:
%s
`
)

type fileSummarySchema struct {
	Summary string `json:"summary" description:"One-line plain-English summary of the file change, ideally under 12 words"`
}

var summarizeFileChangeWithAIFunc = summarizeFileChangeWithAI
var diffForStatusFileFunc = diffForStatusFile
var readUntrackedStatusFileFunc = readUntrackedStatusFile

func summarizeFileChangeWithAI(ctx context.Context, workDir string, agent clickyai.Agent, file FileStatus) (string, error) {
	details, err := buildFileSummaryDetails(workDir, file)
	if err != nil {
		return "", err
	}

	schema := &fileSummarySchema{}
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             fmt.Sprintf("status summary: %s", file.Path),
		Prompt:           fmt.Sprintf(fileSummaryPrompt, details),
		StructuredOutput: schema,
	})
	if err != nil {
		return "", fmt.Errorf("execute AI file-summary prompt: %w", err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("AI file-summary prompt returned error: %s", resp.Error)
	}

	summary := parseAISummary(schema.Summary, resp.Result)
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

func parseAISummary(schemaSummary, raw string) string {
	if summary := normalizeAISummary(schemaSummary); summary != "" {
		return summary
	}

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"`)
	return normalizeAISummary(raw)
}

func normalizeAISummary(summary string) string {
	summary = strings.TrimSpace(summary)
	summary = strings.ReplaceAll(summary, "\n", " ")
	summary = strings.Join(strings.Fields(summary), " ")
	return summary
}
