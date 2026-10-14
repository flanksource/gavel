package fixtures

import (
	"fmt"
	"strings"

	"github.com/flanksource/gomplate/v3"
)

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

// stepCELVars is the environment a step's own expressions see: its outputs as
// top-level roots, then the reserved roots `outputs`, `env` and `steps`, which
// win over an output of the same name (reach those through `outputs.<name>`).
func stepCELVars(outputs map[string]any, scope *StepScope) map[string]any {
	vars := make(map[string]any, len(outputs)+3)
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
