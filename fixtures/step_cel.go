package fixtures

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/flanksource/gomplate/v3"
)

// PlainValue converts a typed value to the maps and slices CEL navigates. Step
// outputs must hold only such plain values: a struct left in them is rewritten
// by CEL evaluation, dropping its own JSON encoding (durations become integers).
func PlainValue(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode %T for CEL: %w", value, err)
	}
	var plain any
	if err := json.Unmarshal(raw, &plain); err != nil {
		return nil, fmt.Errorf("decode %T for CEL: %w", value, err)
	}
	return plain, nil
}

// evalStepCEL evaluates a boolean CEL expression of a step — its `cel:`, a
// wait's until/failWhen, a retry's when — through the same gomplate CEL
// environment an exec expectation uses. A false or failed evaluation carries
// the rendered trace of the values the expression saw.
func evalStepCEL(expression string, vars map[string]any) (bool, string, error) {
	template := gomplate.Template{Expression: expression, CelEnvs: ANSICelFunctions()}
	output, err := gomplate.RunExpression(vars, template)
	if err != nil {
		return false, traceCELFailure(expression, vars, template, celFailureError), err
	}
	switch v := output.(type) {
	case bool:
		if !v {
			return false, traceCELFailure(expression, vars, template, celFailureFalse), nil
		}
		return true, "", nil
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true":
			return true, "", nil
		case "false":
			return false, "", nil
		}
	}
	return false, "", fmt.Errorf("%s did not return a boolean: got %T(%v)", expression, output, output)
}

// CELVars is the CEL environment this step's own expressions evaluate against,
// given its outputs: what `cel:`, `if:` and a kind's own expressions (a query's
// `findings:`) all see.
func (s *StepContext) CELVars(outputs map[string]any) map[string]any {
	return stepCELVars(outputs, s.Scope)
}

// ReservedRoots are the CEL and template roots a step id may not take: a step
// with one of these ids is a load error.
var ReservedRoots = []string{"outputs", "env", "steps", "fixture", "tmp"}

// rootIdentifier is what a step id must look like to be bound as a root: a
// CEL identifier, which is also a Go template field name.
var rootIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// celKeywords cannot name a CEL root even though they match rootIdentifier.
var celKeywords = []string{"true", "false", "null", "in", "as", "break", "const", "continue", "else", "for",
	"function", "if", "import", "let", "loop", "package", "namespace", "return", "var", "void", "while"}

// IsRootID reports whether a step id is bound as a CEL and template root.
func IsRootID(id string) bool {
	return rootIdentifier.MatchString(id) && !slices.Contains(celKeywords, id) && !slices.Contains(ReservedRoots, id)
}

// stepIDRoots binds every completed step's outputs under its id, for ids that
// can be a root (IsRootID); the rest stay reachable through steps.<id>.
func stepIDRoots(scope *StepScope) map[string]any {
	roots := map[string]any{}
	for id, outputs := range scope.StepOutputs() {
		if IsRootID(id) {
			roots[id] = outputs
		}
	}
	return roots
}

// stepCELVars is the environment a step's own expressions see: every earlier
// step's outputs under its id, its own outputs as top-level roots over them,
// then the reserved roots `outputs`, `env` and `steps`, which win over both
// (reach a shadowed output through `outputs.<name>`).
func stepCELVars(outputs map[string]any, scope *StepScope) map[string]any {
	vars := stepIDRoots(scope)
	for k, v := range outputs {
		vars[k] = v
	}
	if outputs == nil {
		outputs = map[string]any{}
	}
	vars["outputs"] = outputs
	vars["env"] = scope.Env.Map()
	vars["steps"] = scope.StepOutputs()
	return vars
}
