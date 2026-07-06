package fixtures

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestParseFrontmatterCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected []string
	}{
		{
			name:     "single language",
			yaml:     "codeBlocks: [bash]",
			expected: []string{"bash"},
		},
		{
			name:     "multiple languages",
			yaml:     "codeBlocks: [bash, go, python]",
			expected: []string{"bash", "go", "python"},
		},
		{
			name:     "empty defaults to bash",
			yaml:     "",
			expected: []string{"bash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fm FrontMatter
			if tt.yaml != "" {
				err := yaml.Unmarshal([]byte(tt.yaml), &fm)
				assert.NoError(t, err)
			}

			// Apply default
			if len(fm.CodeBlocks) == 0 {
				fm.CodeBlocks = []string{"bash"}
			}

			assert.Equal(t, tt.expected, fm.CodeBlocks)
		})
	}
}

func TestExtractLanguageFromInfoString(t *testing.T) {
	tests := []struct {
		name       string
		infoString string
		expected   string
	}{
		{
			name:       "bash only",
			infoString: "bash",
			expected:   "bash",
		},
		{
			name:       "bash with attributes",
			infoString: "bash exitCode=1",
			expected:   "bash",
		},
		{
			name:       "go with attributes",
			infoString: "go timeout=60",
			expected:   "go",
		},
		{
			name:       "empty",
			infoString: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := extractLanguage(tt.infoString)
			assert.Equal(t, tt.expected, lang)
		})
	}
}

func TestRunnerStepKind(t *testing.T) {
	tests := []struct {
		name       string
		infoString string
		expected   string
	}{
		{name: "yaml test", infoString: "yaml test", expected: "test"},
		{name: "yaml lint", infoString: "yaml lint", expected: "lint"},
		{name: "bare test", infoString: "test", expected: "test"},
		{name: "bare lint", infoString: "lint", expected: "lint"},
		{name: "case insensitive", infoString: "YAML Test", expected: "test"},
		{name: "trailing space", infoString: "yaml test ", expected: "test"},
		{name: "plain yaml is not a step", infoString: "yaml", expected: ""},
		{name: "bash is not a step", infoString: "bash", expected: ""},
		{name: "bash with attrs", infoString: "bash exitCode=1", expected: ""},
		{name: "empty", infoString: "", expected: ""},
		{name: "non-yaml prefix rejected", infoString: "json test", expected: ""},
		{name: "extra tokens rejected", infoString: "yaml test extra", expected: ""},
		{name: "test not last rejected", infoString: "test yaml", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, runnerStepKind(tt.infoString))
		})
	}
}

func TestShouldExecuteCodeBlock(t *testing.T) {
	tests := []struct {
		name       string
		language   string
		codeBlocks []string
		expected   bool
	}{
		{
			name:       "bash in list",
			language:   "bash",
			codeBlocks: []string{"bash"},
			expected:   true,
		},
		{
			name:       "go not in list",
			language:   "go",
			codeBlocks: []string{"bash"},
			expected:   false,
		},
		{
			name:       "multiple languages, match",
			language:   "python",
			codeBlocks: []string{"bash", "go", "python"},
			expected:   true,
		},
		{
			name:       "exec in list",
			language:   "exec",
			codeBlocks: []string{"exec"},
			expected:   true,
		},
		{
			name:       "empty language",
			language:   "",
			codeBlocks: []string{"bash"},
			expected:   false,
		},
		{
			name:       "case insensitive",
			language:   "Bash",
			codeBlocks: []string{"bash"},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExecuteCodeBlock(tt.language, tt.codeBlocks)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseExecutableFenceConfigObject(t *testing.T) {
	tree, err := ParseMarkdownContentWithTree("suite.md", `# Exec

`+"```exec"+`
content: |
  echo ok
exitCode: 0
cel: stdout.contains("ok")
properties:
  target: api
`+"```"+`
`, ".", &FrontMatter{})
	assert.NoError(t, err)

	test := firstFixtureTest(tree)
	if assert.NotNil(t, test) {
		assert.Equal(t, "bash", test.Exec)
		assert.Equal(t, []string{"-c", "echo ok\n"}, test.Args)
		if assert.NotNil(t, test.Expected.ExitCode) {
			assert.Equal(t, 0, *test.Expected.ExitCode)
		}
		assert.Equal(t, `stdout.contains("ok")`, test.Expected.CEL)
		assert.Equal(t, "api", test.Expected.Properties["target"])
	}
}

func firstFixtureTest(node *FixtureNode) *FixtureTest {
	if node == nil {
		return nil
	}
	if node.Test != nil {
		return node.Test
	}
	for _, child := range node.Children {
		if found := firstFixtureTest(child); found != nil {
			return found
		}
	}
	return nil
}
