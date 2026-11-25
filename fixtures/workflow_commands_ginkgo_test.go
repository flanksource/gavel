package fixtures

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/models"
)

var _ = Describe("workflow commands", func() {
	DescribeTable("parsing a line",
		func(line string, ok bool, want workflowCommand) {
			cmd, parsed, err := parseWorkflowCommand(line)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(ok))
			if ok {
				Expect(cmd).To(Equal(want))
			}
		},
		Entry("a group", "::group::Build images", true, workflowCommand{Name: "group", Message: "Build images"}),
		Entry("an endgroup with a CRLF ending", "::endgroup::\r", true, workflowCommand{Name: "endgroup"}),
		Entry("an error with properties", "::error file=app.go,line=12,title=Lint::unused var", true,
			workflowCommand{Name: "error", Message: "unused var", Properties: map[string]string{"file": "app.go", "line": "12", "title": "Lint"}}),
		Entry("escaped message and properties", "::warning title=a%3Ab%2Cc::50%25 done%0Anext", true,
			workflowCommand{Name: "warning", Message: "50% done\nnext", Properties: map[string]string{"title": "a:b,c"}}),
		Entry("progress with properties", "::progress done=3,total=10::", true,
			workflowCommand{Name: "progress", Properties: map[string]string{"done": "3", "total": "10"}}),
		Entry("an unknown command stays output", "::set-env name=FOO::bar", false, workflowCommand{}),
		Entry("text that only starts with colons", "::not a command", false, workflowCommand{}),
		Entry("an indented command stays output", "  ::error::nope", false, workflowCommand{}),
	)

	It("fails a known command whose property is not key=value", func() {
		_, ok, err := parseWorkflowCommand("::error file::boom")
		Expect(ok).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring(`property "file" is not key=value`)))
	})

	DescribeTable("reading progress",
		func(line string, done, total int, failure string) {
			cmd, _, err := parseWorkflowCommand(line)
			Expect(err).NotTo(HaveOccurred())
			d, t, err := workflowProgress(cmd)
			if failure != "" {
				Expect(err).To(MatchError(ContainSubstring(failure)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect([]int{d, t}).To(Equal([]int{done, total}))
		},
		Entry("done/total message", "::progress::3/10", 3, 10, ""),
		Entry("done and total properties", "::progress done=7,total=9::", 7, 9, ""),
		Entry("a message with no slash", "::progress::half", 0, 0, "want a <done>/<total> message"),
		Entry("a non-numeric done", "::progress::x/10", 0, 0, `done "x" is not a number`),
		Entry("a negative done", "::progress::-1/10", 0, 0, "is negative"),
	)

	It("maps an annotation's properties onto the finding, in its group", func() {
		cmd, _, err := parseWorkflowCommand("::warning file=db/q.sql,line=4,col=2,endLine=6,title=Slow query::full scan")
		Expect(err).NotTo(HaveOccurred())

		v, err := workflowFinding(cmd, "migrations")

		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(models.Violation{Severity: models.SeverityWarning, Message: models.StringPtr("full scan"),
			Source: WorkflowCommandSource, Rule: &models.Rule{Method: "Slow query"}, Code: models.StringPtr("migrations"),
			File: "db/q.sql", Line: 4, Column: 2, EndLine: 6}))
	})

	It("rejects an annotation whose line is not a number", func() {
		cmd, _, _ := parseWorkflowCommand("::error file=a.go,line=twelve::boom")
		_, err := workflowFinding(cmd, "")
		Expect(err).To(MatchError(`::error line="twelve" is not a line or column number`))
	})

	DescribeTable("stripping output",
		func(raw, want string) { Expect(StripWorkflowCommands(raw)).To(Equal(want)) },
		Entry("plain output is unchanged", "one\ntwo", "one\ntwo"),
		Entry("annotations, progress and status are removed",
			"start\n::notice::hi\n::progress::1/2\n::status::halfway\n::description::copying\n::debug::x\nend\n", "start\nend\n"),
		Entry("group markers are kept in canonical form",
			"::group::Build\nstep\n::endgroup::\r\n", "::group::Build\nstep\n::endgroup::\n"),
		Entry("stop-commands passes commands through until its token",
			"::stop-commands::tok\n::error::shown\n::tok::\n::error::hidden\n", "::error::shown\n"),
		Entry("unknown commands stay", "::set-output name=x::1\n", "::set-output name=x::1\n"),
		Entry("a final line without a newline", "a\n::warning::w", "a\n"),
	)
})
