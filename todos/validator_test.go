package todos

import (
	"context"
	"fmt"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

func TestValidationPipeline_AllPass(t *testing.T) {
	// Create a simple TODO with no custom validations
	todo := &types.TODO{
		FilePath:          ".todos/test.md",
		CustomValidations: []*fixtures.FixtureNode{},
	}

	// Mock validator that succeeds for all checks
	validator := &mockValidator{
		archUnitPass: true,
		linterPass:   true,
		buildPass:    true,
	}

	vp := &ValidationPipeline{
		validator: validator,
		language:  types.LanguageGo,
	}

	stats, err := vp.Validate(context.Background(), todo, []string{})
	if err != nil {
		t.Fatalf("Expected validation to pass, got error: %v", err)
	}

	// Should have 3 passed: arch-unit, linter, build
	if stats.Passed != 3 {
		t.Errorf("Expected 3 passed validations, got %d", stats.Passed)
	}

	if stats.Failed != 0 {
		t.Errorf("Expected 0 failed validations, got %d", stats.Failed)
	}
}

func TestValidationPipeline_ArchUnitCheckFails(t *testing.T) {
	todo := &types.TODO{
		FilePath:          ".todos/test.md",
		CustomValidations: []*fixtures.FixtureNode{},
	}

	// Mock validator where arch-unit check fails
	validator := &mockValidator{
		archUnitPass: false,
		linterPass:   true,
		buildPass:    true,
	}

	vp := &ValidationPipeline{
		validator: validator,
		language:  types.LanguageGo,
	}

	stats, err := vp.Validate(context.Background(), todo, []string{})
	if err == nil {
		t.Fatal("Expected validation to fail when arch-unit check fails")
	}

	// Should fail immediately, no other checks run
	if stats.Passed != 0 {
		t.Errorf("Expected 0 passed validations, got %d", stats.Passed)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed validation, got %d", stats.Failed)
	}
}

func TestValidationPipeline_LinterFails(t *testing.T) {
	todo := &types.TODO{
		FilePath:          ".todos/test.md",
		CustomValidations: []*fixtures.FixtureNode{},
	}

	// Mock validator where linter fails
	validator := &mockValidator{
		archUnitPass: true,
		linterPass:   false,
		buildPass:    true,
	}

	vp := &ValidationPipeline{
		validator: validator,
		language:  types.LanguageGo,
	}

	stats, err := vp.Validate(context.Background(), todo, []string{})
	if err == nil {
		t.Fatal("Expected validation to fail when linter fails")
	}

	// arch-unit passed, linter failed
	if stats.Passed != 1 {
		t.Errorf("Expected 1 passed validation, got %d", stats.Passed)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed validation, got %d", stats.Failed)
	}
}

func TestValidationPipeline_BuildFails(t *testing.T) {
	todo := &types.TODO{
		FilePath:          ".todos/test.md",
		CustomValidations: []*fixtures.FixtureNode{},
	}

	// Mock validator where build fails
	validator := &mockValidator{
		archUnitPass: true,
		linterPass:   true,
		buildPass:    false,
	}

	vp := &ValidationPipeline{
		validator: validator,
		language:  types.LanguageGo,
	}

	stats, err := vp.Validate(context.Background(), todo, []string{})
	if err == nil {
		t.Fatal("Expected validation to fail when build fails")
	}

	// arch-unit and linter passed, build failed
	if stats.Passed != 2 {
		t.Errorf("Expected 2 passed validations, got %d", stats.Passed)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed validation, got %d", stats.Failed)
	}
}

// mockValidator implements the Validator interface for testing
type mockValidator struct {
	archUnitPass bool
	linterPass   bool
	buildPass    bool
}

func (m *mockValidator) runArchUnitCheck(ctx context.Context) error {
	if !m.archUnitPass {
		return fmt.Errorf("arch-unit check failed")
	}
	return nil
}

func (m *mockValidator) runLinter(ctx context.Context, language types.Language) error {
	if !m.linterPass {
		return fmt.Errorf("linter failed")
	}
	return nil
}

func (m *mockValidator) runBuild(ctx context.Context) error {
	if !m.buildPass {
		return fmt.Errorf("build failed")
	}
	return nil
}
