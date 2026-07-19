package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/spf13/cobra"
)

const todoVerificationExample = "  # verification.md\n" +
	"  ---\n" +
	"  cwd: .                  # Resolve paths from the repository root.\n" +
	"  timeout: 10m            # Limit the complete verification run.\n" +
	"  codeBlocks: [test, lint] # Execute only Gavel runner fences below.\n" +
	"  ai: {}                  # Score the acceptance checklist against the diff.\n" +
	"  verify:\n" +
	"    scope: diff            # Review only working-tree changes.\n" +
	"    threshold: 80          # Require an 80 percent checklist score.\n" +
	"  ---\n\n" +
	"  ## Focused tests\n" +
	"  <!-- Run discovered Go and Ginkgo tests through the Gavel test engine. -->\n" +
	"  ```yaml test\n" +
	"  paths: [./cmd/gavel, ./todos/...]\n" +
	"  framework: [go, ginkgo]\n" +
	"  test-timeout: 2m\n" +
	"  show-passed: true\n" +
	"  ```\n\n" +
	"  ## Changed-code lint\n" +
	"  <!-- Run detected linters on changed files without applying fixes. -->\n" +
	"  ```yaml lint\n" +
	"  changed: true\n" +
	"  fix: false\n" +
	"  timeout: 5m\n" +
	"  ```\n\n" +
	"  ## Acceptance Criteria\n" +
	"  <!-- Each unchecked task is scored against the implementation diff. -->\n" +
	"  - [ ] The supplied plan is persisted and selected on the TODO.\n" +
	"  - [ ] Focused tests and lint complete without failures.\n" +
	"  - [ ] The CLI help documents every supported content input."

func todosCreateHelp(cmd *cobra.Command) api.Text {
	const (
		heading = "font-bold text-purple-600"
		flag    = "text-cyan-600 font-bold"
		code    = "text-green-500"
		muted   = "text-muted"
	)

	t := clicky.Text("Create a PostgreSQL-backed TODO with an optional durable plan and definition-of-done fixture.", "font-bold").
		NewLine().NewLine().
		Append("USAGE", heading).NewLine().
		Append("  ").Append("gavel todos create [title...] [flags]", code).NewLine().NewLine().
		Append("CONTENT INPUTS", heading).NewLine().
		Append("  ").Append("--body", flag).Append(", ", muted).Append("--plan", flag).Append(", and ", muted).Append("--verification", flag).
		Append(" accept inline text or ", muted).Append("@path", code).Append(".", muted).NewLine().
		Append("  Relative paths resolve from ", muted).Append("--cwd", flag).Append("; use ", muted).Append(`\@text`, code).
		Append(" for a literal leading @.", muted).NewLine().
		Append("  An explicit ", muted).Append("--verification", flag).Append(" replaces any Verification section in the body.", muted).NewLine().NewLine().
		Append("PLAN LIFECYCLE", heading).NewLine().
		Append("  ").Append("--plan", flag).Append(" stores an immutable Captain plan revision selected on the TODO.", muted).NewLine().
		Append("  By default the plan awaits approval and the TODO appears in ", muted).Append("review", code).Append(".", muted).NewLine().
		Append("  ").Append("--status approved", flag).Append(" records the supplied plan as reviewed and approved,", muted).NewLine().
		Append("  leaving the TODO pending and ready for ", muted).Append("gavel todos run", code).Append(".", muted).NewLine().NewLine().
		Append("VERIFICATION FIXTURES", heading).NewLine().
		Append("  Supply fixture markdown without an outer ", muted).Append("## Verification", code).Append(" heading.", muted).NewLine().
		Append("  This annotated file enables checklist review and runs focused test and lint gates:", muted).NewLine().
		Append(todoVerificationExample, code).NewLine().NewLine().
		Append("  Use it with ", muted).Append(`--verification @verification.md`, code).Append("; it is also runnable on its own.", muted).NewLine().
		Append("  See ", muted).Append("gavel fixtures --help", code).Append(" for frontmatter, checklists, command blocks, tables, CEL assertions, and every runner key.", muted).NewLine().NewLine().
		Append("EXAMPLES", heading).NewLine().
		Append("  ").Append(`gavel todos create "Fix flaky parser"`, code).Append("  create a pending TODO", muted).NewLine().
		Append("  ").Append(`gavel todos create --title "Fix parser" --body @description.md`, code).Append("  read the body from --cwd", muted).NewLine().
		Append("  ").Append(`gavel todos create "Fix parser" --plan @plan.md`, code).Append("  create a plan awaiting review", muted).NewLine().
		Append("  ").Append(`gavel todos create "Fix parser" --plan @plan.md --status approved`, code).Append("  create a runnable approved plan", muted).NewLine().
		Append("  ").Append(`gavel todos create "Fix parser" --verification @verification.md`, code).Append("  attach a definition of done", muted).NewLine().NewLine().
		Add(renderHelpFlags("FLAGS", cmd.NonInheritedFlags())).
		Add(renderHelpFlags("GLOBAL FLAGS", cmd.InheritedFlags()))

	return t
}
