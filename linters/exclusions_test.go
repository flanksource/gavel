package linters

import (
	"context"
	"errors"

	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule exclusions", func() {
	sources := map[string]string{"uir": "uir", "eslint": "eslint", "betterleaks": "betterleaks", "secrets": "betterleaks"}
	It("normalizes comma lists and aliases without duplicating selectors", func() {
		rules, err := ParseRuleExclusions([]string{" uir:small-* ,eslint:no-debugger", "secrets:token", "betterleaks:token"}, sources)
		Expect(err).NotTo(HaveOccurred())
		Expect(rules).To(Equal(RuleExclusions{{Source: "uir", Pattern: "small-*"}, {Source: "eslint", Pattern: "no-debugger"}, {Source: "betterleaks", Pattern: "token"}}))
	})
	DescribeTable("rejects malformed selectors", func(value string) {
		_, err := ParseRuleExclusions([]string{value}, sources)
		Expect(err).To(HaveOccurred())
	}, Entry("empty", ""), Entry("empty list entry", "one,,two"), Entry("empty source", ":one"), Entry("empty rule", "uir:"), Entry("unknown source", "missing:one"))
	It("matches case-sensitive full rule IDs while preserving paths and unrelated sources", func() {
		rules, err := ParseRuleExclusions([]string{"uir:small-*", "*/no-debugger"}, sources)
		Expect(err).NotTo(HaveOccurred())
		finding := func(source, rule string) models.Violation {
			return models.Violation{Source: source, File: "/project/sample.go", Rule: &models.Rule{Method: rule}, Severity: models.SeverityError}
		}
		keep := []models.Violation{finding("eslint", "small-function"), finding("uir", "Small-function"), {File: "/project/diagnostic.go"}}
		all := append([]models.Violation{finding("uir", "small-function"), finding("eslint", "eslint/no-debugger")}, keep...)
		Expect(rules.Filter("uir", all)).To(Equal(keep))
	})
	It("routes qualified rules only to their source and bare rules only to a single selected linter", func() {
		rules, err := ParseRuleExclusions([]string{"shared", "uir:specific", "eslint:other", "uir:shared"}, sources)
		Expect(err).NotTo(HaveOccurred())
		Expect(rules.NativeFor("uir", false)).To(Equal([]string{"specific", "shared"}))
		Expect(rules.NativeFor("uir", true)).To(Equal([]string{"shared", "specific"}))
	})
	It("removes finding failures without hiding execution errors or timeouts", func() {
		rules, err := ParseRuleExclusions([]string{"*"}, sources)
		Expect(err).NotTo(HaveOccurred())
		result := &LinterResult{Linter: "uir", Success: true, Violations: []models.Violation{{Rule: &models.Rule{Method: "sample"}, Severity: models.SeverityError}}}
		result.Violations = rules.Filter(result.Linter, result.Violations)
		Expect(HasFailures([]*LinterResult{result})).To(BeFalse())
		result.Success, result.Error = false, "tool failed"
		Expect(HasFailures([]*LinterResult{result})).To(BeTrue())
		result.Success, result.Error, result.TimedOut = true, "", true
		Expect(HasFailures([]*LinterResult{result})).To(BeTrue())
	})
	It("filters adapter output before recording the task verdict", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		rules, err := ParseRuleExclusions([]string{"excluded"}, sources)
		Expect(err).NotTo(HaveOccurred())
		opts := RunOptions{WorkDir: GinkgoT().TempDir(), RuleExclusions: rules}
		task := clicky.StartTask[*LinterResult]("excluded findings", func(ctx commonsContext.Context, task *clicky.Task) (*LinterResult, error) {
			return RunLinterWithTask(ctx, task, exclusionTestLinter{}, opts), nil
		})
		result, err := task.GetResult()
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Violations).To(BeEmpty())
		Expect(task.IsOk()).To(BeTrue())
		failed := RunLinterWithTask(context.Background(), nil, exclusionTestLinter{failure: errors.New("execution failed")}, opts)
		Expect(failed.Violations).To(BeEmpty())
		Expect(failed.Success).To(BeFalse())
		Expect(failed.Error).To(ContainSubstring("execution failed"))
	})
})

type exclusionTestLinter struct {
	dryRunLinter
	failure error
}

func (l exclusionTestLinter) Run(commonsContext.Context, *clicky.Task) ([]models.Violation, error) {
	return []models.Violation{{Rule: &models.Rule{Method: "excluded"}, Severity: models.SeverityError}}, l.failure
}
