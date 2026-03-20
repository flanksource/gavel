package fixtures

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/stretchr/testify/assert"
)

func TestMergeExpectations(t *testing.T) {
	tests := []struct {
		name           string
		inlineAttrs    map[string]string
		yamlExpects    *Expectations
		expected       *Expectations
		expectedFields map[string]interface{}
	}{
		{
			name:        "inline exitCode overrides YAML",
			inlineAttrs: map[string]string{"exitCode": "0"},
			yamlExpects: &Expectations{ExitCode: intPtr(1)},
			expectedFields: map[string]interface{}{
				"exitCode": 0,
			},
		},
		{
			name:        "inline timeout overrides YAML",
			inlineAttrs: map[string]string{"timeout": "30"},
			yamlExpects: &Expectations{Timeout: durationPtr(60 * time.Second)},
			expectedFields: map[string]interface{}{
				"timeout": 30 * time.Second,
			},
		},
		{
			name:        "non-conflicting fields merge",
			inlineAttrs: map[string]string{"exitCode": "0"},
			yamlExpects: &Expectations{
				Stdout:  "success",
				Timeout: durationPtr(30 * time.Second),
			},
			expectedFields: map[string]interface{}{
				"exitCode": 0,
				"stdout":   "success",
				"timeout":  30 * time.Second,
			},
		},
		{
			name:        "inline only",
			inlineAttrs: map[string]string{"exitCode": "1", "timeout": "10"},
			yamlExpects: nil,
			expectedFields: map[string]interface{}{
				"exitCode": 1,
				"timeout":  10 * time.Second,
			},
		},
		{
			name:        "YAML only",
			inlineAttrs: nil,
			yamlExpects: &Expectations{
				ExitCode: intPtr(0),
				Stdout:   "output",
				Stderr:   "error",
			},
			expectedFields: map[string]interface{}{
				"exitCode": 0,
				"stdout":   "output",
				"stderr":   "error",
			},
		},
		{
			name:        "both empty",
			inlineAttrs: nil,
			yamlExpects: nil,
			expectedFields: map[string]interface{}{
				"exitCode": 0, // default
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeExpectations(tt.inlineAttrs, tt.yamlExpects)

			if exitCode, ok := tt.expectedFields["exitCode"]; ok {
				assert.NotNil(t, result.ExitCode)
				assert.Equal(t, exitCode, *result.ExitCode)
			}

			if stdout, ok := tt.expectedFields["stdout"]; ok {
				assert.Equal(t, stdout, result.Stdout)
			}

			if stderr, ok := tt.expectedFields["stderr"]; ok {
				assert.Equal(t, stderr, result.Stderr)
			}

			if timeout, ok := tt.expectedFields["timeout"]; ok {
				assert.NotNil(t, result.Timeout)
				assert.Equal(t, timeout, *result.Timeout)
			}
		})
	}
}

func TestParseInlineExpectations(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]string
		expected *Expectations
	}{
		{
			name:  "exitCode only",
			attrs: map[string]string{"exitCode": "1"},
			expected: &Expectations{
				ExitCode: intPtr(1),
			},
		},
		{
			name:  "timeout only",
			attrs: map[string]string{"timeout": "30"},
			expected: &Expectations{
				Timeout: durationPtr(30 * time.Second),
			},
		},
		{
			name:  "both exitCode and timeout",
			attrs: map[string]string{"exitCode": "0", "timeout": "60"},
			expected: &Expectations{
				ExitCode: intPtr(0),
				Timeout:  durationPtr(60 * time.Second),
			},
		},
		{
			name:     "empty",
			attrs:    map[string]string{},
			expected: &Expectations{},
		},
		{
			name:     "invalid exitCode ignored",
			attrs:    map[string]string{"exitCode": "invalid"},
			expected: &Expectations{},
		},
		{
			name:     "invalid timeout ignored",
			attrs:    map[string]string{"timeout": "invalid"},
			expected: &Expectations{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInlineExpectations(tt.attrs)

			if tt.expected.ExitCode != nil {
				assert.NotNil(t, result.ExitCode)
				assert.Equal(t, *tt.expected.ExitCode, *result.ExitCode)
			} else {
				assert.Nil(t, result.ExitCode)
			}

			if tt.expected.Timeout != nil {
				assert.NotNil(t, result.Timeout)
				assert.Equal(t, *tt.expected.Timeout, *result.Timeout)
			} else {
				assert.Nil(t, result.Timeout)
			}
		})
	}
}

func TestCELFailureMessageIncludesContext(t *testing.T) {
	exitCode := 0
	exp := Expectations{
		ExitCode: &exitCode,
		CEL:      "stdout.contains('hello')",
	}
	fixture := FixtureResult{
		Name:     "test-cel",
		Status:   "pending",
		Metadata: map[string]interface{}{},
		Test: FixtureTest{
			Name: "test-cel",
		},
	}
	result := exp.Evaluate(fixture, exec.ExecResult{
		Stdout:   "goodbye world",
		ExitCode: 0,
	})

	assert.Equal(t, task.StatusFAIL, result.Status)
	assert.Contains(t, result.Error, "CEL expression evaluated to false")
	assert.Contains(t, result.Error, "stdout.contains('hello')")
	assert.Contains(t, result.Error, "stdout=goodbye world")
}

func TestExitCodeFailureIncludesOutput(t *testing.T) {
	exitCode := 0
	exp := Expectations{ExitCode: &exitCode}
	fixture := FixtureResult{
		Name:     "test-exit-code",
		Status:   "pending",
		Metadata: map[string]interface{}{},
		Test:     FixtureTest{Name: "test-exit-code"},
	}
	result := exp.Evaluate(fixture, exec.ExecResult{
		Stdout:   "some output here",
		Stderr:   "error: something went wrong",
		ExitCode: 1,
	})

	assert.Equal(t, task.StatusFAIL, result.Status)
	assert.Contains(t, result.Error, "expected exit code 0, got 1")
	assert.Contains(t, result.Error, "some output here")
	assert.Contains(t, result.Error, "error: something went wrong")
}

func TestTruncateForError(t *testing.T) {
	short := "hello"
	assert.Equal(t, "hello", truncateForError(short))

	long := strings.Repeat("x", 300)
	result := truncateForError(long)
	assert.Equal(t, 203, len(result)) // 200 + "..."
	assert.True(t, strings.HasSuffix(result, "..."))
}

func TestFormatCELVarsLimit(t *testing.T) {
	vars := map[string]any{
		"stdout": strings.Repeat("a", 250),
	}
	result := formatCELVars(vars)
	assert.Contains(t, result, "stdout=")
	assert.Contains(t, result, "...")
	assert.Less(t, len(result), 250)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
