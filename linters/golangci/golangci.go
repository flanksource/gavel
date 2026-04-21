package golangci

import (
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

// GolangciLint implements the Linter interface for golangci-lint
type GolangciLint struct {
	linters.RunOptions
}

// NewGolangciLint creates a new golangci-lint linter
func NewGolangciLint(workDir string) *GolangciLint {
	return &GolangciLint{
		RunOptions: linters.RunOptions{
			WorkDir: workDir,
		},
	}
}

// SetOptions sets the run options for the linter
func (g *GolangciLint) SetOptions(opts linters.RunOptions) {
	g.RunOptions = opts
}

// Name returns the linter name
func (g *GolangciLint) Name() string {
	return "golangci-lint"
}

// ProjectRootMarkers identifies the file(s) that anchor a Go module. The lint
// driver runs golangci-lint once per discovered module with WorkDir set to the
// module directory.
func (g *GolangciLint) ProjectRootMarkers() []string {
	return []string{"go.mod"}
}

// DefaultIncludes returns default file patterns this linter should process
func (g *GolangciLint) DefaultIncludes() []string {
	return []string{"**/*.go"}
}

// DefaultExcludes returns patterns this linter should ignore by default
// Note: Common patterns like .git/**, vendor/**, examples/**, hack/** are now
// handled by the all_language_excludes macro. This only returns GolangciLint-specific excludes.
func (g *GolangciLint) DefaultExcludes() []string {
	return []string{
		"**/*_gen.go",    // Generated Go files
		"**/*.pb.go",     // Protocol buffer generated files
		"**/testdata/**", // Go test data directories
	}
}

// GetSupportedLanguages returns the languages this linter can process
func (g *GolangciLint) GetSupportedLanguages() []string {
	return []string{"go"}
}

// GetEffectiveExcludes returns the complete list of exclusion patterns
// using the all_language_excludes macro for the given language and config
func (g *GolangciLint) GetEffectiveExcludes(language string, config *models.Config) []string {
	if config == nil {
		// Fallback to default excludes if no config
		return g.DefaultExcludes()
	}

	// Use the all_language_excludes macro
	return config.GetAllLanguageExcludes(language, g.DefaultExcludes())
}

// GetEffectiveIncludes returns the complete list of inclusion patterns
// for the given language and config
func (g *GolangciLint) GetEffectiveIncludes(language string, config *models.Config) []string {
	if config == nil {
		// Fallback to default includes if no config
		return g.DefaultIncludes()
	}

	// Use the combined includes system
	return config.GetAllLanguageIncludes(language, g.DefaultIncludes())
}

// SupportsJSON returns true if linter supports JSON output
func (g *GolangciLint) SupportsJSON() bool {
	return true
}

// JSONArgs returns additional args needed for JSON output
func (g *GolangciLint) JSONArgs() []string {
	return []string{"--output.json.path=stdout", "--output.text.path=stderr"}
}

// SupportsFix returns true if linter supports auto-fixing violations
func (g *GolangciLint) SupportsFix() bool {
	return true
}

// FixArgs returns additional args needed for fix mode
func (g *GolangciLint) FixArgs() []string {
	return []string{"--fix"}
}

// ValidateConfig validates linter-specific configuration
func (g *GolangciLint) ValidateConfig(config *models.LinterConfig) error {
	// Basic validation - could be extended
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	return nil
}

// buildArgs assembles the argv (without the command name) that Run would use.
func (g *GolangciLint) buildArgs() []string {
	args := []string{"run"}

	if g.Config != nil {
		args = append(args, g.Config.Args...)
	}

	if g.ForceJSON && !g.hasFormatArg(args, "--output.json") {
		args = append(args, "--output.json.path=stdout", "--output.text.path=stderr")
	}

	args = append(args, g.ExtraArgs...)

	if len(g.Files) > 0 {
		args = append(args, g.Files...)
	} else if !g.hasPathArg(args) {
		args = append(args, ".")
	}
	return args
}

// DryRunCommand reports the command golangci-lint would execute.
func (g *GolangciLint) DryRunCommand() (string, []string) {
	return "golangci-lint", g.buildArgs()
}

// Run executes golangci-lint and returns violations
func (g *GolangciLint) Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error) {
	args := g.buildArgs()

	// Execute command — capture stdout (JSON) separately from stderr (text)
	cmd := exec.CommandContext(ctx, "golangci-lint", args...)
	cmd.Dir = g.WorkDir

	logger.Infof("Executing: golangci-lint %s", strings.Join(args, " "))

	var stdout, stderr strings.Builder
	cmd.Stdout = g.WrapWriter(&stdout)
	cmd.Stderr = g.WrapWriter(&stderr)

	err := cmd.Run()

	// Handle golangci-lint specific exit codes
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Exit code 1 means violations found — this is expected
			if stdout.Len() > 0 {
				logger.Debugf("golangci-lint exit code 1 with output - treating as success with violations")
				err = nil
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("golangci-lint execution failed: %w\nOutput:\n%s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []models.Violation{}, nil
	}

	// golangci-lint v2 outputs JSON on the first line, followed by a text summary.
	// Only parse the first line which contains the JSON object.
	if idx := strings.Index(output, "\n"); idx > 0 {
		output = output[:idx]
	}

	return g.parseViolations([]byte(output))
}

// hasFormatArg checks if the args already contain a format argument
func (g *GolangciLint) hasFormatArg(args []string, formatPrefix string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, formatPrefix) {
			return true
		}
	}
	return false
}

// hasPathArg checks if the args already contain a path argument
func (g *GolangciLint) hasPathArg(args []string) bool {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return true
		}
	}
	return false
}

// parseViolations parses golangci-lint JSON output into violations
func (g *GolangciLint) parseViolations(output []byte) ([]models.Violation, error) {
	var result GolangciOutput
	if err := json.Unmarshal(output, &result); err != nil {
		// If JSON parsing fails, log the output for debugging
		logger.Debugf("Failed to parse golangci-lint JSON output: %v\nOutput: %s", err, string(output))
		return nil, fmt.Errorf("failed to parse golangci-lint JSON output: %w", err)
	}

	var violations []models.Violation
	for _, issue := range result.Issues {
		violation := issue.ToViolation(g.WorkDir)
		violations = append(violations, violation)
	}

	return violations, nil
}

// GolangciOutput represents the JSON structure from golangci-lint
type GolangciOutput struct {
	Issues []GolangciIssue `json:"Issues"`
}

// GolangciIssue represents a single issue from golangci-lint
type GolangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
}

// ToViolation converts a GolangciIssue to a generic Violation
func (issue *GolangciIssue) ToViolation(workDir string) models.Violation {
	// Use Pos field as default, but for typecheck errors extract real location from Text
	filename := issue.Pos.Filename
	line := issue.Pos.Line
	column := issue.Pos.Column
	message := issue.Text

	// For typecheck errors, parse the actual location from the Text field
	if issue.FromLinter == "typecheck" {
		// Look for patterns like "./file.go:line:col: message" in multi-line text
		lines := strings.Split(message, "\n")
		for _, textLine := range lines {
			if strings.HasPrefix(textLine, "./") {
				if colonIdx := strings.Index(textLine, ":"); colonIdx != -1 {
					// Found a file reference like "./lint_test.go:6:2: ..."
					remainder := textLine[2:] // Skip "./"
					parts := strings.SplitN(remainder, ":", 4)
					if len(parts) >= 3 {
						// parts[0] = filename, parts[1] = line, parts[2] = column, parts[3] = message
						filename = parts[0]
						if l := parseInt(parts[1]); l > 0 {
							line = l
						}
						if c := parseInt(parts[2]); c > 0 {
							column = c
						}
						if len(parts) >= 4 {
							message = strings.TrimSpace(parts[3])
						}
						break
					}
				}
			}
		}
	}

	// Clean message using the same logic as the original implementation
	message = cleanGolangciMessage(message)

	// Make filename absolute if it's relative
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(workDir, filename)
	}

	return models.NewViolationBuilder().
		WithFile(filename).
		WithLocation(line, column).
		WithCaller(filepath.Dir(filename), "unknown").
		WithCalled("golangci-lint", issue.FromLinter).
		WithMessage(message).
		WithSource("golangci-lint").
		WithRuleFromLinter("golangci-lint", issue.FromLinter).
		Build()
}
