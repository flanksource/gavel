package lifecycle

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/cel-go/cel"
	celast "github.com/google/cel-go/common/ast"
)

// Variables every predicate can read.
const (
	VarSubject   = "subject"
	VarRuns      = "runs"
	VarLast      = "last"
	VarRun       = "run"
	VarVerify    = "verify"
	VarEnvelope  = "envelope"
	VarPlan      = "plan"
	VarQuestions = "questions"
)

// Env compiles the lifecycle's predicates. It is strict on purpose: a name no
// declaration covers is a compile error, and a predicate that does not yield a
// bool is an evaluation error — a status transition must never rest on a value
// that quietly evaluated to "not set".
type Env struct {
	subject map[string]string
	when    *cel.Env
	outcome *cel.Env
}

// NewEnv builds the two environments from the subject declarations: the
// applicability env over {subject, runs, last}, and the outcome env that adds
// the finished run's facts.
func NewEnv(subject map[string]string) (*Env, error) {
	for name, typ := range subject {
		if _, err := celType(typ); err != nil {
			return nil, fmt.Errorf("subject %s: %w", name, err)
		}
	}
	dynMap := cel.MapType(cel.StringType, cel.DynType)
	whenVars := []cel.EnvOption{
		cel.Variable(VarSubject, dynMap),
		cel.Variable(VarRuns, cel.ListType(dynMap)),
		cel.Variable(VarLast, dynMap),
	}
	when, err := cel.NewEnv(whenVars...)
	if err != nil {
		return nil, fmt.Errorf("lifecycle when env: %w", err)
	}
	outcome, err := cel.NewEnv(append(whenVars,
		cel.Variable(VarRun, dynMap),
		cel.Variable(VarVerify, dynMap),
		cel.Variable(VarEnvelope, dynMap),
		cel.Variable(VarPlan, dynMap),
		cel.Variable(VarQuestions, cel.ListType(cel.DynType)),
	)...)
	if err != nil {
		return nil, fmt.Errorf("lifecycle outcome env: %w", err)
	}
	return &Env{subject: subject, when: when, outcome: outcome}, nil
}

// Program is one compiled predicate or input expression.
type Program struct {
	expr    string
	program cel.Program
}

// Expr is the source the program was compiled from.
func (p *Program) Expr() string { return p.expr }

// CompileWhen compiles an applicability predicate or an input expression over
// {subject, runs, last}.
func (e *Env) CompileWhen(expr string) (*Program, error) { return e.compile(e.when, expr) }

// CompileOutcome compiles an outcome predicate over the full variable set.
func (e *Env) CompileOutcome(expr string) (*Program, error) { return e.compile(e.outcome, expr) }

func (e *Env) compile(env *cel.Env, expr string) (*Program, error) {
	ast, issues := env.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("%q: %w", expr, issues.Err())
	}
	if err := e.checkSubjectFields(ast); err != nil {
		return nil, fmt.Errorf("%q: %w", expr, err)
	}
	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", expr, err)
	}
	return &Program{expr: expr, program: program}, nil
}

// checkSubjectFields rejects `subject.<field>` for a field the lifecycle does
// not declare. The subject is a dyn map to CEL, which would otherwise resolve a
// typo to a "no such key" error only when the predicate runs — and only on the
// todo whose state happened to reach it.
func (e *Env) checkSubjectFields(ast *cel.Ast) error {
	var undeclared []string
	visitor := celast.NewExprVisitor(func(expr celast.Expr) {
		if expr.Kind() != celast.SelectKind {
			return
		}
		sel := expr.AsSelect()
		operand := sel.Operand()
		if operand.Kind() != celast.IdentKind || operand.AsIdent() != VarSubject {
			return
		}
		if _, ok := e.subject[sel.FieldName()]; !ok {
			undeclared = append(undeclared, sel.FieldName())
		}
	})
	celast.PreOrderVisit(ast.NativeRep().Expr(), visitor)
	if len(undeclared) == 0 {
		return nil
	}
	sort.Strings(undeclared)
	declared := make([]string, 0, len(e.subject))
	for name := range e.subject {
		declared = append(declared, name)
	}
	sort.Strings(declared)
	return fmt.Errorf("subject.%s is not declared (declared: %s)",
		strings.Join(undeclared, ", subject."), strings.Join(declared, ", "))
}

// Bool evaluates a predicate. Anything but a bool result is an error.
func (p *Program) Bool(vars map[string]any) (bool, error) {
	out, _, err := p.program.Eval(vars)
	if err != nil {
		return false, fmt.Errorf("%q: %w", p.expr, err)
	}
	value, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("%q did not return a bool: got %T(%v)", p.expr, out.Value(), out.Value())
	}
	return value, nil
}

// Value evaluates an input expression to a native Go value.
func (p *Program) Value(vars map[string]any) (any, error) {
	out, _, err := p.program.Eval(vars)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", p.expr, err)
	}
	native, err := out.ConvertToNative(reflect.TypeOf((*any)(nil)).Elem())
	if err != nil {
		return nil, fmt.Errorf("%q: %w", p.expr, err)
	}
	return native, nil
}

// CheckSubject verifies a host-built subject against the declarations: every
// declared field present with a value of the declared shape. A host that drifts
// from the definition fails here, before any predicate reads a missing key.
func (e *Env) CheckSubject(subject map[string]any) error {
	var missing []string
	for name, typ := range e.subject {
		value, ok := subject[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		if err := checkSubjectValue(typ, value); err != nil {
			return fmt.Errorf("subject.%s: %w", name, err)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("subject is missing declared field(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

func checkSubjectValue(typ string, value any) error {
	kind := reflect.Invalid
	if value != nil {
		kind = reflect.TypeOf(value).Kind()
	}
	switch typ {
	case "string":
		if kind != reflect.String {
			return fmt.Errorf("declared string, got %T", value)
		}
	case "bool":
		if kind != reflect.Bool {
			return fmt.Errorf("declared bool, got %T", value)
		}
	case "int":
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		default:
			return fmt.Errorf("declared int, got %T", value)
		}
	case "double":
		if kind != reflect.Float32 && kind != reflect.Float64 {
			return fmt.Errorf("declared double, got %T", value)
		}
	case "list<string>", "list<dyn>":
		if kind != reflect.Slice && kind != reflect.Array {
			return fmt.Errorf("declared %s, got %T", typ, value)
		}
	case "map<string,dyn>":
		if kind != reflect.Map {
			return fmt.Errorf("declared %s, got %T", typ, value)
		}
	}
	return nil
}

// celType validates a subject type name. The subject itself is handed to CEL
// as a dyn map — the declaration's job is documentation, the field check above,
// and the value check at evaluation.
func celType(name string) (*cel.Type, error) {
	switch strings.ReplaceAll(strings.TrimSpace(name), " ", "") {
	case "string":
		return cel.StringType, nil
	case "bool":
		return cel.BoolType, nil
	case "int":
		return cel.IntType, nil
	case "double":
		return cel.DoubleType, nil
	case "dyn":
		return cel.DynType, nil
	case "list<string>":
		return cel.ListType(cel.StringType), nil
	case "list<dyn>":
		return cel.ListType(cel.DynType), nil
	case "map<string,dyn>":
		return cel.MapType(cel.StringType, cel.DynType), nil
	}
	return nil, fmt.Errorf("type %q is not one of %s", name, strings.Join(SubjectTypes, ", "))
}
