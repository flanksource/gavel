package prwatch

import (
	"context"
	"strings"

	"github.com/flanksource/clicky/api"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	githublint "github.com/flanksource/gavel/linters/github"
	"github.com/flanksource/gavel/pr/model"
)

func fetchCodeQLFindings(opts github.Options, pr *model.PRInfo) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	for i := range pr.StatusCheckRollup {
		check := &pr.StatusCheckRollup[i]
		if check.AppSlug != "github-advanced-security" || !strings.EqualFold(check.Name, "CodeQL") {
			continue
		}
		check.LintError = ""
		target := *pr
		target.StatusCheckRollup = model.StatusChecks{*check}
		linter := githublint.New(opts.WorkDir)
		linter.AuditOptions = github.AuditOptions{Client: opts, PR: &target}
		result, err := linter.RunResult(commonsContext.NewContext(ctx), nil)
		check.Lint = result.Violations
		if err != nil {
			check.LintError = err.Error()
		}
	}
}

func prettyCheckLint(check model.StatusCheck) api.Text {
	if len(check.Lint) == 0 && check.LintError == "" {
		return api.Text{}
	}
	result := &linters.LinterResult{Linter: "github", Success: check.LintError == "", Violations: check.Lint, Error: check.LintError}
	return api.Text{}.NewLine().Add(treeText(result))
}
