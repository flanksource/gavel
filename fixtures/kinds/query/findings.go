package query

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gomplate/v3"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// findingsSource is Violation.Source and the fingerprint's linter name.
const findingsSource = "query"

// findingsSpec is a step's `findings:`: a CEL expression returning a list of
// violation maps, or a per-row mapper.
type findingsSpec struct {
	expression string
	mapper     *rowMapper
}

// rowMapper maps each row where When holds to one violation. String fields
// are Go templates over .row (and the step's CEL roots); Severity is a
// literal severity or a CEL expression returning one.
type rowMapper struct {
	When      string `json:"when,omitempty"`
	Severity  string `json:"severity"`
	Rule      string `json:"rule,omitempty"`
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	Locations []any  `json:"locations,omitempty"`
}

// finding is the violation shape a findings element decodes from.
type finding struct {
	Severity  models.ViolationSeverity   `json:"severity"`
	Rule      string                     `json:"rule,omitempty"`
	Message   string                     `json:"message"`
	Code      string                     `json:"code,omitempty"`
	Locations []models.ViolationLocation `json:"locations,omitempty"`
}

func parseFindings(raw json.RawMessage) (*findingsSpec, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var expression string
	if err := json.Unmarshal(raw, &expression); err == nil {
		if strings.TrimSpace(expression) == "" {
			return nil, errors.New("findings expression is empty")
		}
		return &findingsSpec{expression: expression}, nil
	}
	mapper := &rowMapper{}
	if err := decodeStrict(raw, mapper); err != nil {
		return nil, fmt.Errorf("findings must be a CEL expression or a row mapper: %w", err)
	}
	if mapper.Severity == "" || mapper.Message == "" {
		return nil, errors.New("findings row mapper needs severity and message")
	}
	return &findingsSpec{mapper: mapper}, nil
}

// evaluate is the step's violations, fingerprinted; nil without findings.
func (f *findingsSpec) evaluate(step *fixtures.StepContext, outputs map[string]any) ([]models.Violation, error) {
	if f == nil {
		return nil, nil
	}
	var elements []map[string]any
	var err error
	if f.mapper != nil {
		elements, err = f.mapper.apply(step, outputs)
	} else {
		elements, err = evalFindingsExpression(f.expression, step.CELVars(outputs))
	}
	if err != nil {
		return nil, err
	}
	violations := make([]models.Violation, 0, len(elements))
	for i, element := range elements {
		violation, err := decodeFinding(element, stepRule(step))
		if err != nil {
			return nil, fmt.Errorf("findings[%d]: %w", i, err)
		}
		violations = append(violations, violation)
	}
	linters.AssignFingerprints(&linters.LinterResult{Linter: findingsSource, WorkDir: step.WorkDir, Violations: violations})
	return violations, nil
}

func evalFindingsExpression(expression string, vars map[string]any) ([]map[string]any, error) {
	output, err := gomplate.RunExpression(vars, gomplate.Template{Expression: expression, CelEnvs: fixtures.ANSICelFunctions()})
	if err != nil {
		return nil, fmt.Errorf("findings %s: %w", expression, err)
	}
	plain, err := fixtures.PlainValue(output)
	if err != nil {
		return nil, err
	}
	list, ok := plain.([]any)
	if !ok {
		return nil, fmt.Errorf("findings %s returned %T, want a list of maps", expression, plain)
	}
	elements := make([]map[string]any, len(list))
	for i, item := range list {
		if elements[i], ok = item.(map[string]any); !ok {
			return nil, fmt.Errorf("findings[%d] is %T, want a map", i, item)
		}
	}
	return elements, nil
}

// apply renders one violation map per row where When holds, over the rows of
// every result set in order; resultSet is the 0-based index of a row's set.
func (m *rowMapper) apply(step *fixtures.StepContext, outputs map[string]any) ([]map[string]any, error) {
	sets, _ := outputs["resultSets"].([]any)
	var elements []map[string]any
	for s, set := range sets {
		rows, _ := set.([]any)
		for i, row := range rows {
			vars := step.CELVars(outputs)
			vars["row"], vars["resultSet"] = row, s
			element, keep, err := m.applyRow(step, vars)
			if err != nil {
				return nil, fmt.Errorf("findings result set %d row %d: %w", s, i, err)
			}
			if keep {
				elements = append(elements, element)
			}
		}
	}
	return elements, nil
}

func (m *rowMapper) applyRow(step *fixtures.StepContext, vars map[string]any) (map[string]any, bool, error) {
	if m.When != "" {
		keep, err := evalBool(m.When, vars)
		if err != nil || !keep {
			return nil, false, err
		}
	}
	severity, err := m.severity(vars)
	if err != nil {
		return nil, false, err
	}
	render := func(text string) (string, error) { return step.Scope.Render(text, vars) }
	element := map[string]any{"severity": severity}
	for key, text := range map[string]string{"rule": m.Rule, "message": m.Message, "code": m.Code} {
		if text == "" {
			continue
		}
		if element[key], err = render(text); err != nil {
			return nil, false, fmt.Errorf("%s: %w", key, err)
		}
	}
	locations, err := renderLeaves(m.Locations, render)
	if err != nil {
		return nil, false, fmt.Errorf("locations: %w", err)
	}
	if locations != nil {
		element["locations"] = locations
	}
	return element, true, nil
}

// severity is the literal severity, else the CEL expression's result.
func (m *rowMapper) severity(vars map[string]any) (string, error) {
	switch models.ViolationSeverity(m.Severity) {
	case models.SeverityError, models.SeverityWarning, models.SeverityInfo:
		return m.Severity, nil
	}
	output, err := gomplate.RunExpression(vars, gomplate.Template{Expression: m.Severity, CelEnvs: fixtures.ANSICelFunctions()})
	if err != nil {
		return "", fmt.Errorf("severity %s: %w", m.Severity, err)
	}
	severity, ok := output.(string)
	if !ok {
		return "", fmt.Errorf("severity %s returned %T, want error, warning or info", m.Severity, output)
	}
	return severity, nil
}

// renderLeaves renders every string inside value, keeping its shape.
func renderLeaves(value any, render func(string) (string, error)) (any, error) {
	switch v := value.(type) {
	case string:
		return render(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			rendered, err := renderLeaves(item, render)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			out[i] = rendered
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			rendered, err := renderLeaves(item, render)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}
			out[key] = rendered
		}
		return out, nil
	}
	return value, nil
}

func decodeFinding(element map[string]any, defaultRule string) (models.Violation, error) {
	raw, err := json.Marshal(element)
	if err != nil {
		return models.Violation{}, err
	}
	var f finding
	if err := decodeStrict(raw, &f); err != nil {
		return models.Violation{}, err
	}
	switch f.Severity {
	case models.SeverityError, models.SeverityWarning, models.SeverityInfo:
	default:
		return models.Violation{}, fmt.Errorf("unknown severity %q: want error, warning or info", f.Severity)
	}
	if strings.TrimSpace(f.Message) == "" {
		return models.Violation{}, errors.New("message is required")
	}
	for i, location := range f.Locations {
		if err := location.Validate(); err != nil {
			return models.Violation{}, fmt.Errorf("locations[%d]: %w", i, err)
		}
	}
	return f.violation(defaultRule), nil
}

func (f finding) violation(defaultRule string) models.Violation {
	rule := f.Rule
	if rule == "" {
		rule = defaultRule
	}
	v := models.Violation{
		Severity: f.Severity, Message: models.StringPtr(f.Message), Source: findingsSource,
		Rule: &models.Rule{Method: rule}, Locations: f.Locations,
	}
	if f.Code != "" {
		v.Code = models.StringPtr(f.Code)
	}
	for _, location := range f.Locations {
		if location.File != "" {
			v.File, v.Line, v.Column, v.EndLine = location.File, location.Line, location.Column, location.EndLine
			break
		}
	}
	return v
}

func stepRule(step *fixtures.StepContext) string {
	if step.Spec.ID != "" {
		return step.Spec.ID
	}
	return stepName(step)
}

// findingsResult is the step's result: each violation a child, any error
// failing the step and any warning or info warning it.
func findingsResult(step *fixtures.StepContext, violations []models.Violation) fixtures.FixtureResult {
	result := fixtures.FixtureResult{Name: stepName(step), Type: "query"}
	var errorCount, warnCount int
	for _, v := range violations {
		switch v.Severity {
		case models.SeverityError:
			errorCount++
		default:
			warnCount++
		}
		result.Children = append(result.Children, fixtures.ViolationNode(v))
	}
	result.Status = fixtures.FindingsStatus(errorCount, warnCount)
	switch result.Status {
	case task.StatusFAIL:
		result.Error = fmt.Sprintf("%d error finding(s), %d other", errorCount, warnCount)
	case task.StatusWarning:
		result.Error = fmt.Sprintf("%d warning/info finding(s)", warnCount)
	}
	return result
}

func evalBool(expression string, vars map[string]any) (bool, error) {
	output, err := gomplate.RunExpression(vars, gomplate.Template{Expression: expression, CelEnvs: fixtures.ANSICelFunctions()})
	if err != nil {
		return false, fmt.Errorf("%s: %w", expression, err)
	}
	value, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("%s returned %T(%v), want a boolean", expression, output, output)
	}
	return value, nil
}
