package fixtures

import (
	"testing"
	"time"

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

// Helper functions
func intPtr(i int) *int {
	return &i
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
