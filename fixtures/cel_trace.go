package fixtures

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gomplate/v3"
	celast "github.com/google/cel-go/common/ast"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types/ref"
)

type celFailureKind int

const (
	celFailureFalse celFailureKind = iota
	celFailureError
)

type celTraceValue struct {
	line, column int
	start, stop  int32
	priority     int
	value        ref.Val
}

func traceCELFailure(expression string, values map[string]any, template gomplate.Template, kind celFailureKind) string {
	tracker := gomplate.NewCELTracker()
	ctx := gomplate.WithCELTracker(commonsContext.NewContext(context.Background()), tracker)
	output, err := gomplate.RunExpressionContext(ctx, values, template)
	if !celFailureReproduced(kind, output, err) {
		return ""
	}
	return renderCELTrace(expression, tracker.Snapshot())
}

func celFailureReproduced(kind celFailureKind, output any, err error) bool {
	if kind == celFailureError {
		return err != nil
	}
	value, ok := output.(bool)
	return err == nil && ok && !value
}

func renderCELTrace(expression string, snapshot gomplate.CELTraceSnapshot) string {
	values := collectCELTraceValues(snapshot)
	if len(values) == 0 {
		return ""
	}

	var rendered []string
	for index, line := range strings.Split(expression, "\n") {
		prefix := "     "
		if index == 0 {
			prefix = "cel: "
		}
		rendered = append(rendered, prefix+line)
		rendered = append(rendered, renderCELTraceLine(prefix, valuesForCELLine(values, index+1))...)
	}
	return strings.Join(rendered, "\n")
}

func collectCELTraceValues(snapshot gomplate.CELTraceSnapshot) []celTraceValue {
	if snapshot.AST == nil || snapshot.Details == nil || snapshot.Details.State() == nil {
		return nil
	}

	nativeAST := snapshot.AST.NativeRep()
	state := snapshot.Details.State()
	suppressed := suppressedCELTraceIDs(nativeAST.Expr())
	var values []celTraceValue
	celast.PostOrderVisit(nativeAST.Expr(), celast.NewExprVisitor(func(expr celast.Expr) {
		if suppressed[expr.ID()] {
			return
		}
		value, found := state.Value(expr.ID())
		offset, sourced := nativeAST.SourceInfo().GetOffsetRange(expr.ID())
		priority, useful := celTracePriority(expr)
		if !found || !sourced || !useful || value == nil {
			return
		}
		location := nativeAST.SourceInfo().GetStartLocation(expr.ID())
		stop := max(offset.Stop, offset.Start+1)
		values = append(values, celTraceValue{
			line: location.Line(), column: location.Column(), start: offset.Start, stop: stop,
			priority: priority, value: value,
		})
	}))
	return deduplicateCELTraceValues(values)
}

func suppressedCELTraceIDs(root celast.Expr) map[int64]bool {
	suppressed := map[int64]bool{}
	celast.PostOrderVisit(root, celast.NewExprVisitor(func(expr celast.Expr) {
		switch expr.Kind() {
		case celast.SelectKind:
			markCELTraceDescendants(expr.AsSelect().Operand(), suppressed)
		case celast.CallKind:
			name := expr.AsCall().FunctionName()
			if name == operators.Index || name == operators.OptIndex || name == operators.OptSelect {
				for _, argument := range expr.AsCall().Args() {
					markCELTraceDescendants(argument, suppressed)
				}
			}
		}
	}))
	return suppressed
}

func markCELTraceDescendants(root celast.Expr, suppressed map[int64]bool) {
	celast.PostOrderVisit(root, celast.NewExprVisitor(func(expr celast.Expr) {
		suppressed[expr.ID()] = true
	}))
}

func celTracePriority(expr celast.Expr) (int, bool) {
	switch expr.Kind() {
	case celast.IdentKind, celast.LiteralKind:
		return 1, true
	case celast.SelectKind:
		return 2, true
	case celast.CallKind:
		name := expr.AsCall().FunctionName()
		if name == operators.Index || name == operators.OptIndex || name == operators.OptSelect {
			return 3, true
		}
		_, structural := operators.FindReverse(name)
		return 1, !structural
	default:
		return 0, false
	}
}

func deduplicateCELTraceValues(values []celTraceValue) []celTraceValue {
	result := make([]celTraceValue, 0, len(values))
	for i, candidate := range values {
		discard := false
		for j, other := range values {
			if i != j && overlapsCELRange(candidate, other) && other.priority > candidate.priority {
				discard = true
				break
			}
		}
		if !discard && !hasCELTracePosition(result, candidate.line, candidate.column) {
			result = append(result, candidate)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].line < result[j].line || result[i].line == result[j].line && result[i].column < result[j].column
	})
	return result
}

func overlapsCELRange(a, b celTraceValue) bool {
	return a.start < b.stop && b.start < a.stop
}

func hasCELTracePosition(values []celTraceValue, line, column int) bool {
	for _, value := range values {
		if value.line == line && value.column == column {
			return true
		}
	}
	return false
}

func valuesForCELLine(values []celTraceValue, line int) []celTraceValue {
	var result []celTraceValue
	for _, value := range values {
		if value.line == line {
			result = append(result, value)
		}
	}
	return result
}

func renderCELTraceLine(prefix string, values []celTraceValue) []string {
	if len(values) == 0 {
		return nil
	}
	width := values[len(values)-1].column + 1
	guides := make([]rune, width)
	fillRunes(guides, ' ')
	for _, value := range values {
		guides[value.column] = '│'
	}
	lines := []string{strings.Repeat(" ", len(prefix)) + string(guides)}
	for index := len(values) - 1; index >= 0; index-- {
		detail := make([]rune, values[index].column)
		fillRunes(detail, ' ')
		for prior := 0; prior < index; prior++ {
			detail[values[prior].column] = '│'
		}
		lines = append(lines, strings.Repeat(" ", len(prefix))+string(detail)+"└─ "+formatCELTraceValue(values[index].value))
	}
	return lines
}

func fillRunes(values []rune, value rune) {
	for index := range values {
		values[index] = value
	}
}

func formatCELTraceValue(value ref.Val) string {
	var rendered string
	switch native := value.Value().(type) {
	case string:
		rendered = strconv.Quote(truncateCELTraceValue(native))
	case []byte:
		rendered = strconv.Quote(truncateCELTraceValue(string(native)))
	default:
		rendered = truncateCELTraceValue(fmt.Sprintf("%v", native))
	}
	return fmt.Sprintf("%s(%s)", value.Type().TypeName(), rendered)
}

func truncateCELTraceValue(value string) string {
	const maxValueRunes = 200
	runes := []rune(value)
	if len(runes) <= maxValueRunes {
		return value
	}
	return string(runes[:maxValueRunes-1]) + "…"
}
