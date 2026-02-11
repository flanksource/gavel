package rules

import (
	"fmt"

	"github.com/flanksource/gavel/models"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Engine evaluates CEL-based severity rules against change contexts
type Engine struct {
	config        *SeverityConfig
	celEnv        *cel.Env
	compiledRules []compiledRule
}

type compiledRule struct {
	expression string
	program    cel.Program
	severity   models.Severity
}

// NewEngine creates a new severity evaluation engine
func NewEngine(config *SeverityConfig) (*Engine, error) {
	if config == nil {
		config = DefaultSeverityConfig()
	}

	// Create CEL environment with all context fields declared
	env, err := cel.NewEnv(
		cel.Variable("commit", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("change", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("kubernetes", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("file", cel.MapType(cel.StringType, cel.AnyType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	engine := &Engine{
		config:        config,
		celEnv:        env,
		compiledRules: make([]compiledRule, 0, len(config.Rules)),
	}

	// Compile all rules
	for expr, severity := range config.Rules {
		program, err := engine.compileExpression(expr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile rule '%s': %w", expr, err)
		}

		engine.compiledRules = append(engine.compiledRules, compiledRule{
			expression: expr,
			program:    program,
			severity:   severity,
		})
	}

	return engine, nil
}

// compileExpression compiles a CEL expression into a program
func (e *Engine) compileExpression(expr string) (cel.Program, error) {
	ast, issues := e.celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("CEL compilation error: %w", issues.Err())
	}

	program, err := e.celEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("CEL program creation error: %w", err)
	}

	return program, nil
}

// Evaluate runs all rules against the context and returns the first matching severity
// If no rules match, returns the default severity
func (e *Engine) Evaluate(ctx map[string]any) models.Severity {
	// Evaluate rules in order (first match wins)
	for _, rule := range e.compiledRules {
		result, _, err := rule.program.Eval(ctx)
		if err != nil {
			// Log error but continue with other rules
			continue
		}

		// Check if result is true
		if boolVal, ok := result.(types.Bool); ok && boolVal == types.True {
			return rule.severity
		}
	}

	// No rules matched, return default
	return e.config.Default
}

// EvaluateWithDetails evaluates rules and returns severity plus matched rule expression
func (e *Engine) EvaluateWithDetails(ctx map[string]any) (models.Severity, string, error) {
	for _, rule := range e.compiledRules {
		result, _, err := rule.program.Eval(ctx)
		if err != nil {
			continue
		}

		if boolVal, ok := result.(types.Bool); ok && boolVal == types.True {
			return rule.severity, rule.expression, nil
		}
	}

	return e.config.Default, "", nil
}

// TestExpression tests a single CEL expression against a context
// Useful for debugging and validation
func (e *Engine) TestExpression(expr string, ctx map[string]any) (bool, error) {
	program, err := e.compileExpression(expr)
	if err != nil {
		return false, err
	}

	result, _, err := program.Eval(ctx)
	if err != nil {
		return false, fmt.Errorf("evaluation error: %w", err)
	}

	boolVal, ok := result.(ref.Val)
	if !ok {
		return false, fmt.Errorf("expression did not return a boolean value")
	}

	if boolVal.Type().TypeName() != "bool" {
		return false, fmt.Errorf("expression returned %s, expected bool", boolVal.Type().TypeName())
	}

	return boolVal.Equal(types.True) == types.True, nil
}

// GetConfig returns the engine's configuration
func (e *Engine) GetConfig() *SeverityConfig {
	return e.config
}

// RuleCount returns the number of compiled rules
func (e *Engine) RuleCount() int {
	return len(e.compiledRules)
}
