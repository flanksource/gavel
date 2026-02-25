package prwatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func failedResult(prNum int, workflow, jobName, stepName, logs string) *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{Number: prNum, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: workflow, Status: "completed", Conclusion: "failure",
				Jobs: []github.Job{{
					Name: jobName, Status: "completed", Conclusion: "failure",
					Steps: []github.Step{{Name: stepName, Status: "completed", Conclusion: "failure", Logs: logs}},
				}},
			},
		},
	}
}

func successResult(prNum int, workflow, jobName string) *PRWatchResult {
	return &PRWatchResult{
		PR: &github.PRInfo{Number: prNum, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: workflow, Status: "completed", Conclusion: "success",
				Jobs: []github.Job{{Name: jobName, Status: "completed", Conclusion: "success"}},
			},
		},
	}
}

func TestSyncTodosCreateOnFailure(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL: TestFoo")

	require.NoError(t, SyncTodos(result, dir))

	path := filepath.Join(dir, "99", "ci-unit-tests.md")
	require.FileExists(t, path)

	parsed, err := todos.ParseFrontmatterFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, types.PriorityHigh, parsed.Frontmatter.Priority)
	assert.Equal(t, types.StatusPending, parsed.Frontmatter.Status)
	assert.Equal(t, 1, parsed.Frontmatter.Attempts)
	assert.Equal(t, "git fetch origin && git checkout feat/x", parsed.Frontmatter.Build)
	assert.Contains(t, parsed.MarkdownContent, "CI / unit-tests")
	assert.Contains(t, parsed.MarkdownContent, "FAIL: TestFoo")

	require.NotNil(t, parsed.Frontmatter.PR)
	assert.Equal(t, 99, parsed.Frontmatter.PR.Number)
	assert.Equal(t, "feat/x", parsed.Frontmatter.PR.Head)
	assert.Equal(t, "main", parsed.Frontmatter.PR.Base)
}

func TestSyncTodosCreateWithWorkflowYAML(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL: TestFoo")
	result.Runs[1].WorkflowYAML = `name: CI
on: [push]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make lint
  unit-tests:
    name: unit-tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...
  deploy:
    runs-on: ubuntu-latest
    steps:
      - run: make deploy`

	require.NoError(t, SyncTodos(result, dir))

	path := filepath.Join(dir, "99", "ci-unit-tests.md")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(raw)
	assert.Contains(t, content, "## Workflow Steps")
	assert.Contains(t, content, "go test ./...")
	assert.NotContains(t, content, "make lint", "should not include lint job steps")
	assert.NotContains(t, content, "make deploy", "should not include deploy job steps")
}

func TestSyncTodosUpdateOnRepeatedFailure(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL: TestFoo attempt1")
	require.NoError(t, SyncTodos(result, dir))

	result2 := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL: TestFoo attempt2")
	require.NoError(t, SyncTodos(result2, dir))

	parsed, err := todos.ParseFrontmatterFromFile(filepath.Join(dir, "99", "ci-unit-tests.md"))
	require.NoError(t, err)
	assert.Equal(t, 2, parsed.Frontmatter.Attempts)
	assert.Contains(t, parsed.MarkdownContent, "Attempt 2")
	assert.Contains(t, parsed.MarkdownContent, "FAIL: TestFoo attempt2")
}

func TestSyncTodosCompleteOnSuccess(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL")
	require.NoError(t, SyncTodos(result, dir))

	success := successResult(99, "CI", "unit-tests")
	require.NoError(t, SyncTodos(success, dir))

	parsed, err := todos.ParseFrontmatterFromFile(filepath.Join(dir, "99", "ci-unit-tests.md"))
	require.NoError(t, err)
	assert.Equal(t, types.StatusCompleted, parsed.Frontmatter.Status)
}

func TestSyncTodosNoopForAlreadyCompleted(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL")
	require.NoError(t, SyncTodos(result, dir))

	success := successResult(99, "CI", "unit-tests")
	require.NoError(t, SyncTodos(success, dir))

	path := filepath.Join(dir, "99", "ci-unit-tests.md")
	before, _ := os.ReadFile(path)

	require.NoError(t, SyncTodos(success, dir))
	after, _ := os.ReadFile(path)

	assert.Equal(t, string(before), string(after))
}

func TestSyncTodosReopenOnRegression(t *testing.T) {
	dir := t.TempDir()
	result := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL v1")
	require.NoError(t, SyncTodos(result, dir))

	success := successResult(99, "CI", "unit-tests")
	require.NoError(t, SyncTodos(success, dir))

	regression := failedResult(99, "CI", "unit-tests", "Run tests", "FAIL regression")
	require.NoError(t, SyncTodos(regression, dir))

	parsed, err := todos.ParseFrontmatterFromFile(filepath.Join(dir, "99", "ci-unit-tests.md"))
	require.NoError(t, err)
	assert.Equal(t, types.StatusPending, parsed.Frontmatter.Status)
	assert.Equal(t, 2, parsed.Frontmatter.Attempts)
	assert.Contains(t, parsed.MarkdownContent, "FAIL regression")
}

func TestExtractJobSteps(t *testing.T) {
	workflow := `name: CI
on: [push]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make lint
  test:
    name: unit-tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...`

	t.Run("match by name property", func(t *testing.T) {
		steps := extractJobSteps(workflow, "unit-tests")
		assert.Contains(t, steps, "go test")
		assert.NotContains(t, steps, "make lint")
	})

	t.Run("match by key", func(t *testing.T) {
		steps := extractJobSteps(workflow, "lint")
		assert.Contains(t, steps, "make lint")
		assert.NotContains(t, steps, "go test")
	})

	t.Run("no match returns empty", func(t *testing.T) {
		assert.Empty(t, extractJobSteps(workflow, "deploy"))
	})

	t.Run("invalid YAML returns empty", func(t *testing.T) {
		assert.Empty(t, extractJobSteps("{{invalid", "test"))
	})

	t.Run("matrix name matching", func(t *testing.T) {
		matrixWorkflow := `name: Test
on: [push]
jobs:
  test:
    name: "Test (${{ matrix.os }}, Go ${{ matrix.go-version }})"
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [ubuntu-latest]
        go-version: ["1.25"]
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...`
		steps := extractJobSteps(matrixWorkflow, "Test (ubuntu-latest, Go 1.25)")
		assert.Contains(t, steps, "go test", "should match matrix-expanded job name")
	})
}

func TestTodoFilename(t *testing.T) {
	tests := []struct {
		workflow, job, expected string
	}{
		{"CI", "unit-tests", "ci-unit-tests.md"},
		{"Build & Deploy", "lint check", "build-deploy-lint-check.md"},
		{"  Spaces  ", "  around  ", "spaces-around.md"},
		{"CI/CD", "go-test (ubuntu)", "ci-cd-go-test-ubuntu.md"},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.expected, todoFilename(tc.workflow, tc.job), "%s + %s", tc.workflow, tc.job)
	}
}

func TestSyncTodosSkipsInProgressJobs(t *testing.T) {
	dir := t.TempDir()
	result := &PRWatchResult{
		PR: &github.PRInfo{Number: 99, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: "CI", Status: "in_progress",
				Jobs: []github.Job{{Name: "unit-tests", Status: "in_progress", Conclusion: ""}},
			},
		},
	}
	require.NoError(t, SyncTodos(result, dir))

	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries)
}

func TestFormatJobLogs(t *testing.T) {
	t.Run("step logs", func(t *testing.T) {
		job := github.Job{
			Steps: []github.Step{
				{Name: "Run tests", Conclusion: "failure", Logs: "error line"},
				{Name: "Setup", Conclusion: "success", Logs: "ok"},
			},
		}
		out := formatJobLogs(job)
		assert.Contains(t, out, "### Step: Run tests")
		assert.Contains(t, out, "error line")
		assert.NotContains(t, out, "Setup")
	})

	t.Run("fallback to job logs", func(t *testing.T) {
		job := github.Job{
			Logs:  "job level error",
			Steps: []github.Step{{Name: "Run tests", Conclusion: "failure"}},
		}
		out := formatJobLogs(job)
		assert.Contains(t, out, "job level error")
	})
}

func TestFormatJobBody(t *testing.T) {
	t.Run("without workflow YAML", func(t *testing.T) {
		run := &github.WorkflowRun{Name: "CI"}
		job := github.Job{Name: "lint", URL: "https://github.com/org/repo/actions/runs/1/job/2"}
		pr := &github.PRInfo{Number: 42, HeadRefName: "fix/lint"}

		body := formatJobBody(run, job, pr)
		assert.True(t, strings.HasPrefix(body, "\n# CI / lint"))
		assert.Contains(t, body, "PR #42")
		assert.Contains(t, body, "fix/lint")
		assert.Contains(t, body, job.URL)
		assert.NotContains(t, body, "## Workflow Definition")
	})

	t.Run("with workflow YAML extracts job steps", func(t *testing.T) {
		run := &github.WorkflowRun{Name: "CI", WorkflowYAML: `name: CI
on: [push]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make lint
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make build`}
		job := github.Job{Name: "lint"}
		pr := &github.PRInfo{Number: 42, HeadRefName: "fix/lint"}

		body := formatJobBody(run, job, pr)
		assert.Contains(t, body, "## Workflow Steps")
		assert.Contains(t, body, "make lint")
		assert.NotContains(t, body, "make build", "should not include other job steps")
	})

	t.Run("with workflow YAML falls back when job not found", func(t *testing.T) {
		run := &github.WorkflowRun{Name: "CI", WorkflowYAML: "name: CI\non: [push]\njobs:\n  build:\n    steps:\n      - run: make build"}
		job := github.Job{Name: "nonexistent-job"}
		pr := &github.PRInfo{Number: 42, HeadRefName: "fix/lint"}

		body := formatJobBody(run, job, pr)
		assert.Contains(t, body, "## Workflow Definition")
		assert.Contains(t, body, "make build")
	})

	t.Run("workflow steps appear before logs", func(t *testing.T) {
		run := &github.WorkflowRun{Name: "CI", WorkflowYAML: `name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./...`}
		job := github.Job{Name: "test", Steps: []github.Step{
			{Name: "Run", Conclusion: "failure", Logs: "FAIL: TestFoo"},
		}}
		pr := &github.PRInfo{Number: 42, HeadRefName: "fix/test"}

		body := formatJobBody(run, job, pr)
		stepsIdx := strings.Index(body, "## Workflow Steps")
		logsIdx := strings.Index(body, "## Logs")
		assert.Greater(t, logsIdx, stepsIdx, "workflow steps should appear before logs")
	})
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "H1 heading",
			body:     "# Fix the null pointer\n\nSome description",
			expected: "Fix the null pointer",
		},
		{
			name:     "bold text when no H1",
			body:     "_⚠️ Potential issue_\n\n**Fragile nodeID reconstruction.**\n\nMore text",
			expected: "Fragile nodeID reconstruction.",
		},
		{
			name:     "H1 takes precedence over bold",
			body:     "**bold first**\n\n# Heading wins",
			expected: "Heading wins",
		},
		{
			name:     "empty body",
			body:     "",
			expected: "",
		},
		{
			name:     "no title markers",
			body:     "just plain text without markers",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractTitle(tc.body))
		})
	}
}

func TestParseDetailsBlocks(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		summaryPrefix string
		expected      []detailsBlock
	}{
		{
			name:          "single matching block",
			body:          `<details><summary>Fix all issues with AI agents</summary>Please fix the linting errors in main.go</details>`,
			summaryPrefix: "Fix all issues with AI agents",
			expected:      []detailsBlock{{Summary: "Fix all issues with AI agents", Body: "Please fix the linting errors in main.go"}},
		},
		{
			name:          "no match",
			body:          `<details><summary>Other summary</summary>content</details>`,
			summaryPrefix: "Fix all issues with AI agents",
			expected:      nil,
		},
		{
			name: "multiple blocks only matching ones returned",
			body: `Some text
<details><summary>Fix all issues with AI agents</summary>
Fix error 1
</details>
<details><summary>Unrelated</summary>ignore this</details>
<details><summary>Fix all issues with AI agents</summary>
Fix error 2
</details>`,
			summaryPrefix: "Fix all issues with AI agents",
			expected: []detailsBlock{
				{Summary: "Fix all issues with AI agents", Body: "Fix error 1"},
				{Summary: "Fix all issues with AI agents", Body: "Fix error 2"},
			},
		},
		{
			name:          "empty body",
			body:          "",
			summaryPrefix: "Fix all issues with AI agents",
			expected:      nil,
		},
		{
			name:          "no details tags",
			body:          "just plain text",
			summaryPrefix: "Fix all issues with AI agents",
			expected:      nil,
		},
		{
			name:          "emoji prefix stripped for matching but preserved in summary",
			body:          `<details><summary>🤖 Fix all issues with AI agents</summary>emoji prefixed content</details>`,
			summaryPrefix: "Fix all issues with AI agents",
			expected:      []detailsBlock{{Summary: "🤖 Fix all issues with AI agents", Body: "emoji prefixed content"}},
		},
		{
			name:          "suggested fix with description in summary",
			body:          `<details><summary>♻️ Suggested fix — capture and reuse the generated ID</summary>Use defer instead</details>`,
			summaryPrefix: "Suggested fix",
			expected:      []detailsBlock{{Summary: "♻️ Suggested fix — capture and reuse the generated ID", Body: "Use defer instead"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDetailsBlocks(tc.body, tc.summaryPrefix)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractNonDetailsText(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "plain text only",
			body:     "This is a description of the issue.",
			expected: "This is a description of the issue.",
		},
		{
			name: "text with details blocks removed",
			body: `_⚠️ Potential issue_ | _🟠 Major_

**Fragile nodeID reconstruction.**

<details>
<summary>♻️ Suggested fix</summary>
code here
</details>`,
			expected: "_⚠️ Potential issue_ | _🟠 Major_\n\n**Fragile nodeID reconstruction.**",
		},
		{
			name:     "HTML comments removed",
			body:     "visible text\n<!-- hidden comment -->\nmore visible",
			expected: "visible text\n\nmore visible",
		},
		{
			name:     "empty after stripping",
			body:     `<details><summary>only details</summary>content</details>`,
			expected: "",
		},
		{
			name: "collapses excessive newlines",
			body: `First line


<!-- comment -->


<details><summary>s</summary>b</details>


Last line`,
			expected: "First line\n\nLast line",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractNonDetailsText(tc.body))
		})
	}
}

func TestSyncCommentTodosCreatesFromMatchingComment(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{{
		ID:     100,
		Body:   `<details><summary>🤖 Fix all issues with AI agents</summary>Please fix the null pointer in handler.go</details>`,
		Author: "reviewer",
		URL:    "https://github.com/org/repo/pull/42#issuecomment-100",
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	path := commentTodoPath(dir, pr.Number, comments[0])
	require.FileExists(t, path)

	parsed, err := todos.ParseFrontmatterFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, types.PriorityMedium, parsed.Frontmatter.Priority)
	assert.Equal(t, types.StatusPending, parsed.Frontmatter.Status)
	assert.Contains(t, parsed.MarkdownContent, "null pointer in handler.go")

	require.NotNil(t, parsed.Frontmatter.PR)
	assert.Equal(t, 42, parsed.Frontmatter.PR.Number)
	assert.Equal(t, "reviewer", parsed.Frontmatter.PR.CommentAuthor)
	assert.Equal(t, "https://github.com/org/repo/pull/42#issuecomment-100", parsed.Frontmatter.PR.CommentURL)
}

func TestSyncCommentTodosCreatesFromSeverity(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{
		{ID: 901, Body: "This variable is unused", Author: "reviewer", Severity: "major", Path: "pkg/handler.go", Line: 10},
		{ID: 902, Body: "Missing error check", Author: "reviewer", Severity: "critical"},
		{ID: 903, Body: "Consider renaming", Author: "reviewer", Severity: "minor"},
	}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	for _, c := range comments {
		path := commentTodoPath(dir, pr.Number, c)
		require.FileExists(t, path, "comment %d with severity %q should create a todo", c.ID, c.Severity)
	}

	parsed, err := todos.ParseFrontmatterFromFile(commentTodoPath(dir, pr.Number, comments[0]))
	require.NoError(t, err)
	assert.Equal(t, types.PriorityMedium, parsed.Frontmatter.Priority)
	assert.Contains(t, parsed.MarkdownContent, "This variable is unused")

	parsed, err = todos.ParseFrontmatterFromFile(commentTodoPath(dir, pr.Number, comments[1]))
	require.NoError(t, err)
	assert.Equal(t, types.PriorityHigh, parsed.Frontmatter.Priority)
}

func TestSyncCommentTodosSkipsNonMatchingComments(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{
		{ID: 200, Body: "LGTM!", Author: "reviewer"},
		{ID: 201, Body: `<details><summary>Other thing</summary>not matching</details>`, Author: "bot"},
	}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	entries, _ := os.ReadDir(dir)
	assert.Empty(t, entries)
}

func TestSyncCommentTodosCreatesFromSuggestedFix(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{{
		ID:     500,
		Body:   `<details><summary>♻️ Suggested fix — capture and reuse the generated ID</summary>Use defer to close the file</details>`,
		Author: "coderabbit",
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	path := commentTodoPath(dir, pr.Number, comments[0])
	require.FileExists(t, path)

	parsed, err := todos.ParseFrontmatterFromFile(path)
	require.NoError(t, err)
	assert.Contains(t, parsed.MarkdownContent, "## ♻️ Suggested fix — capture and reuse the generated ID")
	assert.Contains(t, parsed.MarkdownContent, "Use defer to close the file")
}

func TestSyncCommentTodosFullStructure(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	body := `_⚠️ Potential issue_ | _🟠 Major_

**Fragile nodeID reconstruction — use the value returned by generateNodeID().**

Description of the problem outside details blocks.
<!-- hidden reviewer note -->

<details>
<summary>🤖 Fix all issues with AI agents</summary>
Fix the nil check in handler.go and add proper error handling.
</details>

<details>
<summary>♻️ Suggested fix — capture and reuse the generated ID</summary>
` + "```diff\n-old code\n+new code\n```" + `
</details>

<details>
<summary>📝 Committable suggestion</summary>
ignore this block
</details>`

	comments := []github.PRComment{{
		ID:     600,
		Body:   body,
		Author: "coderabbit",
		URL:    "https://github.com/org/repo/pull/42#discussion_r600",
		Path:   "pkg/handler.go",
		Line:   42,
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	path := commentTodoPath(dir, pr.Number, comments[0])
	require.FileExists(t, path)

	parsed, err := todos.ParseFrontmatterFromFile(path)
	require.NoError(t, err)

	// Title extracted from first bold text
	assert.Equal(t, "Fragile nodeID reconstruction — use the value returned by generateNodeID().", parsed.Frontmatter.Title)

	md := parsed.MarkdownContent
	// File:line info is no longer in body (moved to frontmatter path)
	assert.NotContains(t, md, "File: `pkg/handler.go:42`")
	// Non-details text appears at top
	assert.Contains(t, md, "Fragile nodeID reconstruction")
	assert.Contains(t, md, "_⚠️ Potential issue_")
	// HTML comment stripped
	assert.NotContains(t, md, "hidden reviewer note")
	// Prompt block content inserted directly
	assert.Contains(t, md, "Fix the nil check in handler.go")
	// Suggested fix uses full summary as heading
	assert.Contains(t, md, "## ♻️ Suggested fix — capture and reuse the generated ID")
	// Committable suggestion block is ignored (not matched by our prefixes)
	assert.NotContains(t, md, "ignore this block")

	// Verify ordering: description before prompt before suggested fix
	descIdx := strings.Index(md, "Fragile nodeID")
	promptIdx := strings.Index(md, "Fix the nil check")
	fixIdx := strings.Index(md, "## ♻️ Suggested fix")
	assert.Greater(t, promptIdx, descIdx, "prompt content should come after description")
	assert.Greater(t, fixIdx, promptIdx, "suggested fix should come after prompt content")
}

func TestSyncTodosMultipleJobsIsolation(t *testing.T) {
	dir := t.TempDir()
	workflowYAML := `name: CI
on: [push]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: make lint
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./...
  build:
    runs-on: ubuntu-latest
    steps:
      - run: make build`

	result := &PRWatchResult{
		PR: &github.PRInfo{Number: 99, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: "CI", Status: "completed", Conclusion: "failure",
				WorkflowYAML: workflowYAML,
				Jobs: []github.Job{
					{Name: "lint", Status: "completed", Conclusion: "failure",
						Steps: []github.Step{{Name: "Lint", Conclusion: "failure", Logs: "lint-error-output"}}},
					{Name: "test", Status: "completed", Conclusion: "failure",
						Steps: []github.Step{{Name: "Test", Conclusion: "failure", Logs: "test-error-output"}}},
					{Name: "build", Status: "completed", Conclusion: "success"},
				},
			},
		},
	}

	require.NoError(t, SyncTodos(result, dir))

	// Two failed jobs should create two separate todo files
	lintContent, err := os.ReadFile(filepath.Join(dir, "99", "ci-lint.md"))
	require.NoError(t, err)
	testContent, err := os.ReadFile(filepath.Join(dir, "99", "ci-test.md"))
	require.NoError(t, err)

	// Lint todo: contains only lint content
	assert.Contains(t, string(lintContent), "lint-error-output")
	assert.Contains(t, string(lintContent), "make lint")
	assert.NotContains(t, string(lintContent), "test-error-output")
	assert.NotContains(t, string(lintContent), "go test")

	// Test todo: contains only test content
	assert.Contains(t, string(testContent), "test-error-output")
	assert.Contains(t, string(testContent), "go test")
	assert.NotContains(t, string(testContent), "lint-error-output")
	assert.NotContains(t, string(testContent), "make lint")

	// Build (success) should not have a todo
	assert.NoFileExists(t, filepath.Join(dir, "99", "ci-build.md"))
}

func TestSyncTodosMatrixJobsIsolation(t *testing.T) {
	dir := t.TempDir()
	workflowYAML := `name: Test
on: [push]
jobs:
  test:
    name: "Test (${{ matrix.os }})"
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest]
    steps:
      - run: go test ./...`

	result := &PRWatchResult{
		PR: &github.PRInfo{Number: 99, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: "Test", Status: "completed", Conclusion: "failure",
				WorkflowYAML: workflowYAML,
				Jobs: []github.Job{
					{Name: "Test (ubuntu-latest)", Status: "completed", Conclusion: "failure",
						Steps: []github.Step{{Name: "Run", Conclusion: "failure", Logs: "ubuntu-fail"}}},
					{Name: "Test (windows-latest)", Status: "completed", Conclusion: "success"},
				},
			},
		},
	}

	require.NoError(t, SyncTodos(result, dir))

	// Only the failed matrix instance should create a todo
	ubuntuPath := filepath.Join(dir, "99", "test-test-ubuntu-latest.md")
	require.FileExists(t, ubuntuPath)

	content, err := os.ReadFile(ubuntuPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "ubuntu-fail")
	assert.Contains(t, string(content), "go test", "should extract workflow steps via matrix name matching")

	// Successful matrix instance should not have a todo
	windowsPath := filepath.Join(dir, "99", "test-test-windows-latest.md")
	assert.NoFileExists(t, windowsPath)
}

func TestSyncCommentTodosPathFromComment(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review", BaseRefName: "main", URL: "https://github.com/org/repo/pull/42"}
	comments := []github.PRComment{{
		ID:     700,
		Body:   `<details><summary>🤖 Fix all issues with AI agents</summary>fix it</details>`,
		Author: "reviewer",
		URL:    "https://github.com/org/repo/pull/42#discussion_r700",
		Path:   "pkg/handler.go",
		Line:   10,
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	parsed, err := todos.ParseFrontmatterFromFile(commentTodoPath(dir, pr.Number, comments[0]))
	require.NoError(t, err)
	assert.Equal(t, types.StringOrSlice{"pkg/handler.go:10"}, parsed.Frontmatter.Path)

	require.NotNil(t, parsed.Frontmatter.PR)
	assert.Equal(t, 42, parsed.Frontmatter.PR.Number)
	assert.Equal(t, "https://github.com/org/repo/pull/42", parsed.Frontmatter.PR.URL)
	assert.Equal(t, "feat/review", parsed.Frontmatter.PR.Head)
	assert.Equal(t, "main", parsed.Frontmatter.PR.Base)
	assert.Equal(t, int64(700), parsed.Frontmatter.PR.CommentID)
	assert.Equal(t, "reviewer", parsed.Frontmatter.PR.CommentAuthor)
	assert.Equal(t, "https://github.com/org/repo/pull/42#discussion_r700", parsed.Frontmatter.PR.CommentURL)
}

func TestSyncCommentTodosNoPathWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{{
		ID:     800,
		Body:   `<details><summary>🤖 Fix all issues with AI agents</summary>fix it</details>`,
		Author: "reviewer",
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))

	parsed, err := todos.ParseFrontmatterFromFile(commentTodoPath(dir, pr.Number, comments[0]))
	require.NoError(t, err)
	assert.Nil(t, parsed.Frontmatter.Path)
}

func TestExtractFilePathsFromLogs(t *testing.T) {
	tests := []struct {
		name     string
		job      github.Job
		expected types.StringOrSlice
	}{
		{
			name: "go compile error",
			job: github.Job{Steps: []github.Step{{
				Logs: "pkg/auth/login.go:42:10: undefined: Foo\npkg/auth/session.go:15:3: missing return",
			}}},
			expected: types.StringOrSlice{"pkg/auth/login.go", "pkg/auth/session.go"},
		},
		{
			name: "deduplicates files",
			job: github.Job{Steps: []github.Step{{
				Logs: "main.go:1:1: error\nmain.go:5:2: another error",
			}}},
			expected: types.StringOrSlice{"main.go"},
		},
		{
			name:     "no file references",
			job:      github.Job{Steps: []github.Step{{Logs: "FAIL: TestFoo"}}},
			expected: nil,
		},
		{
			name: "falls back to job logs",
			job: github.Job{
				Logs:  "cmd/main.go:10:5: error here",
				Steps: []github.Step{{Logs: "no file refs here"}},
			},
			expected: types.StringOrSlice{"cmd/main.go"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractFilePathsFromLogs(tc.job))
		})
	}
}

func TestSyncTodosPathFromCIFailure(t *testing.T) {
	dir := t.TempDir()
	result := &PRWatchResult{
		PR: &github.PRInfo{Number: 99, HeadRefName: "feat/x", BaseRefName: "main"},
		Runs: map[int64]*github.WorkflowRun{
			1: {
				DatabaseID: 1, Name: "CI", Status: "completed", Conclusion: "failure",
				Jobs: []github.Job{{
					Name: "build", Status: "completed", Conclusion: "failure",
					Steps: []github.Step{{Name: "Build", Conclusion: "failure", Logs: "pkg/api/handler.go:42:10: undefined: Response"}},
				}},
			},
		},
	}

	require.NoError(t, SyncTodos(result, dir))

	parsed, err := todos.ParseFrontmatterFromFile(filepath.Join(dir, "99", "ci-build.md"))
	require.NoError(t, err)
	assert.Equal(t, types.StringOrSlice{"pkg/api/handler.go"}, parsed.Frontmatter.Path)
}

func TestSyncCommentTodosIdempotent(t *testing.T) {
	dir := t.TempDir()
	pr := &github.PRInfo{Number: 42, HeadRefName: "feat/review"}
	comments := []github.PRComment{{
		ID:     300,
		Body:   `<details><summary>🤖 Fix all issues with AI agents</summary>fix this</details>`,
		Author: "reviewer",
	}}

	require.NoError(t, SyncCommentTodos(comments, pr, dir))
	path := commentTodoPath(dir, pr.Number, comments[0])
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	require.NoError(t, SyncCommentTodos(comments, pr, dir))
	after, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Equal(t, string(before), string(after))
}
