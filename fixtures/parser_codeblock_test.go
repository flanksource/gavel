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

func TestParseInlineCodeFenceAttributes(t *testing.T) {
	tests := []struct {
		name       string
		infoString string
		expected   map[string]string
	}{
		{
			name:       "exitCode only",
			infoString: "bash exitCode=1",
			expected:   map[string]string{"exitCode": "1"},
		},
		{
			name:       "timeout only",
			infoString: "bash timeout=30",
			expected:   map[string]string{"timeout": "30"},
		},
		{
			name:       "both exitCode and timeout",
			infoString: "bash exitCode=1 timeout=30",
			expected:   map[string]string{"exitCode": "1", "timeout": "30"},
		},
		{
			name:       "language only, no attributes",
			infoString: "bash",
			expected:   map[string]string{},
		},
		{
			name:       "with extra spaces",
			infoString: "bash  exitCode=1   timeout=30  ",
			expected:   map[string]string{"exitCode": "1", "timeout": "30"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := parseCodeFenceAttributes(tt.infoString)
			assert.Equal(t, tt.expected, attrs)
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
