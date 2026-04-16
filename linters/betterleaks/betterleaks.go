package betterleaks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// Betterleaks implements the Linter interface for the betterleaks
// secret-scanning tool (a gitleaks-compatible successor).
type Betterleaks struct {
	linters.RunOptions
}

func NewBetterleaks(workDir string) *Betterleaks {
	return &Betterleaks{
		RunOptions: linters.RunOptions{WorkDir: workDir},
	}
}

func (b *Betterleaks) SetOptions(opts linters.RunOptions) {
	b.RunOptions = opts
}

func (b *Betterleaks) Name() string { return "betterleaks" }

// DefaultIncludes: secrets can hide in any text file. Betterleaks has its own
// binary-skipping + allowlist logic so we cast a wide net here.
func (b *Betterleaks) DefaultIncludes() []string {
	return []string{"**/*"}
}

func (b *Betterleaks) DefaultExcludes() []string {
	return []string{
		".git/**",
		"node_modules/**",
		"vendor/**",
		"**/testdata/**",
		"*.min.*",
		".tmp/**",
	}
}

func (b *Betterleaks) SupportsJSON() bool { return true }
func (b *Betterleaks) JSONArgs() []string { return []string{"--report-format", "json"} }
func (b *Betterleaks) SupportsFix() bool  { return false }
func (b *Betterleaks) FixArgs() []string  { return nil }

func (b *Betterleaks) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// reportPath returns the absolute path where we ask betterleaks to write its
// JSON report. Kept deterministic per-workdir so successive runs overwrite
// rather than accumulate.
func (b *Betterleaks) reportPath() string {
	return filepath.Join(b.WorkDir, ".tmp", "betterleaks-report.json")
}

// buildArgs returns the argv (without the command name) that Run would use.
// tomlPath is optional — when empty, no -c flag is added and betterleaks uses
// its built-in ruleset or auto-detects a local config.
func (b *Betterleaks) buildArgs(tomlPath, reportPath string) []string {
	args := []string{"dir"}
	if len(b.Files) > 0 {
		args = append(args, b.Files...)
	} else {
		args = append(args, ".")
	}
	args = append(args,
		"--report-format", "json",
		"--report-path", reportPath,
		"--no-banner",
		"--no-color",
		// Stay exit-0 on findings; we read the JSON report for violations.
		"--exit-code", "0",
	)
	if tomlPath != "" {
		args = append(args, "-c", tomlPath)
	}
	if b.Config != nil {
		args = append(args, b.Config.Args...)
	}
	args = append(args, b.ExtraArgs...)
	return args
}

// DryRunCommand reports the command that would be executed. The merged
// secrets config is NOT materialized in dry-run mode; we show a placeholder
// so users see the intent without side effects.
func (b *Betterleaks) DryRunCommand() (string, []string) {
	args := b.buildArgs("<merged-secrets-config>", b.reportPath())
	return "betterleaks", args
}

func (b *Betterleaks) Run(ctx commonsContext.Context, _ *clicky.Task) ([]models.Violation, error) {
	configs := DiscoverConfigs(b.WorkDir)
	tomlPath, err := ResolveConfig(b.WorkDir, configs)
	if err != nil {
		return nil, fmt.Errorf("resolve betterleaks config: %w", err)
	}

	reportPath := b.reportPath()
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return nil, fmt.Errorf("create betterleaks report dir: %w", err)
	}
	_ = os.Remove(reportPath)

	args := b.buildArgs(tomlPath, reportPath)
	cmd := exec.CommandContext(ctx, "betterleaks", args...)
	cmd.Dir = b.WorkDir

	logger.Infof("Executing: betterleaks %s", strings.Join(args, " "))
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return nil, fmt.Errorf("betterleaks execution failed: %w\nOutput:\n%s", runErr, string(output))
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No report file means betterleaks ran cleanly with zero findings.
			return []models.Violation{}, nil
		}
		return nil, fmt.Errorf("read betterleaks report %s: %w", reportPath, err)
	}
	return b.parseFindings(data)
}

// Finding matches the JSON shape betterleaks emits (an array of these). Only
// the fields gavel actually displays are modeled; unknown fields are ignored.
type Finding struct {
	RuleID      string   `json:"RuleID"`
	Description string   `json:"Description"`
	File        string   `json:"File"`
	StartLine   int      `json:"StartLine"`
	EndLine     int      `json:"EndLine"`
	StartColumn int      `json:"StartColumn"`
	EndColumn   int      `json:"EndColumn"`
	Match       string   `json:"Match"`
	Secret      string   `json:"Secret"`
	Entropy     float64  `json:"Entropy"`
	Fingerprint string   `json:"Fingerprint"`
	Commit      string   `json:"Commit"`
	Tags        []string `json:"Tags"`
}

func (b *Betterleaks) parseFindings(data []byte) ([]models.Violation, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		return []models.Violation{}, nil
	}
	var findings []Finding
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, fmt.Errorf("parse betterleaks JSON: %w\nOutput:\n%s", err, string(data))
	}
	violations := make([]models.Violation, 0, len(findings))
	for i := range findings {
		violations = append(violations, findings[i].ToViolation(b.WorkDir))
	}
	return violations, nil
}

// ToViolation converts a Finding into a gavel Violation. The raw Secret is
// never copied into the violation — only the rule id + redacted length —
// so gavel output cannot itself become a secret-leak vector.
func (f *Finding) ToViolation(workDir string) models.Violation {
	file := f.File
	if !filepath.IsAbs(file) && workDir != "" {
		file = filepath.Join(workDir, file)
	}

	redacted := redactSecret(f.Match, f.Secret)
	msg := f.Description
	if msg == "" {
		msg = f.RuleID
	}
	if redacted != "" {
		msg = fmt.Sprintf("%s (match: %s)", msg, redacted)
	}

	v := models.NewViolationBuilder().
		WithFile(file).
		WithLocation(f.StartLine, f.StartColumn).
		WithMessage(msg).
		WithSource("betterleaks").
		WithRuleFromLinter("betterleaks", f.RuleID).
		Build()
	v.Severity = models.SeverityError
	return v
}

// redactSecret returns a placeholder of the same length as the detected
// secret so users see something about the match without leaking the value.
func redactSecret(match, secret string) string {
	if secret == "" {
		if match == "" {
			return ""
		}
		return fmt.Sprintf("<%d chars>", len(match))
	}
	return fmt.Sprintf("<redacted %d chars>", len(secret))
}
