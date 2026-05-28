package fixtures

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
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

// EvaluateOptions carries optional context for Expectations.Evaluate.
// It is a separate struct so future fields (e.g. an io.Writer for logs)
// can be added without breaking call sites.
type EvaluateOptions struct {
	// SourceDir is the directory of the fixture markdown file, used to
	// resolve @-prefixed file references in Stdout/Stderr expectations.
	SourceDir string
	// UpdateGolden, when true, causes mismatched @file expectations to
	// be overwritten with the actual output instead of failing the test.
	UpdateGolden bool
}

func (e Expectations) Evaluate(fixture FixtureResult, p exec.ExecResult, opts EvaluateOptions) FixtureResult {

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
		return fixture.Failf("expected exit code %d, got %d\n  stdout: %s\n  stderr: %s",
			expectedExitCode, p.ExitCode,
			truncateForError(p.Stdout), truncateForError(p.Stderr))
	}
	if fixture.Metadata == nil {
		fixture.Metadata = map[string]interface{}{}
	}
	if updated, failed := evaluateStream(&fixture, e.Stdout, p.Stdout, "stdout", opts); failed {
		return fixture
	} else if updated {
		fixture.Metadata["golden_updated_stdout"] = true
	}
	if updated, failed := evaluateStream(&fixture, e.Stderr, p.Stderr, "stderr", opts); failed {
		return fixture
	} else if updated {
		fixture.Metadata["golden_updated_stderr"] = true
	}
	if e.CEL != "" {
		// Use RunExpression for CEL expressions, not RunTemplate
		t := fixture.Test.AsMap()
		t["output"] = p.Stdout
		t["stdout"] = p.Stdout
		t["stderr"] = p.Stderr
		t["exitCode"] = p.ExitCode
		combined := p.Stdout + p.Stderr
		dups := duplicateLines(combined)
		dupList := make([]map[string]any, 0, len(dups))
		for _, d := range dups {
			dupList = append(dupList, map[string]any{"text": d.Text, "count": d.Count})
		}
		t["ansi"] = map[string]any{
			"has_any":         hasAnyANSI(combined),
			"has_color":       hasColorCodes(combined),
			"has_updates":     hasCursorUpdates(combined),
			"has_cursor_hide": hasCursorHide(combined),
			"has_cursor_show": hasCursorShow(combined),
			"has_reset":       hasSGRReset(combined),
			"stray_controls":  hasStrayControls(combined),
			"final_text":      finalText(combined),
			"duplicate_lines": dupList,
			"has_duplicates":  len(dups) > 0,
		}
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
			CelEnvs:    ANSICelFunctions(),
		})
		if err != nil {
			return fixture.Errorf(err, "failed to evaluate CEL expression with gomplate")
		}

		switch v := output.(type) {
		case bool:
			if !v {
				fixture.CELExpression = e.CEL
				fixture.CELVars = t
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

func truncateForError(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// evaluateStream compares an expected value (literal or @file) against
// got. Returns (updated=true) when the @file was rewritten under
// UpdateGolden mode; returns (failed=true) when the fixture should
// stop and be marked failed (fixture.Failf has already been called).
// If expected is empty, the stream is not checked.
func evaluateStream(fixture *FixtureResult, expected, got, label string, opts EvaluateOptions) (updated, failed bool) {
	if expected == "" {
		return false, false
	}
	ref, err := ResolveFileRef(opts.SourceDir, expected)
	if err != nil {
		*fixture = fixture.Errorf(err, "load expected %s", label)
		return false, true
	}
	wantText, wantLabel := ref.Raw, label+" (expected)"
	if ref.IsFile {
		wantText, wantLabel = ref.Contents, ref.Path
	}
	diff := UnifiedDiff(wantText, got, wantLabel, label)
	if diff == "" {
		return false, false
	}
	if opts.UpdateGolden && ref.IsFile {
		if err := WriteGolden(ref.Path, got); err != nil {
			*fixture = fixture.Errorf(err, "update golden %s", ref.Path)
			return false, true
		}
		logger.Infof("golden updated: %s", ref.Path)
		return true, false
	}
	*fixture = fixture.Failf("%s differs:\n%s", label, diff)
	return false, true
}
