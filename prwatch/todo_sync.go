package prwatch

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"gopkg.in/yaml.v3"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func todoFilename(workflowName, jobName string) string {
	combined := strings.ToLower(workflowName + "-" + jobName)
	cleaned := nonAlphanumeric.ReplaceAllString(combined, "-")
	return strings.Trim(cleaned, "-") + ".md"
}

func todoPath(todosDir string, prNumber int, workflowName, jobName string) string {
	return filepath.Join(todosDir, strconv.Itoa(prNumber), todoFilename(workflowName, jobName))
}

func SyncTodos(result *PRWatchResult, todosDir string) error {
	if result == nil || result.PR == nil {
		return nil
	}

	for _, run := range result.Runs {
		for _, job := range run.Jobs {
			if !strings.EqualFold(job.Status, "completed") {
				continue
			}
			path := todoPath(todosDir, result.PR.Number, run.Name, job.Name)

			if strings.EqualFold(job.Conclusion, "failure") {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					logger.Infof("creating todo for failed job %s/%s: %s", run.Name, job.Name, path)
					if err := createJobTodo(path, run, job, result.PR); err != nil {
						return fmt.Errorf("create todo %s: %w", path, err)
					}
				} else {
					logger.Infof("updating todo for failed job %s/%s: %s", run.Name, job.Name, path)
					if err := updateJobTodo(path, run, job); err != nil {
						return fmt.Errorf("update todo %s: %w", path, err)
					}
				}
			} else if strings.EqualFold(job.Conclusion, "success") {
				if _, err := os.Stat(path); err == nil {
					logger.Infof("completing todo for passing job %s/%s: %s", run.Name, job.Name, path)
					if err := completeJobTodo(path); err != nil {
						return fmt.Errorf("complete todo %s: %w", path, err)
					}
				}
			}
		}
	}
	return nil
}

func createJobTodo(path string, run *github.WorkflowRun, job github.Job, pr *github.PRInfo) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	now := time.Now()
	fm := types.TODOFrontmatter{
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
		Attempts: 1,
		LastRun:  &now,
	}
	fm.Build = fmt.Sprintf("git fetch origin && git checkout %s", pr.HeadRefName)

	body := formatJobBody(run, job, pr)
	content, err := todos.WriteFrontmatter(&fm, body)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func updateJobTodo(path string, _ *github.WorkflowRun, job github.Job) error {
	parsed, err := todos.ParseFrontmatterFromFile(path)
	if err != nil {
		return err
	}

	parsed.Frontmatter.Attempts++
	now := time.Now()
	parsed.Frontmatter.LastRun = &now

	if parsed.Frontmatter.Status == types.StatusCompleted {
		parsed.Frontmatter.Status = types.StatusPending
	}

	logs := formatJobLogs(job)
	history := fmt.Sprintf("\n\n## Attempt %d\n\n%s", parsed.Frontmatter.Attempts, logs)
	markdown := parsed.MarkdownContent + history

	content, err := todos.WriteFrontmatter(&parsed.Frontmatter, markdown)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func completeJobTodo(path string) error {
	parsed, err := todos.ParseFrontmatterFromFile(path)
	if err != nil {
		return err
	}

	if parsed.Frontmatter.Status == types.StatusCompleted {
		return nil
	}

	parsed.Frontmatter.Status = types.StatusCompleted
	now := time.Now()
	parsed.Frontmatter.LastRun = &now

	content, err := todos.WriteFrontmatter(&parsed.Frontmatter, parsed.MarkdownContent)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func formatJobBody(run *github.WorkflowRun, job github.Job, pr *github.PRInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n# %s / %s\n\n", run.Name, job.Name))
	sb.WriteString(fmt.Sprintf("PR #%d (`%s`)\n", pr.Number, pr.HeadRefName))
	if job.URL != "" {
		sb.WriteString(fmt.Sprintf("Job: %s\n", job.URL))
	}
	if run.WorkflowYAML != "" {
		if steps := extractJobSteps(run.WorkflowYAML, job.Name); steps != "" {
			sb.WriteString("\n## Workflow Steps\n\n")
			sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n", strings.TrimSpace(steps)))
		} else {
			sb.WriteString("\n## Workflow Definition\n\n")
			sb.WriteString(fmt.Sprintf("```yaml\n%s\n```\n", strings.TrimSpace(run.WorkflowYAML)))
		}
	}
	sb.WriteString("\n## Logs\n\n")
	sb.WriteString(formatJobLogs(job))
	return sb.String()
}

func extractJobSteps(workflowYAML, jobName string) string {
	var wf map[string]any
	if err := yaml.Unmarshal([]byte(workflowYAML), &wf); err != nil {
		return ""
	}
	jobsRaw, ok := wf["jobs"]
	if !ok {
		return ""
	}
	jobs, ok := jobsRaw.(map[string]any)
	if !ok {
		return ""
	}

	for key, val := range jobs {
		jobMap, ok := val.(map[string]any)
		if !ok {
			continue
		}
		name, _ := jobMap["name"].(string)
		if !matchJobName(name, jobName) && !strings.EqualFold(key, jobName) {
			continue
		}
		steps, ok := jobMap["steps"]
		if !ok {
			return ""
		}
		out, err := yaml.Marshal(steps)
		if err != nil {
			return ""
		}
		return string(out)
	}
	return ""
}

var matrixPlaceholder = regexp.MustCompile(`\$\{\{[^}]*\}\}`)

func matchJobName(yamlName, apiName string) bool {
	if strings.EqualFold(yamlName, apiName) {
		return true
	}
	if !strings.Contains(yamlName, "${{") {
		return false
	}
	// Split on placeholders, quote each literal segment, rejoin with .*
	parts := matrixPlaceholder.Split(yamlName, -1)
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	re, err := regexp.Compile("(?i)^" + strings.Join(parts, ".*") + "$")
	if err != nil {
		return false
	}
	return re.MatchString(apiName)
}

var detailsBlockRegex = regexp.MustCompile(`(?s)<details>\s*<summary>([^<]*)</summary>(.*?)</details>`)

var leadingNonASCII = regexp.MustCompile(`^[^\x00-\x7F\s]+\s*`)

type detailsBlock struct {
	Summary string // full original summary e.g. "♻️ Suggested fix — capture and reuse the generated ID"
	Body    string
}

func parseDetailsBlocks(body, summaryPrefix string) []detailsBlock {
	matches := detailsBlockRegex.FindAllStringSubmatch(body, -1)
	var results []detailsBlock
	for _, m := range matches {
		fullSummary := strings.TrimSpace(m[1])
		stripped := leadingNonASCII.ReplaceAllString(fullSummary, "")
		if strings.HasPrefix(stripped, summaryPrefix) {
			results = append(results, detailsBlock{
				Summary: fullSummary,
				Body:    strings.TrimSpace(m[2]),
			})
		}
	}
	return results
}

var htmlCommentRegex = regexp.MustCompile(`(?s)<!--.*?-->`)
var h1Regex = regexp.MustCompile(`(?m)^#\s+(.+)$`)
var boldRegex = regexp.MustCompile(`\*\*(.+?)\*\*`)

func extractTitle(body string) string {
	if m := h1Regex.FindStringSubmatch(body); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := boldRegex.FindStringSubmatch(body); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractNonDetailsText(body string) string {
	// Remove HTML comments
	text := htmlCommentRegex.ReplaceAllString(body, "")
	// Remove all <details>...</details> blocks
	text = detailsBlockRegex.ReplaceAllString(text, "")
	// Collapse 3+ consecutive newlines into 2
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(text)
}

func SyncCommentTodos(comments []github.PRComment, pr *github.PRInfo, todosDir string) error {
	if pr == nil {
		return nil
	}
	for _, comment := range comments {
		promptBlocks := parseDetailsBlocks(comment.Body, "Fix all issues with AI agents")
		suggestedFixBlocks := parseDetailsBlocks(comment.Body, "Suggested fix")
		if len(promptBlocks) == 0 && len(suggestedFixBlocks) == 0 {
			continue
		}
		path := filepath.Join(todosDir, strconv.Itoa(pr.Number), fmt.Sprintf("code-review-%d.md", comment.ID))
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
		}

		now := time.Now()
		fm := types.TODOFrontmatter{
			Priority: types.PriorityMedium,
			Status:   types.StatusPending,
			Attempts: 0,
			LastRun:  &now,
		}
		fm.Build = fmt.Sprintf("git fetch origin && git checkout %s", pr.HeadRefName)
		if title := extractTitle(comment.Body); title != "" {
			fm.Title = title
		}

		var sb strings.Builder
		sb.WriteString("\n# Code Review Comment\n\n")
		sb.WriteString(fmt.Sprintf("PR #%d (`%s`)\n", pr.Number, pr.HeadRefName))
		sb.WriteString(fmt.Sprintf("Author: %s\n", comment.Author))
		if comment.URL != "" {
			sb.WriteString(fmt.Sprintf("Comment: %s\n", comment.URL))
		}
		if comment.Path != "" {
			if comment.Line > 0 {
				sb.WriteString(fmt.Sprintf("File: `%s:%d`\n", comment.Path, comment.Line))
			} else {
				sb.WriteString(fmt.Sprintf("File: `%s`\n", comment.Path))
			}
		}

		if desc := extractNonDetailsText(comment.Body); desc != "" {
			sb.WriteString("\n")
			sb.WriteString(desc)
			sb.WriteString("\n")
		}

		for _, block := range promptBlocks {
			sb.WriteString("\n")
			sb.WriteString(block.Body)
			sb.WriteString("\n")
		}

		for _, block := range suggestedFixBlocks {
			sb.WriteString(fmt.Sprintf("\n## %s\n\n", block.Summary))
			sb.WriteString(block.Body)
			sb.WriteString("\n")
		}

		content, err := todos.WriteFrontmatter(&fm, sb.String())
		if err != nil {
			return fmt.Errorf("write frontmatter for comment %d: %w", comment.ID, err)
		}
		logger.Infof("creating todo for code review comment %d: %s", comment.ID, path)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write todo %s: %w", path, err)
		}
	}
	return nil
}

func formatJobLogs(job github.Job) string {
	var sb strings.Builder
	hasStepLogs := false
	for _, step := range job.Steps {
		if !strings.EqualFold(step.Conclusion, "failure") || step.Logs == "" {
			continue
		}
		hasStepLogs = true
		sb.WriteString(fmt.Sprintf("### Step: %s\n\n```\n%s\n```\n\n", step.Name, strings.TrimSpace(step.Logs)))
	}
	if !hasStepLogs && job.Logs != "" {
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", strings.TrimSpace(job.Logs)))
	}
	return sb.String()
}
