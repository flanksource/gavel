// Package linters provides integration with external linters for comprehensive code quality analysis.
//
// The linters package enables:
//   - Running multiple linters (golangci-lint, ruff, eslint, etc.)
//   - Consolidated violation reporting across all tools
//   - Language-specific linter configuration
//   - Parallel linter execution with task coordination
//
// # Supported Linters
//
//   - arch-unit: Architecture dependency rules (built-in)
//   - golangci-lint: Comprehensive Go linter
//   - ruff: Fast Python linter and formatter
//   - pyright: Python type checker
//   - eslint: JavaScript/TypeScript linter
//   - markdownlint: Markdown style checker
//   - vale: Prose and documentation linter
//
// # Configuration
//
// Linters are configured in arch-unit.yaml:
//
//	linters:
//	  golangci-lint:
//	    enabled: true
//	    languages: ["go"]
//	    args: ["--config=.golangci.yml"]
//	  ruff:
//	    enabled: true
//	    languages: ["python"]
//	    args: ["--select=E,W,F"]
//
// # Programmatic Usage
//
// Run linters programmatically:
//
//	config := linters.DefaultConfig()
//	config.Linters["golangci-lint"].Enabled = true
//
//	runner := linters.NewRunner(config)
//	results, _ := runner.RunLinters(ctx, []string{"/path/to/code"})
//
//	for _, result := range results {
//	    fmt.Printf("%s: %d violations\n", result.Linter, len(result.Violations))
//	}
//
// # Consolidated Results
//
// Linter results are consolidated with architecture violations:
//
//	consolidated := models.NewConsolidatedResult(archResult, linterResults)
//	fmt.Printf("Total violations: %d\n", consolidated.Summary.TotalViolations)
//	fmt.Printf("  Architecture: %d\n", consolidated.Summary.ArchViolations)
//	fmt.Printf("  Linters: %d\n", consolidated.Summary.LinterViolations)
//
// # Custom Linters
//
// Add custom linters by implementing the Linter interface:
//
//	type MyLinter struct {}
//
//	func (l *MyLinter) Name() string { return "my-linter" }
//
//	func (l *MyLinter) Run(ctx context.Context, paths []string) (*LinterResult, error) {
//	    // Run linter and return results
//	    return &LinterResult{
//	        Linter:     "my-linter",
//	        Success:    true,
//	        Violations: violations,
//	    }, nil
//	}
//
//	// Register custom linter
//	linters.Register("my-linter", &MyLinter{})
//
// # Parallel Execution
//
// Linters run in parallel using the clicky task framework:
//   - Automatic timeout handling
//   - Progress reporting
//   - Error recovery
//   - Resource management
//
// # Violation Mapping
//
// Linter output is parsed and mapped to standard violation format:
//
//	type Violation struct {
//	    File     string // Source file path
//	    Line     int    // Line number
//	    Column   int    // Column number
//	    Message  string // Violation message
//	    Source   string // Linter name
//	    Severity string // Severity level
//	}
//
// See also:
//   - github.com/flanksource/gavel/config for configuration
//   - github.com/flanksource/gavel/models for result structures
package linters
