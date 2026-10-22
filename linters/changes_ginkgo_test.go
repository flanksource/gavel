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
})

type scopeTestLinter struct{ dryRunLinter }

func (scopeTestLinter) Run(commonsContext.Context, *clicky.Task) ([]models.Violation, error) {
	return []models.Violation{{File: "other.go", Severity: models.SeverityError}}, nil
}
