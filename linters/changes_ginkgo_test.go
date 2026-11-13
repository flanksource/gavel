package linters

import (
	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Changed-file task verdicts", func() {
	It("drops out-of-scope errors before setting the task status", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opts := RunOptions{WorkDir: GinkgoT().TempDir(), Scopes: []string{"selected.go"}}
		task := clicky.StartTask[*LinterResult]("scoped findings", func(ctx commonsContext.Context, task *clicky.Task) (*LinterResult, error) {
			return RunLinterWithTask(ctx, task, scopeTestLinter{}, opts), nil
		})
		result, err := task.GetResult()
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Violations).To(BeEmpty())
		Expect(task.IsOk()).To(BeTrue())
	})

	DescribeTable("passes a task whose findings are all info, and warns on any above it",
		func(severities []models.ViolationSeverity, ok bool) {
			GinkgoT().Setenv("HOME", GinkgoT().TempDir())
			opts := RunOptions{WorkDir: GinkgoT().TempDir()}
			task := clicky.StartTask[*LinterResult]("severity findings", func(ctx commonsContext.Context, task *clicky.Task) (*LinterResult, error) {
				return RunLinterWithTask(ctx, task, severityTestLinter{severities: severities}, opts), nil
			})
			result, err := task.GetResult()
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Violations).To(HaveLen(len(severities)))
			Expect(task.IsOk()).To(Equal(ok))
		},
		Entry("info only", []models.ViolationSeverity{models.SeverityInfo, models.SeverityInfo}, true),
		Entry("info and a warning", []models.ViolationSeverity{models.SeverityInfo, models.SeverityWarning}, false),
	)
})

type severityTestLinter struct {
	dryRunLinter
	severities []models.ViolationSeverity
}

func (l severityTestLinter) Run(commonsContext.Context, *clicky.Task) ([]models.Violation, error) {
	var out []models.Violation
	for _, severity := range l.severities {
		out = append(out, models.Violation{File: "rule.xml", Severity: severity})
	}
	return out, nil
}

type scopeTestLinter struct{ dryRunLinter }

func (scopeTestLinter) Run(commonsContext.Context, *clicky.Task) ([]models.Violation, error) {
	return []models.Violation{{File: "other.go", Severity: models.SeverityError}}, nil
}
