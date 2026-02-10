package fixtures

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gomplate/v3"
)

// Expectations represents expected outcomes for code block execution.
// It can be populated from inline code fence attributes or YAML expects blocks.
type Expectations struct {
	ExitCode *int `yaml:"exitCode,omitempty" json:"exitCode,omitempty"`
	// Matches stdout contains expected string
	Stdout string `yaml:"stdout,omitempty" json:"stdout,omitempty"`
	// Matches stderr contains expected string
	Stderr string `yaml:"stderr,omitempty" json:"stderr,omitempty"`
	// Verifies a non-zero exit code and stderr contains expected substring
	Error string `yaml:"error,omitempty" json:"error,omitempty"`
	// Verifies output format (e.g., json, yaml)
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
	// Matches expected output substring in either stdout or stderr
	Count      *int                   `yaml:"count,omitempty" json:"count,omitempty"`
	Output     string                 `yaml:"output,omitempty" json:"output,omitempty"`
	Timeout    *time.Duration         `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	CEL        string                 `yaml:"cel,omitempty" json:"cel,omitempty"`
	Properties map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
}

func (e Expectations) Evaluate(fixture FixtureResult, p exec.ExecResult) FixtureResult {

	fixture.Stdout = p.Stdout
	fixture.Stderr = p.Stderr
	fixture.ExitCode = p.ExitCode
	// Build full command string
	if p.Command != "" {
		if len(p.Args) > 0 {
			fixture.Command = p.Command + " " + strings.Join(p.Args, " ")
		} else {
			fixture.Command = p.Command
		}
	}
	// Default exit code expectation to 0 if not specified
	expectedExitCode := 0
	if e.ExitCode != nil {
		expectedExitCode = *e.ExitCode
	}
	if p.ExitCode != expectedExitCode {
		return fixture.Failf("expected exit code %d, got %d", expectedExitCode, p.ExitCode)
	}
	if e.Stdout != "" && p.Stdout != e.Stdout {
		return fixture.Failf("expected stdout:\n%s\n got:\n%s", e.Stdout, p.Stdout)
	}
	if e.Stderr != "" && p.Stderr != e.Stderr {
		return fixture.Failf("expected stderr:\n%s\n got:\n%s", e.Stderr, p.Stderr)
	}
	if e.CEL != "" {
		// Use RunExpression for CEL expressions, not RunTemplate
		t := fixture.Test.AsMap()
		t["output"] = p.Stdout
		t["stdout"] = p.Stdout
		t["stderr"] = p.Stderr
		t["exitCode"] = p.ExitCode
		// Try to parse JSON output if it looks like JSON
		if strings.HasPrefix(strings.TrimSpace(p.Stdout), "{") || strings.HasPrefix(strings.TrimSpace(p.Stdout), "[") {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(p.Stdout), &jsonData); err == nil {
				t["json"] = jsonData
				fixture.Metadata["json"] = jsonData
			}
		}

		// Add temp file data to CEL context
		for name, tempFile := range fixture.Test.TempFiles {
			t[name] = tempFile.GetCELData()
		}
		output, err := gomplate.RunExpression(t, gomplate.Template{
			Expression: e.CEL,
		})
		if err != nil {
			return fixture.Errorf(err, "failed to evaluate CEL expression with gomplate")
		}

		switch v := output.(type) {
		case bool:
			if !v {
				return fixture.Failf("CEL expression evaluated to false")
			}
		case string:
			if strings.ToLower(strings.TrimSpace(v)) != "true" {
				return fixture.Failf("%s != true", v)
			}
		default:
			return fixture.Failf("CEL expression did not return a boolean: got %T(%v)", output, output)
		}
	}
	fixture.Status = task.StatusPASS
	return fixture
}

// ParseInlineExpectations converts inline code fence attributes to Expectations.
// Supports: exitCode=N, timeout=N (seconds)
func ParseInlineExpectations(attrs map[string]string) *Expectations {
	exp := &Expectations{}

	if exitCodeStr, ok := attrs["exitCode"]; ok {
		if exitCode, err := strconv.Atoi(exitCodeStr); err == nil {
			exp.ExitCode = &exitCode
		}
	}

	if timeoutStr, ok := attrs["timeout"]; ok {
		if timeoutSec, err := strconv.Atoi(timeoutStr); err == nil {
			timeout := time.Duration(timeoutSec) * time.Second
			exp.Timeout = &timeout
		}
	}

	return exp
}

// MergeExpectations merges inline attributes with YAML expects block.
// Inline attributes take precedence over YAML values on conflicts.
// Non-conflicting fields from both sources are combined.
func MergeExpectations(inlineAttrs map[string]string, yamlExpects *Expectations) *Expectations {
	// Start with YAML expects (or empty if nil)
	result := &Expectations{}
	if yamlExpects != nil {
		result.ExitCode = yamlExpects.ExitCode
		result.Stdout = yamlExpects.Stdout
		result.Stderr = yamlExpects.Stderr
		result.Timeout = yamlExpects.Timeout
	}

	// Parse inline attributes
	if len(inlineAttrs) > 0 {
		inline := ParseInlineExpectations(inlineAttrs)

		// Inline overrides YAML
		if inline.ExitCode != nil {
			result.ExitCode = inline.ExitCode
		}
		if inline.Timeout != nil {
			result.Timeout = inline.Timeout
		}
	}

	// Default exitCode to 0 if not specified
	if result.ExitCode == nil {
		defaultExitCode := 0
		result.ExitCode = &defaultExitCode
	}

	return result
}
