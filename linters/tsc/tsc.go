// Package tsc implements a gavel linter that runs the TypeScript compiler
// in type-check-only mode and surfaces compile errors as violations.
//
// It shells out to `node` running an embedded wrapper script (tsc-json.cjs)
// that uses the TypeScript compiler API to emit diagnostics as JSON. The
// wrapper resolves the `typescript` package from the project's own
// node_modules so the user's pinned compiler version is used.
//
// Config.Args and ExtraArgs are ignored: tsc's behavior is driven by the
// project's tsconfig.json. Override the tsconfig by editing it, not by
// passing flags.
package tsc

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// The wrapper is named .cjs so Node always treats it as CommonJS, even when
// the project's package.json sets "type": "module" (which would otherwise
// block require()).
//
//go:embed tsc-json.cjs
var tscJSONScript []byte

// TSC implements the Linter interface for the TypeScript compiler.
type TSC struct {
	linters.RunOptions
}

func NewTSC(workDir string) *TSC {
	return &TSC{RunOptions: linters.RunOptions{WorkDir: workDir}}
}

func (t *TSC) SetOptions(opts linters.RunOptions) { t.RunOptions = opts }

func (t *TSC) Name() string { return "tsc" }

func (t *TSC) DefaultIncludes() []string {
	return []string{
		"**/*.ts",
		"**/*.tsx",
		"**/*.mts",
		"**/*.cts",
	}
}

func (t *TSC) DefaultExcludes() []string {
	return []string{
		"*.d.ts",
		"dist/**",
		"build/**",
		".next/**",
		"coverage/**",
	}
}

func (t *TSC) GetSupportedLanguages() []string { return []string{"typescript"} }

func (t *TSC) GetEffectiveExcludes(language string, config *models.Config) []string {
	if config == nil {
		return t.DefaultExcludes()
	}
	return config.GetAllLanguageExcludes(language, t.DefaultExcludes())
}

func (t *TSC) GetEffectiveIncludes(language string, config *models.Config) []string {
	if config == nil {
		return t.DefaultIncludes()
	}
	return config.GetAllLanguageIncludes(language, t.DefaultIncludes())
}

func (t *TSC) SupportsJSON() bool { return true }
func (t *TSC) JSONArgs() []string { return nil }
func (t *TSC) SupportsFix() bool  { return false }
func (t *TSC) FixArgs() []string  { return nil }

func (t *TSC) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// resolveScript writes the embedded wrapper into <WorkDir>/.tmp/ keyed by a
// content hash so repeated runs reuse the same file. Returns the absolute path.
func (t *TSC) resolveScript() (string, error) {
	if t.WorkDir == "" {
		return "", fmt.Errorf("tsc: WorkDir is empty")
	}
	dir := filepath.Join(t.WorkDir, ".tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create .tmp dir: %w", err)
	}
	name := fmt.Sprintf("gavel-tsc-json-%s.cjs", scriptHash(tscJSONScript))
	path := filepath.Join(dir, name)

	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, tscJSONScript) {
		return path, nil
	}
	if err := os.WriteFile(path, tscJSONScript, 0o644); err != nil {
		return "", fmt.Errorf("write wrapper script: %w", err)
	}
	return path, nil
}

func (t *TSC) DryRunCommand() (string, []string) {
	path, err := t.resolveScript()
	if err != nil {
		return "node", []string{"<gavel-tsc-wrapper>"}
	}
	return "node", []string{path}
}

func (t *TSC) Run(ctx commonsContext.Context, _ *clicky.Task) ([]models.Violation, error) {
	if t.Config != nil && len(t.Config.Args) > 0 {
		logger.Debugf("tsc: Config.Args is ignored; configure via tsconfig.json")
	}
	if len(t.ExtraArgs) > 0 {
		logger.Debugf("tsc: ExtraArgs is ignored; configure via tsconfig.json")
	}

	scriptPath, err := t.resolveScript()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "node", scriptPath)
	cmd.Dir = t.WorkDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = t.WrapWriter(&stdout)
	cmd.Stderr = t.WrapWriter(&stderr)

	logger.Infof("Executing: node %s (cwd=%s)", scriptPath, t.WorkDir)

	runErr := cmd.Run()
	if runErr != nil {
		if _, ok := runErr.(*exec.Error); ok {
			return nil, fmt.Errorf("tsc wrapper failed to launch (is node installed?): %w\nStderr:\n%s", runErr, stderr.String())
		}
		return nil, fmt.Errorf("tsc wrapper failed: %w\nStderr:\n%s", runErr, stderr.String())
	}

	return parseViolations(stdout.Bytes(), t.WorkDir, io.Discard)
}

// parseViolations decodes the wrapper's JSON payload into violations.
// extra is reserved for future diagnostic output; callers pass io.Discard.
func parseViolations(output []byte, workDir string, _ io.Writer) ([]models.Violation, error) {
	trimmed := bytes.TrimSpace(output)
	if len(trimmed) == 0 {
		return []models.Violation{}, nil
	}

	var diagnostics []TSCDiagnostic
	if err := json.Unmarshal(trimmed, &diagnostics); err != nil {
		return nil, fmt.Errorf("parse tsc JSON output: %w\nRaw output:\n%s", err, string(output))
	}

	violations := make([]models.Violation, 0, len(diagnostics))
	for _, d := range diagnostics {
		violations = append(violations, d.toViolation(workDir))
	}
	return violations, nil
}

// TSCDiagnostic mirrors one entry emitted by tsc-json.js.
type TSCDiagnostic struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Code     int    `json:"code"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

func (d *TSCDiagnostic) toViolation(workDir string) models.Violation {
	filename := d.File
	if filename != "" && !filepath.IsAbs(filename) {
		filename = filepath.Join(workDir, filename)
	}

	rule := fmt.Sprintf("TS%d", d.Code)

	v := models.NewViolationBuilder().
		WithFile(filename).
		WithLocation(d.Line, d.Column).
		WithMessage(d.Message).
		WithSource("tsc").
		WithRuleFromLinter("tsc", rule).
		Build()
	v.Severity = categoryToSeverity(d.Category)
	return v
}

func categoryToSeverity(category string) models.ViolationSeverity {
	switch category {
	case "Error":
		return models.SeverityError
	case "Warning":
		return models.SeverityWarning
	default:
		return models.SeverityInfo
	}
}
