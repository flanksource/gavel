package oxlint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// Oxlint implements the Linter interface for the oxlint JavaScript/TypeScript linter.
type Oxlint struct {
	linters.RunOptions
	fileCount int
	ruleCount int
}

// NewOxlint creates a new oxlint linter rooted at workDir.
func NewOxlint(workDir string) *Oxlint {
	return &Oxlint{
		RunOptions: linters.RunOptions{
			WorkDir: workDir,
		},
	}
}

// SetOptions sets the run options for the linter.
func (o *Oxlint) SetOptions(opts linters.RunOptions) {
	o.RunOptions = opts
}

// Name returns the linter name.
func (o *Oxlint) Name() string {
	return "oxlint"
}

// ProjectRootMarkers identifies a JS/TS project — a package.json or any of
// oxlint's own config file names. Listing config files means a repo that ships
// an oxlint config without a package.json still lints.
func (o *Oxlint) ProjectRootMarkers() []string {
	return []string{
		"package.json",
		".oxlintrc.json",
		".oxlintrc.jsonc",
		"oxlint.json",
		"oxlintrc.json",
	}
}

// DefaultIncludes returns default file patterns this linter should process.
func (o *Oxlint) DefaultIncludes() []string {
	return []string{
		"**/*.js",
		"**/*.jsx",
		"**/*.ts",
		"**/*.tsx",
		"**/*.mjs",
		"**/*.cjs",
		"**/*.mts",
		"**/*.cts",
		"**/*.vue",
		"**/*.astro",
		"**/*.svelte",
	}
}

// DefaultExcludes returns patterns this linter should ignore by default.
// Note: Common patterns like .git/**, node_modules/**, examples/**, hack/** are
// handled by the all_language_excludes macro. This only returns oxlint-specific excludes.
func (o *Oxlint) DefaultExcludes() []string {
	return []string{
		"bower_components/**", // Bower packages (legacy package manager)
		"jspm_packages/**",    // JSPM packages (legacy package manager)
		"public/**",           // Public assets (usually not source code)
		".cache/**",           // Cache directories
	}
}

// GetSupportedLanguages returns the languages this linter can process.
func (o *Oxlint) GetSupportedLanguages() []string {
	return []string{"javascript", "typescript"}
}

// GetEffectiveExcludes returns the complete list of exclusion patterns
// using the all_language_excludes macro for the given language and config.
func (o *Oxlint) GetEffectiveExcludes(language string, config *models.Config) []string {
	if config == nil {
		return o.DefaultExcludes()
	}
	return config.GetAllLanguageExcludes(language, o.DefaultExcludes())
}

// GetEffectiveIncludes returns the complete list of inclusion patterns
// for the given language and config.
func (o *Oxlint) GetEffectiveIncludes(language string, config *models.Config) []string {
	if config == nil {
		return o.DefaultIncludes()
	}
	return config.GetAllLanguageIncludes(language, o.DefaultIncludes())
}

// SupportsJSON returns true if linter supports JSON output.
func (o *Oxlint) SupportsJSON() bool {
	return true
}

// JSONArgs returns additional args needed for JSON output.
func (o *Oxlint) JSONArgs() []string {
	return []string{"--format=json"}
}

// SupportsFix returns true if linter supports auto-fixing violations.
func (o *Oxlint) SupportsFix() bool {
	return true
}

// FixArgs returns additional args needed for fix mode.
func (o *Oxlint) FixArgs() []string {
	return []string{"--fix"}
}

// ValidateConfig validates linter-specific configuration.
func (o *Oxlint) ValidateConfig(config *models.LinterConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// GetFileCount reports the number of files oxlint scanned in the last run.
func (o *Oxlint) GetFileCount() int {
	return o.fileCount
}

// GetRuleCount reports the number of rules oxlint evaluated in the last run.
func (o *Oxlint) GetRuleCount() int {
	return o.ruleCount
}

// buildArgs assembles the argv (without the command name) that Run would use.
func (o *Oxlint) buildArgs() []string {
	var args []string
	if o.Config != nil {
		args = append(args, o.Config.Args...)
	}
	if o.ForceJSON && !o.hasFormatArg(args) {
		args = append(args, "--format=json")
	}
	if o.Fix && o.SupportsFix() && !o.hasArg(args, "--fix") {
		args = append(args, o.FixArgs()...)
	}
	args = append(args, o.ExtraArgs...)
	if len(o.Files) > 0 {
		args = append(args, o.Files...)
	} else if !o.hasPathArg(args) {
		args = append(args, ".")
	}
	return args
}

// DryRunCommand reports the command oxlint would execute.
func (o *Oxlint) DryRunCommand() (string, []string) {
	return "oxlint", o.buildArgs()
}

// Run executes oxlint and returns violations.
//
// oxlint writes the JSON report to stdout and diagnostic warnings (e.g. "No
// files found to lint") to stderr, so we capture the two streams separately to
// keep the JSON parseable.
func (o *Oxlint) Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error) {
	args := o.buildArgs()

	cmd := exec.CommandContext(ctx, "oxlint", args...)
	cmd.Dir = o.WorkDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = o.WrapWriter(&stdout)
	cmd.Stderr = &stderr

	logger.Infof("Executing: oxlint %s", strings.Join(args, " "))

	err := cmd.Run()

	// oxlint exits with 1 when deny-level (error) diagnostics are found, which is
	// expected when there are violations. Any other non-zero code is a real failure.
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			if stdout.Len() > 0 {
				logger.Debugf("oxlint exit code 1 with output - treating as success with violations")
				err = nil
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("oxlint execution failed: %w\nOutput:\n%s", err, stderr.String())
	}

	if stdout.Len() == 0 {
		return []models.Violation{}, nil
	}

	return o.parseViolations(stdout.Bytes())
}

// hasFormatArg checks if the args already contain a format argument.
func (o *Oxlint) hasFormatArg(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--format") || strings.HasPrefix(arg, "-f") {
			return true
		}
	}
	return false
}

// hasArg checks if the exact flag is already present.
func (o *Oxlint) hasArg(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// hasPathArg checks if the args already contain a path argument.
func (o *Oxlint) hasPathArg(args []string) bool {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return true
		}
	}
	return false
}

// parseViolations parses oxlint JSON output into violations.
func (o *Oxlint) parseViolations(output []byte) ([]models.Violation, error) {
	var report OxlintReport
	if err := json.Unmarshal(output, &report); err != nil {
		logger.Debugf("Failed to parse oxlint JSON output: %v\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("failed to parse oxlint JSON output: %w", err)
	}

	o.fileCount = report.NumberOfFiles
	o.ruleCount = report.NumberOfRules

	var violations []models.Violation
	for i := range report.Diagnostics {
		violations = append(violations, report.Diagnostics[i].ToViolation(o.WorkDir))
	}

	return violations, nil
}

// OxlintReport is the top-level oxlint JSON document.
type OxlintReport struct {
	Diagnostics   []OxlintDiagnostic `json:"diagnostics"`
	NumberOfFiles int                `json:"number_of_files"`
	NumberOfRules int                `json:"number_of_rules"`
}

// OxlintDiagnostic is a single diagnostic from oxlint.
type OxlintDiagnostic struct {
	Message  string        `json:"message"`
	Code     string        `json:"code"`
	Severity string        `json:"severity"`
	Help     string        `json:"help"`
	URL      string        `json:"url"`
	Filename string        `json:"filename"`
	Labels   []OxlintLabel `json:"labels"`
}

// OxlintLabel locates a diagnostic within a file.
type OxlintLabel struct {
	Label string     `json:"label,omitempty"`
	Span  OxlintSpan `json:"span"`
}

// OxlintSpan is a 1-based line/column position within a file.
type OxlintSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ToViolation converts an OxlintDiagnostic to a generic Violation.
func (d *OxlintDiagnostic) ToViolation(workDir string) models.Violation {
	filename := d.Filename
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(workDir, filename)
	}

	var line, column int
	if len(d.Labels) > 0 {
		line = d.Labels[0].Span.Line
		column = d.Labels[0].Span.Column
	}

	rule := ruleName(d.Code)

	return models.NewViolationBuilder().
		WithFile(filename).
		WithLocation(line, column).
		WithCaller(filepath.Dir(filename), "unknown").
		WithCalled("oxlint", rule).
		WithMessage(d.Message).
		WithSource("oxlint").
		WithRuleFromLinter("oxlint", rule).
		Build()
}

// ruleName extracts the rule from an oxlint code such as "eslint(no-debugger)",
// returning "eslint/no-debugger". Codes without a plugin wrapper are returned
// unchanged.
func ruleName(code string) string {
	if code == "" {
		return "unknown"
	}
	open := strings.IndexByte(code, '(')
	if open < 0 || !strings.HasSuffix(code, ")") {
		return code
	}
	plugin := strings.TrimSpace(code[:open])
	inner := code[open+1 : len(code)-1]
	if plugin == "" {
		return inner
	}
	return plugin + "/" + inner
}
