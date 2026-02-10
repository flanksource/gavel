package todos

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// Validator defines the interface for running validation checks
type Validator interface {
	runArchUnitCheck(ctx context.Context) error
	runLinter(ctx context.Context, language types.Language) error
	runBuild(ctx context.Context) error
}

// ValidationPipeline orchestrates sequential validation steps
type ValidationPipeline struct {
	validator Validator
	executor  *TODOExecutor // For custom validations
	language  types.Language
	workDir   string
}

// NewValidationPipeline creates a new validation pipeline
func NewValidationPipeline(language types.Language, workDir string, executor *TODOExecutor) *ValidationPipeline {
	return &ValidationPipeline{
		validator: &defaultValidator{workDir: workDir},
		executor:  executor,
		language:  language,
		workDir:   workDir,
	}
}

// Validate runs all validation steps sequentially
// Fails fast on first error
func (vp *ValidationPipeline) Validate(ctx context.Context, todo *types.TODO, changedFiles []string) (*fixtures.Stats, error) {
	stats := &fixtures.Stats{}

	// 1. arch-unit todo check
	if err := vp.validator.runArchUnitCheck(ctx); err != nil {
		stats.Failed++
		return stats, fmt.Errorf("arch-unit check failed: %w", err)
	}
	stats.Passed++

	// 2. Language-specific linter
	if err := vp.validator.runLinter(ctx, vp.language); err != nil {
		stats.Failed++
		return stats, fmt.Errorf("linter failed: %w", err)
	}
	stats.Passed++

	// 3. make build
	if err := vp.validator.runBuild(ctx); err != nil {
		stats.Failed++
		return stats, fmt.Errorf("build failed: %w", err)
	}
	stats.Passed++

	// 4. Custom validations (using fixtures executor)
	if len(todo.CustomValidations) > 0 && vp.executor != nil {
		results := vp.executor.ExecuteSection(ctx, todo.CustomValidations)
		if !AllPassed(results) {
			stats.Failed++
			return stats, fmt.Errorf("custom validations failed")
		}
		stats.Passed++
	}

	return stats, nil
}

// defaultValidator implements Validator interface with actual command execution
type defaultValidator struct {
	workDir string
}

func (v *defaultValidator) runArchUnitCheck(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "arch-unit", "todo", "check")
	cmd.Dir = v.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("arch-unit check failed: %s", string(output))
	}
	return nil
}

func (v *defaultValidator) runLinter(ctx context.Context, language types.Language) error {
	var cmd *exec.Cmd

	switch language {
	case types.LanguageGo:
		cmd = exec.CommandContext(ctx, "golangci-lint", "run", "./...")
	case types.LanguageTypeScript:
		cmd = exec.CommandContext(ctx, "npm", "run", "lint")
	case types.LanguagePython:
		cmd = exec.CommandContext(ctx, "pylint", "**/*.py")
	default:
		return fmt.Errorf("unsupported language: %s", language)
	}

	cmd.Dir = v.workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("linter failed: %s", string(output))
	}
	return nil
}

func (v *defaultValidator) runBuild(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "make", "build")
	cmd.Dir = v.workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %s", string(output))
	}
	return nil
}
