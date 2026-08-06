package fixtures

import (
	"strings"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gomplate/v3"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture CEL failure traces", func() {
	evaluate := func(expression string, values map[string]any) FixtureResult {
		return (Expectations{CEL: expression}).Evaluate(
			FixtureResult{Name: "CEL assertion"},
			exec.ExecResult{},
			EvaluateOptions{CELVars: values},
		)
	}

	It("prints every evaluated operand without side labels", func() {
		result := evaluate("a == b && c > d", map[string]any{
			"a": 1,
			"b": 1,
			"c": 3,
			"d": 4,
		})

		Expect(result.Status).To(Equal(task.StatusFAIL))
		Expect(result.CELTrace).To(Equal(strings.Join([]string{
			"cel: a == b && c > d",
			"     │    │    │   │",
			"     │    │    │   └─ int(4)",
			"     │    │    └─ int(3)",
			"     │    └─ int(1)",
			"     └─ int(1)",
		}, "\n")))
		Expect(result.CELTrace).NotTo(ContainSubstring("left"))
		Expect(result.CELTrace).NotTo(ContainSubstring("right"))
	})

	It("omits operands skipped by CEL short-circuiting", func() {
		result := evaluate("a == b && c > d", map[string]any{
			"a": 1,
			"b": 2,
			"c": 3,
			"d": 4,
		})

		Expect(result.CELTrace).To(Equal(strings.Join([]string{
			"cel: a == b && c > d",
			"     │    │",
			"     │    └─ int(2)",
			"     └─ int(1)",
		}, "\n")))
	})

	It("prefers an indexed value over overlapping container values", func() {
		result := evaluate("payload.items[0] == expected", map[string]any{
			"payload":  map[string]any{"items": []any{1}},
			"expected": 2,
		})

		Expect(strings.Split(result.CELTrace, "\n")).To(Equal([]string{
			"cel: payload.items[0] == expected",
			"                  │      │",
			"                  │      └─ int(2)",
			"                  └─ int(1)",
		}))
	})

	It("groups values beneath their multiline source line", func() {
		result := evaluate("a == b &&\n  c > d", map[string]any{
			"a": 1,
			"b": 1,
			"c": 3,
			"d": 4,
		})

		Expect(result.CELTrace).To(Equal(strings.Join([]string{
			"cel: a == b &&",
			"     │    │",
			"     │    └─ int(1)",
			"     └─ int(1)",
			"       c > d",
			"       │   │",
			"       │   └─ int(4)",
			"       └─ int(3)",
		}, "\n")))
	})

	It("retains partial values for runtime errors", func() {
		result := evaluate("numerator / denominator == expected", map[string]any{
			"numerator":   8,
			"denominator": 0,
			"expected":    2,
		})

		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.CELTrace).To(ContainSubstring("int(8)"))
		Expect(result.CELTrace).To(ContainSubstring("int(0)"))
		Expect(result.CELTrace).To(ContainSubstring("int(2)"))
	})

	It("does not track successful assertions", func() {
		result := evaluate("actual == expected", map[string]any{"actual": 1, "expected": 1})

		Expect(result.Status).NotTo(Equal(task.StatusFAIL))
		Expect(result.CELTrace).To(BeEmpty())
	})

	It("bounds each rendered value", func() {
		result := evaluate("actual == expected", map[string]any{
			"actual":   strings.Repeat("a", 250),
			"expected": "short",
		})

		Expect(result.CELTrace).To(ContainSubstring("…"))
		Expect(result.CELTrace).NotTo(ContainSubstring(strings.Repeat("a", 201)))
	})

	It("includes native function results", func() {
		result := evaluate("size(items) == expected", map[string]any{
			"items":    []any{1, 2},
			"expected": 3,
		})

		Expect(result.CELTrace).To(ContainSubstring("int(2)"))
		Expect(result.CELTrace).To(ContainSubstring("list([1 2])"))
		Expect(result.CELTrace).To(ContainSubstring("int(3)"))
	})

	It("discards a retry that no longer reproduces the failure", func() {
		changed := cel.Function("changed",
			cel.Overload("changed_bool", []*cel.Type{}, cel.BoolType,
				cel.FunctionBinding(func(...ref.Val) ref.Val { return types.True })))

		trace := traceCELFailure("changed()", nil, gomplate.Template{
			Expression: "changed()",
			CelEnvs:    []cel.EnvOption{changed},
		}, celFailureFalse)

		Expect(trace).To(BeEmpty())
	})

	DescribeTable("requires the retry to reproduce the same failure kind",
		func(kind celFailureKind, output any, err error, reproduced bool) {
			Expect(celFailureReproduced(kind, output, err)).To(Equal(reproduced))
		},
		Entry("false remains false", celFailureFalse, false, nil, true),
		Entry("false becomes true", celFailureFalse, true, nil, false),
		Entry("false becomes an error", celFailureFalse, nil, assertiveError("changed"), false),
		Entry("error remains an error", celFailureError, nil, assertiveError("again"), true),
		Entry("error becomes false", celFailureError, false, nil, false),
	)
})

type assertiveError string

func (e assertiveError) Error() string { return string(e) }
