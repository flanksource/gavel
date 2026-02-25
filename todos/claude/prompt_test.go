package claude

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

func newTestTODO(name, impl string) *types.TODO {
	return &types.TODO{
		FilePath: ".todos/" + name + ".md",
		TODOFrontmatter: types.TODOFrontmatter{
			Title: name,
		},
		Implementation: impl,
		StepsToReproduce: []*fixtures.FixtureNode{
			{Test: &fixtures.FixtureTest{Name: name + "-repro", ExecFixtureBase: fixtures.ExecFixtureBase{Exec: "go test ./..."}}},
		},
		Verification: []*fixtures.FixtureNode{
			{Test: &fixtures.FixtureTest{Name: name + "-verify", ExecFixtureBase: fixtures.ExecFixtureBase{Exec: "go test -run TestFoo"}}},
		},
	}
}

func TestBuildPrompt(t *testing.T) {
	todo := newTestTODO("fix-auth", "Fix the auth handler")
	prompt := BuildPrompt(todo, "")

	for _, want := range []string{
		"You are fixing a failing test",
		"## Steps to Reproduce",
		"## Implementation",
		"Fix the auth handler",
		"## Verification",
		"Do NOT run git add or git commit",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("BuildPrompt missing %q", want)
		}
	}
}

func TestBuildGroupPrompt(t *testing.T) {
	todos := []*types.TODO{
		newTestTODO("fix-auth", "Fix the auth handler"),
		newTestTODO("fix-db", "Fix the database query"),
		newTestTODO("fix-cache", "Fix the cache invalidation"),
	}

	prompt := BuildGroupPrompt(todos, "")

	for _, want := range []string{
		"implementing multiple related fixes",
		"## fix-auth",
		"## fix-db",
		"## fix-cache",
		"Fix the auth handler",
		"Fix the database query",
		"Fix the cache invalidation",
		"Implement ALL TODOs",
		"Do NOT run git add or git commit",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("BuildGroupPrompt missing %q", want)
		}
	}
}

func TestBuildGroupPromptSingleTODO(t *testing.T) {
	todos := []*types.TODO{newTestTODO("solo", "Single task")}
	prompt := BuildGroupPrompt(todos, "")

	if !strings.Contains(prompt, "## solo") {
		t.Error("single-element group should contain title heading")
	}
	if !strings.Contains(prompt, "Single task") {
		t.Error("should contain implementation text")
	}
}

func TestBuildPromptWithPR(t *testing.T) {
	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.PR = &types.PR{
		Number:        42,
		URL:           "https://github.com/org/repo/pull/42",
		Head:          "feat/review",
		Base:          "main",
		CommentAuthor: "reviewer",
		CommentURL:    "https://github.com/org/repo/pull/42#discussion_r100",
	}
	prompt := BuildPrompt(todo, "")

	for _, want := range []string{
		"## PR Context",
		"#42",
		"feat/review",
		"reviewer",
		"discussion_r100",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("BuildPrompt with PR missing %q", want)
		}
	}
}

func TestBuildPromptWithCustomPrompt(t *testing.T) {
	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.Prompt = "Focus on the null pointer dereference"
	prompt := BuildPrompt(todo, "")

	if !strings.Contains(prompt, "## Prompt") {
		t.Error("should contain Prompt section")
	}
	if !strings.Contains(prompt, "Focus on the null pointer dereference") {
		t.Error("should contain custom prompt text")
	}
}

func TestBuildPromptWithSourceCode(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pkg/auth.go", 50)

	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.Path = types.StringOrSlice{"pkg/auth.go:25"}
	prompt := BuildPrompt(todo, dir)

	if strings.Contains(prompt, "## Referenced Source Code") {
		t.Error("should not contain old Referenced Source Code section")
	}
	if !strings.Contains(prompt, "```go file=pkg/auth.go:25") {
		t.Error("should contain annotated code block with file=")
	}
	if !strings.Contains(prompt, "line 25 content") {
		t.Error("should contain actual source code lines")
	}
}

func TestBuildGroupPromptExcludesPRButIncludesSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "pkg/auth.go", 50)

	todo := newTestTODO("fix-auth", "Fix the auth handler")
	todo.PR = &types.PR{Number: 42, URL: "https://example.com/pull/42", Head: "feat/x", Base: "main"}
	todo.Path = types.StringOrSlice{"pkg/auth.go:25"}

	prompt := BuildGroupPrompt([]*types.TODO{todo}, dir)

	if strings.Contains(prompt, "## PR Context") {
		t.Error("grouped prompt should not contain PR Context section")
	}
	if !strings.Contains(prompt, "```go file=pkg/auth.go:25") {
		t.Error("grouped prompt should contain annotated code block")
	}
	if !strings.Contains(prompt, "Fix the auth handler") {
		t.Error("grouped prompt should still contain implementation")
	}
}

func TestLangFromExt(t *testing.T) {
	tests := []struct {
		ext, expected string
	}{
		{".go", "go"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".js", "javascript"},
		{".py", "python"},
		{".sql", "sql"},
		{".yaml", "yaml"},
		{".yml", "yaml"},
		{".json", "json"},
		{".sh", "bash"},
		{".rs", "rust"},
		{".rb", "ruby"},
		{".unknown", ""},
		{"", ""},
	}
	for _, tc := range tests {
		if got := langFromExt(tc.ext); got != tc.expected {
			t.Errorf("langFromExt(%q) = %q, want %q", tc.ext, got, tc.expected)
		}
	}
}

func TestStripFileRefLine(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"removes file:line ref", "File: `pkg/handler.go:42`\n\nSome description", "Some description"},
		{"removes file ref without line", "File: `pkg/handler.go`\n\nSome description", "Some description"},
		{"no-op when absent", "Just a description", "Just a description"},
		{"strips from middle", "Before\nFile: `foo.go:10`\nAfter", "Before\nAfter"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripFileRefLine(tc.input); got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}
