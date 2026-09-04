package todos

import (
	"context"
	"errors"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO verification runtime spec", func() {
	BeforeEach(func() {
		// The definition of done loads the user's own ~/.gavel.yaml as a layer;
		// the specs assert against the built-in defaults, not this machine's.
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	})

	// C1: the grader used to execute as the implementer — same model, same
	// backend, same session — so a cmux run spawned a TUI to mark its own work.
	// The verify chain alone decides what grades, and the document carries the
	// grading model; the implementer's runtime has no way in.
	It("grades on the verify spec alone", func() {
		grader := api.Spec{Model: api.Model{Name: "claude-code-sonnet"}}

		dod, err := BuildDefinitionOfDone(DefinitionOfDoneOptions{
			WorkDir: GinkgoT().TempDir(),
			Todos: []*types.TODO{{
				AcceptanceCriteria: []types.AcceptanceCriterion{{Text: "The implementation passes."}},
			}},
			Grader: grader,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(dod.Declared()).To(BeTrue())
		front, _, err := fixtures.SplitFrontMatter(dod.Fixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(front).NotTo(BeNil())
		Expect(front.AI).NotTo(BeNil())
		Expect(front.AI.Model).To(Equal("claude-code-sonnet"))
	})

	verificationTodo := func(markdown string) []*types.TODO {
		return []*types.TODO{{VerificationMarkdown: markdown}}
	}
	const verificationBody = "\n## Verification\n\n### command: unit tests\n\n```bash\necho ok\n```\n\n- cel: exitCode == 0\n"

	// The todo's `## Verification` front matter becomes the generated document's:
	// every key the node runner honours travels, so the steps run the way the
	// todo said they would.
	It("carries the verification section's honoured front matter into the document", func() {
		dod, err := BuildDefinitionOfDone(DefinitionOfDoneOptions{
			WorkDir: GinkgoT().TempDir(),
			Todos: verificationTodo("---\ncodeBlocks: [bash, sh]\ncwd: sub\nexec: bash\nargs: [-e]\n" +
				"env: {CI: 'true'}\nterminal: pty\nos: '!plan9'\narch: arm64\nskip: 'false'\n---" + verificationBody),
		})

		Expect(err).NotTo(HaveOccurred())
		front, _, err := fixtures.SplitFrontMatter(dod.Fixture)
		Expect(err).NotTo(HaveOccurred())
		Expect(front.CodeBlocks).To(Equal([]string{"bash", "sh"}))
		Expect(front.CWD).To(Equal("sub"))
		Expect(front.Exec).To(Equal("bash"))
		Expect(front.Args).To(Equal([]string{"-e"}))
		Expect(front.Env).To(Equal(map[string]any{"CI": "true"}))
		Expect(front.Terminal).To(Equal("pty"))
		Expect(front.OS).To(Equal("!plan9"))
		Expect(front.Arch).To(Equal("arm64"))
		Expect(front.Skip).To(Equal("false"))
	})

	// A key the node runner would ignore is refused, not dropped: a todo whose
	// verification declared a `setup:` would otherwise believe its steps ran in
	// a worktree they never had.
	DescribeTable("refuses verification front matter the run loop cannot honour",
		func(frontMatter, key string) {
			_, err := BuildDefinitionOfDone(DefinitionOfDoneOptions{
				WorkDir: GinkgoT().TempDir(),
				Todos:   verificationTodo("---\n" + frontMatter + "\n---" + verificationBody),
			})

			Expect(err).To(MatchError(ContainSubstring("front matter sets " + key)))
			Expect(err).To(MatchError(ContainSubstring("cannot honour")))
		},
		Entry("setup", "setup:\n  checkout:\n    worktree:\n      mode: new", "setup"),
		Entry("record", "record:\n  http:\n    mode: mitm", "record"),
		Entry("build", "build: make build", "build"),
		Entry("daemon", "daemon: ./serve", "daemon"),
		Entry("files", "files: '**/*.fixture.md'", "files"),
		Entry("the grader, which is the generated document's own", "ai:\n  model: sonnet", "ai"),
		Entry("the retry predicate, which is the generated document's own", "verify:\n  retry: 'true'", "verify"),
		Entry("an unknown key", "codeBlocks: [bash]\nparallel: true", "parallel"),
	)

	// A check dispatches the lifecycle's verify step, so a check with nowhere to
	// dispatch it has judged nothing. Reporting that as a pass is the one
	// outcome that must never happen.
	It("refuses to check with no lifecycle runner rather than passing", func() {
		result := CheckTODO(context.Background(), &types.TODO{ID: "todo-1"}, CheckOptions{})

		Expect(result.AllPassed).To(BeFalse())
		Expect(result.Error).To(MatchError(ContainSubstring("no lifecycle runner")))
	})

	It("reports a verify step that could not run as a failed check", func() {
		result := CheckTODO(context.Background(), &types.TODO{ID: "todo-1"}, CheckOptions{
			Runner: verifyRunnerFunc(func(context.Context, *types.TODO, api.Spec) (*types.CheckResult, error) {
				return nil, errors.New("no verification fixture, acceptance criteria, or configured checks")
			}),
		})

		Expect(result.AllPassed).To(BeFalse())
		Expect(result.Error).To(MatchError(ContainSubstring("no verification fixture")))
	})

	It("refuses a verify step that returned neither a result nor an error", func() {
		result := CheckTODO(context.Background(), &types.TODO{ID: "todo-1"}, CheckOptions{
			Runner: verifyRunnerFunc(func(context.Context, *types.TODO, api.Spec) (*types.CheckResult, error) {
				return nil, nil
			}),
		})

		Expect(result.AllPassed).To(BeFalse())
		Expect(result.Error).To(MatchError(ContainSubstring("produced no result")))
	})

	It("passes the caller's request spec through to the verify step", func() {
		var seen api.Spec
		request := api.Spec{Budget: api.Budget{Timeout: "12m"}, Model: api.Model{Name: "claude-code-sonnet"}}

		result := CheckTODO(context.Background(), &types.TODO{ID: "todo-1"}, CheckOptions{
			Request: request,
			Runner: verifyRunnerFunc(func(_ context.Context, _ *types.TODO, spec api.Spec) (*types.CheckResult, error) {
				seen = spec
				return &types.CheckResult{AllPassed: true}, nil
			}),
		})

		Expect(result.Error).NotTo(HaveOccurred())
		Expect(result.AllPassed).To(BeTrue())
		Expect(seen).To(Equal(request))
	})
})

// verifyRunnerFunc adapts a function to VerifyRunner.
type verifyRunnerFunc func(ctx context.Context, todo *types.TODO, request api.Spec) (*types.CheckResult, error)

func (f verifyRunnerFunc) VerifyStep(ctx context.Context, todo *types.TODO, request api.Spec) (*types.CheckResult, error) {
	return f(ctx, todo, request)
}
