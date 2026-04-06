package fixtures

import (
	"runtime"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/stretchr/testify/assert"
)

func TestStatsAddCountsStatusFAIL(t *testing.T) {
	tests := []struct {
		name           string
		status         task.Status
		expectedFailed int
		expectedPassed int
	}{
		{name: "StatusFAIL counts as failed", status: task.StatusFAIL, expectedFailed: 1},
		{name: "StatusFailed counts as failed", status: task.StatusFailed, expectedFailed: 1},
		{name: "StatusPASS counts as passed", status: task.StatusPASS, expectedPassed: 1},
		{name: "StatusERR counts as error", status: task.StatusERR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Stats{}
			result := &FixtureResult{Status: tt.status}
			s = s.Add(result)
			assert.Equal(t, 1, s.Total)
			assert.Equal(t, tt.expectedFailed, s.Failed)
			assert.Equal(t, tt.expectedPassed, s.Passed)
		})
	}
}

func TestFixtureResultStatsCountsStatusFAIL(t *testing.T) {
	result := FixtureResult{Status: task.StatusFAIL}
	stats := result.Stats()
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Total)
	assert.Equal(t, 0, stats.Passed)
}

func TestStatsVisitCountsStatusFAIL(t *testing.T) {
	s := &Stats{}
	node := &FixtureNode{
		Results: &FixtureResult{Status: task.StatusFAIL},
	}
	s.Visit(node)
	assert.Equal(t, 1, s.Total)
	assert.Equal(t, 1, s.Failed)
}

func TestStatsHasFailuresWithStatusFAIL(t *testing.T) {
	result := FixtureResult{Status: task.StatusFAIL}
	stats := result.Stats()
	assert.True(t, stats.HasFailures())
	assert.False(t, stats.IsOK())
}

func TestAsMapIncludesCustomColumns(t *testing.T) {
	fixture := FixtureTest{
		Name: "test",
		FrontMatter: FrontMatter{
			Metadata: map[string]any{
				"globalVar":  "from-frontmatter",
				"overridden": "global-value",
			},
		},
		Expected: Expectations{
			Properties: map[string]any{
				"url":        "https://example.com",
				"overridden": "column-value",
			},
		},
		TemplateVars: map[string]any{
			"file": "test.md",
		},
	}

	m := fixture.AsMap()
	assert.Equal(t, "https://example.com", m["url"], "custom column should be in AsMap")
	assert.Equal(t, "from-frontmatter", m["globalVar"], "frontmatter metadata should be in AsMap")
	assert.Equal(t, "column-value", m["overridden"], "Properties should override Metadata")
	assert.Equal(t, "test.md", m["file"], "TemplateVars should override Properties")
}

func TestAsMapTemplatesCustomColumns(t *testing.T) {
	fixture := FixtureTest{
		Name: "template test",
		ExecFixtureBase: ExecFixtureBase{
			Exec: "curl",
			Args: []string{"{{.url}}", "--header", "X-Token: {{.token}}"},
		},
		Expected: Expectations{
			Properties: map[string]any{
				"url":   "https://example.com/api",
				"token": "abc123",
			},
		},
	}

	exec, err := fixture.ExecBase().Template(fixture.AsMap())
	assert.NoError(t, err)
	assert.Equal(t, "curl", exec.Exec)
	assert.Equal(t, []string{"https://example.com/api", "--header", "X-Token: abc123"}, exec.Args)
}

func TestCWDParsing(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		frontMatter    *FrontMatter
		expectedCWD    string
		expectedFMCWD  string
		expectedMerged string
	}{
		{
			name:    "file-level frontmatter CWD inherited by test",
			content: "### command: test\n```bash\npwd\n```\n\nValidations:\n* cel: exitCode == 0\n",
			frontMatter: &FrontMatter{
				ExecFixtureBase: ExecFixtureBase{CWD: "./project-root"},
			},
			expectedCWD:    "",
			expectedFMCWD:  "./project-root",
			expectedMerged: "./project-root",
		},
		{
			name:    "test-level CWD overrides frontmatter CWD",
			content: "### command: test\n```bash\npwd\n```\n\n```frontmatter\ncwd: ./specific-dir\n```\n\nValidations:\n* cel: exitCode == 0\n",
			frontMatter: &FrontMatter{
				ExecFixtureBase: ExecFixtureBase{CWD: "./default-dir"},
			},
			expectedCWD:    "./specific-dir",
			expectedFMCWD:  "./default-dir",
			expectedMerged: "./specific-dir",
		},
		{
			name:           "no CWD set anywhere",
			content:        "### command: test\n```bash\necho hello\n```\n",
			frontMatter:    nil,
			expectedCWD:    "",
			expectedFMCWD:  "",
			expectedMerged: "",
		},
		{
			name:           "only test-level CWD set via frontmatter block",
			content:        "### command: test\n```bash\npwd\n```\n\n```frontmatter\ncwd: /absolute/path\n```\n",
			frontMatter:    nil,
			expectedCWD:    "/absolute/path",
			expectedFMCWD:  "",
			expectedMerged: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures, err := parseMarkdownWithGoldmark(tt.content, tt.frontMatter, "/fixtures")
			assert.NoError(t, err)
			assert.Len(t, fixtures, 1)

			f := fixtures[0].Test
			assert.Equal(t, tt.expectedCWD, f.CWD, "test-level CWD")
			assert.Equal(t, tt.expectedFMCWD, f.FrontMatter.CWD, "frontmatter CWD")
			assert.Equal(t, tt.expectedMerged, f.ExecBase().CWD, "merged CWD")
		})
	}
}

func TestCWDFromTableColumns(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedCWD string
	}{
		{
			name:        "CWD column",
			content:     "| Name | CWD | CLI Args |\n|------|-----|----------|\n| test | ./mydir | --help |\n",
			expectedCWD: "./mydir",
		},
		{
			name:        "working directory column",
			content:     "| Name | Working Directory | CLI Args |\n|------|-------------------|----------|\n| test | /tmp/work | --help |\n",
			expectedCWD: "/tmp/work",
		},
		{
			name:        "dir column",
			content:     "| Name | Dir | CLI Args |\n|------|-----|----------|\n| test | ./subdir | --help |\n",
			expectedCWD: "./subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures, err := parseMarkdownWithGoldmark(tt.content, nil, "/tmp/test")
			assert.NoError(t, err)
			assert.Len(t, fixtures, 1)
			assert.Equal(t, tt.expectedCWD, fixtures[0].Test.CWD)
		})
	}
}

func TestSourceDirSetFromParser(t *testing.T) {
	content := "### command: test\n```bash\necho hello\n```\n"
	fixtures, err := parseMarkdownWithGoldmark(content, nil, "/path/to/fixtures")
	assert.NoError(t, err)
	assert.Len(t, fixtures, 1)
	assert.Equal(t, "/path/to/fixtures", fixtures[0].Test.SourceDir)
}

func TestMergeIntoCWD(t *testing.T) {
	tests := []struct {
		name        string
		base        ExecFixtureBase
		other       ExecFixtureBase
		expectedCWD string
	}{
		{
			name:        "other CWD overrides base",
			base:        ExecFixtureBase{CWD: "./base"},
			other:       ExecFixtureBase{CWD: "./other"},
			expectedCWD: "./other",
		},
		{
			name:        "base CWD used when other is empty",
			base:        ExecFixtureBase{CWD: "./base"},
			other:       ExecFixtureBase{},
			expectedCWD: "./base",
		},
		{
			name:        "both empty stays empty",
			base:        ExecFixtureBase{},
			other:       ExecFixtureBase{},
			expectedCWD: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.base.MergeInto(tt.other)
			assert.Equal(t, tt.expectedCWD, result.CWD)
		})
	}
}

func TestFrontMatterShouldSkip(t *testing.T) {
	tests := []struct {
		name     string
		fm       FrontMatter
		wantSkip bool
	}{
		{
			name:     "no constraints",
			fm:       FrontMatter{},
			wantSkip: false,
		},
		{
			name:     "matching os",
			fm:       FrontMatter{OS: runtime.GOOS},
			wantSkip: false,
		},
		{
			name:     "non-matching os",
			fm:       FrontMatter{OS: "plan9"},
			wantSkip: true,
		},
		{
			name:     "negated os excludes current",
			fm:       FrontMatter{OS: "!" + runtime.GOOS},
			wantSkip: true,
		},
		{
			name:     "negated os allows other",
			fm:       FrontMatter{OS: "!plan9"},
			wantSkip: false,
		},
		{
			name:     "matching arch",
			fm:       FrontMatter{Arch: runtime.GOARCH},
			wantSkip: false,
		},
		{
			name:     "non-matching arch",
			fm:       FrontMatter{Arch: "mips"},
			wantSkip: true,
		},
		{
			name:     "skip command returns true",
			fm:       FrontMatter{Skip: "true"},
			wantSkip: true,
		},
		{
			name:     "skip command returns false",
			fm:       FrontMatter{Skip: "false"},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := tt.fm.ShouldSkip()
			if tt.wantSkip {
				assert.NotEmpty(t, reason)
			} else {
				assert.Empty(t, reason)
			}
		})
	}
}
