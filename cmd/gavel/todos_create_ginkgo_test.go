package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestTodosCreateCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO create CLI")
}

var _ = Describe("TODO text inputs", func() {
	It("registers create plan and verification flags without body-file flags", func() {
		for _, name := range []string{"title", "body", "plan", "verification", "priority", "status"} {
			Expect(todosCreateCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing todos create --%s", name)
			Expect(todosEditCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing todos edit --%s", name)
		}
		Expect(todosCreateCmd.Flags().Lookup("body-file")).To(BeNil())
		Expect(todosEditCmd.Flags().Lookup("body-file")).To(BeNil())
		Expect(todosCommentCmd.Flags().Lookup("body-file")).To(BeNil())
		Expect(todosReopenCmd.Flags().Lookup("comment-file")).NotTo(BeNil())
	})

	DescribeTable("resolves literal and file-backed values",
		func(value string, prepare func(string) string, expected string) {
			workDir := GinkgoT().TempDir()
			if prepare != nil {
				value = prepare(workDir)
			}
			resolved, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: value})
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved).To(Equal(expected))
		},
		Entry("literal text", "inline markdown", nil, "inline markdown"),
		Entry("missing path-like text", "missing.md", nil, "missing.md"),
		Entry("escaped leading at sign", `\@reviewer`, nil, "@reviewer"),
		Entry("relative file", "", func(workDir string) string {
			path := filepath.Join(workDir, "body.md")
			Expect(os.WriteFile(path, []byte("from relative file\n"), 0o600)).To(Succeed())
			return "@body.md"
		}, "from relative file"),
		Entry("absolute file", "", func(workDir string) string {
			path := filepath.Join(workDir, "absolute.md")
			Expect(os.WriteFile(path, []byte("from absolute file\n"), 0o600)).To(Succeed())
			return "@" + path
		}, "from absolute file"),
		Entry("existing relative file without prefix", "", func(workDir string) string {
			path := filepath.Join(workDir, "body.md")
			Expect(os.WriteFile(path, []byte("from implicit relative file\n"), 0o600)).To(Succeed())
			return "body.md"
		}, "from implicit relative file"),
		Entry("existing absolute file without prefix", "", func(workDir string) string {
			path := filepath.Join(workDir, "implicit-absolute.md")
			Expect(os.WriteFile(path, []byte("from implicit absolute file\n"), 0o600)).To(Succeed())
			return path
		}, "from implicit absolute file"),
		Entry("existing directory remains literal", "", func(workDir string) string {
			Expect(os.Mkdir(filepath.Join(workDir, "docs"), 0o700)).To(Succeed())
			return "docs"
		}, "docs"),
	)

	It("reports the flag when a referenced file cannot be read", func() {
		_, err := resolveTodoText(todoTextOptions{
			WorkDir: GinkgoT().TempDir(),
			Flag:    "--verification",
			Value:   "@missing.md",
		})
		Expect(err).To(MatchError(ContainSubstring("resolve --verification")))
		Expect(err).To(MatchError(ContainSubstring("@missing.md")))
	})

	It("updates body, plan, and verification from existing paths", func() {
		workDir := GinkgoT().TempDir()
		provider := &testPlanReviewProvider{testTODOProvider: testProviderFor(workDir)}
		oldOpen := openRuntimeTodosProvider
		openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		DeferCleanup(func() { openRuntimeTodosProvider = oldOpen })

		created, err := provider.Create(context.Background(), todos.CreateRequest{
			Title: "Edit file-backed content", Body: "old body", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())

		bodyPath := filepath.Join(workDir, "body.md")
		planPath := filepath.Join(workDir, "plan.md")
		verificationPath := filepath.Join(workDir, "verification.md")
		Expect(os.WriteFile(bodyPath, []byte("new body\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(planPath, []byte("# Revised plan\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(verificationPath, []byte("verification fixture\n"), 0o600)).To(Succeed())

		oldWorkingDir := workingDir
		oldBody, oldPlan, oldVerification := todoEditBody, todoEditPlan, todoEditVerification
		workingDir = workDir
		todoEditBody, todoEditPlan, todoEditVerification = bodyPath, planPath, verificationPath
		DeferCleanup(func() {
			workingDir = oldWorkingDir
			todoEditBody, todoEditPlan, todoEditVerification = oldBody, oldPlan, oldVerification
		})

		cmd := &cobra.Command{}
		for _, name := range []string{"body", "plan", "verification"} {
			cmd.Flags().String(name, "", "")
			Expect(cmd.Flags().Set(name, "changed")).To(Succeed())
		}
		Expect(runTodosEdit(cmd, []string{created.ID})).To(Succeed())

		updated, err := provider.Get(context.Background(), created.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.MarkdownBody).To(Equal("new body"))
		Expect(updated.VerificationMarkdown).To(Equal("verification fixture"))
		Expect(provider.lastPlan).To(Equal("# Revised plan"))
		Expect(provider.planActor).To(BeEmpty())
	})
})

var _ = Describe("TODO create lifecycle", func() {
	DescribeTable("parses create statuses",
		func(value string, status types.Status, approved bool) {
			lifecycle, err := parseTodoCreateLifecycle(value)
			Expect(err).NotTo(HaveOccurred())
			Expect(lifecycle.Status).To(Equal(status))
			Expect(lifecycle.PlanApproved).To(Equal(approved))
		},
		Entry("ordinary pending todo", "pending", types.StatusPending, false),
		Entry("approved supplied plan", "approved", types.StatusPending, true),
	)

	It("rejects the approved status without plan content", func() {
		lifecycle, err := parseTodoCreateLifecycle("approved")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateTodoCreatePlan("", lifecycle)).To(MatchError("--status approved requires --plan"))
	})

	It("rejects empty plan and verification values when explicitly supplied", func() {
		_, err := resolveTodoCreateContent(GinkgoT().TempDir(), todoCreateContentOptions{PlanSet: true})
		Expect(err).To(MatchError("--plan cannot be empty"))

		_, err = resolveTodoCreateContent(GinkgoT().TempDir(), todoCreateContentOptions{VerificationSet: true})
		Expect(err).To(MatchError("--verification cannot be empty"))
	})

	DescribeTable("moves body verification after explicit verification",
		func(body func(string) string) {
			workDir := GinkgoT().TempDir()
			content, err := resolveTodoCreateContent(workDir, todoCreateContentOptions{
				BodySet:         true,
				Body:            body(workDir),
				VerificationSet: true,
				Verification:    "explicit fixture",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(content.Body).To(Equal("Parser failures lose context."))
			Expect(content.Verification).To(Equal("explicit fixture\n\nbody fixture"))
		},
		Entry("from inline body markdown", func(string) string {
			return `Parser failures lose context.

# Verification

body fixture`
		}),
		Entry("from file-backed body markdown", func(workDir string) string {
			path := filepath.Join(workDir, "todo.md")
			Expect(os.WriteFile(path, []byte("Parser failures lose context.\n\n## Verification\n\nbody fixture\n"), 0o600)).To(Succeed())
			return "@todo.md"
		}),
	)
})

var _ = Describe("TODO create help", func() {
	It("renders colored examples and verification fixture snippets", func() {
		help := todosCreateHelp(todosCreateCmd).ANSI()
		plain := stripANSI(help)

		Expect(help).To(ContainSubstring("\x1b["))
		for _, expected := range []string{
			"USAGE",
			"CONTENT INPUTS",
			"PLAN LIFECYCLE",
			"VERIFICATION FIXTURES",
			"gavel todos create",
			"--body @description.md",
			"--plan @plan.md",
			"--verification @verification.md",
			"--status approved",
			"# Verification",
			"extracted body fixtures are appended",
			"# verification.md",
			"cwd: .                  # Resolve paths from the repository root.",
			"ai: {}                  # Score the acceptance checklist against the diff.",
			"```yaml test",
			"```yaml lint",
			"## Acceptance Criteria",
			"- [ ] The supplied plan is persisted and selected on the TODO.",
			"gavel fixtures --help",
			"FLAGS",
			"GLOBAL FLAGS",
		} {
			Expect(plain).To(ContainSubstring(expected))
		}
		Expect(plain).NotTo(ContainSubstring("--body-file"))
	})
})
