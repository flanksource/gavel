package githubpush

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGitHubPush(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO GitHub push")
}

// fakeProvider implements the slice of todos.Provider that push exercises plus
// todos.AliasProvider. Unused methods fail loudly rather than returning zero
// values that would hide a wrong call.
type fakeProvider struct {
	todo       *types.TODO
	aliases    []todos.TodoAlias
	comments   []string
	plan       string
	planModes  []types.RunMode
	addErr     error
	aliasesErr error
	commentErr error
}

func (f *fakeProvider) Get(context.Context, string) (*types.TODO, error) { return f.todo, nil }

func (f *fakeProvider) PlanMarkdown(_ context.Context, _ *types.TODO, mode types.RunMode) (string, error) {
	f.planModes = append(f.planModes, mode)
	return f.plan, nil
}

func (f *fakeProvider) Aliases(context.Context, *types.TODO) ([]todos.TodoAlias, error) {
	return f.aliases, f.aliasesErr
}

func (f *fakeProvider) AddAlias(_ context.Context, _ *types.TODO, alias todos.TodoAlias) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.aliases = append(f.aliases, alias)
	return nil
}

func (f *fakeProvider) Comment(_ context.Context, _ *types.TODO, body string) error {
	if f.commentErr != nil {
		return f.commentErr
	}
	f.comments = append(f.comments, body)
	return nil
}

func (f *fakeProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	panic("List not expected")
}
func (f *fakeProvider) CountByStatus(context.Context) (map[types.Status]int, error) {
	panic("CountByStatus not expected")
}
func (f *fakeProvider) Create(context.Context, todos.CreateRequest) (*types.TODO, error) {
	panic("Create not expected")
}
func (f *fakeProvider) Delete(context.Context, *types.TODO) error { panic("Delete not expected") }
func (f *fakeProvider) Edit(context.Context, *types.TODO, todos.EditRequest) error {
	panic("Edit not expected")
}
func (f *fakeProvider) UpdateState(context.Context, *types.TODO, todos.StateUpdate) error {
	panic("UpdateState not expected")
}
func (f *fakeProvider) UpdateLatestFailure(context.Context, *types.TODO, *types.TestResultInfo) error {
	panic("UpdateLatestFailure not expected")
}
func (f *fakeProvider) SaveAttempt(context.Context, *types.TODO, *todos.ExecutionResult) error {
	panic("SaveAttempt not expected")
}

// bareProvider satisfies todos.Provider without todos.AliasProvider. The
// embedded nil interface supplies the methods push never reaches on it.
type bareProvider struct{ todos.Provider }

func (b *bareProvider) Get(context.Context, string) (*types.TODO, error) { return newTODO(), nil }

// planlessProvider links issues but cannot resolve durable plan content — the
// shape a store without Captain plans would have. It delegates rather than
// embeds so fakeProvider.PlanMarkdown is not promoted onto it.
type planlessProvider struct {
	todos.Provider
	inner *fakeProvider
}

func (p *planlessProvider) Get(ctx context.Context, ref string) (*types.TODO, error) {
	return p.inner.Get(ctx, ref)
}

func (p *planlessProvider) Aliases(ctx context.Context, todo *types.TODO) ([]todos.TodoAlias, error) {
	return p.inner.Aliases(ctx, todo)
}

func (p *planlessProvider) AddAlias(ctx context.Context, todo *types.TODO, alias todos.TodoAlias) error {
	return p.inner.AddAlias(ctx, todo, alias)
}

func (p *planlessProvider) Comment(ctx context.Context, todo *types.TODO, body string) error {
	return p.inner.Comment(ctx, todo, body)
}

type recordedIssue struct {
	opts  github.Options
	input github.IssueInput
}

// fakeGitHub records the issue writes and answers with a canned issue that
// reports the same create-vs-update outcome the real API would.
type fakeGitHub struct {
	calls  []recordedIssue
	result *github.IssueResult
	err    error
}

func (f *fakeGitHub) saveIssue(opts github.Options, in github.IssueInput) (*github.IssueResult, error) {
	f.calls = append(f.calls, recordedIssue{opts: opts, input: in})
	if f.err != nil {
		return nil, f.err
	}
	issue := *f.result
	if in.Number > 0 {
		issue.Number, issue.Updated = in.Number, true
		issue.URL = fmt.Sprintf("https://github.com/%s/issues/%d", issue.Repo, in.Number)
	}
	return &issue, nil
}

func newTODO() *types.TODO {
	return &types.TODO{
		ID:              "5f0b1b4e-0a5c-4c1f-9f9e-2f2a1b3c4d5e",
		ShortID:         "5f0b1b",
		MarkdownBody:    "The parser drops trailing commas.",
		Labels:          []string{"bug", "parser"},
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"},
	}
}

func pushTODO(provider todos.Provider, gh *fakeGitHub, opts Options) (*Result, error) {
	return pushWithDeps(context.Background(), provider, "5f0b1b", opts, deps{saveIssue: gh.saveIssue})
}

func okIssue() *github.IssueResult {
	return &github.IssueResult{
		Number: 7, URL: "https://github.com/acme/api/issues/7", Repo: "acme/api",
	}
}

var _ = Describe("pushing a TODO to GitHub", func() {
	var (
		provider *fakeProvider
		gh       *fakeGitHub
	)

	BeforeEach(func() {
		provider = &fakeProvider{todo: newTODO()}
		gh = &fakeGitHub{result: okIssue()}
	})

	It("sends the portable markdown body with no YAML frontmatter", func() {
		provider.todo.VerificationMarkdown = "```exec\ngo test ./parser\n```"

		result, err := pushTODO(provider, gh, Options{})

		Expect(err).ToNot(HaveOccurred())
		Expect(gh.calls).To(HaveLen(1))
		body := gh.calls[0].input.Body
		Expect(body).To(ContainSubstring("The parser drops trailing commas."))
		Expect(body).To(ContainSubstring("# Verification"))
		Expect(body).To(ContainSubstring("go test ./parser"))
		Expect(body).ToNot(HavePrefix("---"))
		Expect(gh.calls[0].input.Title).To(Equal("Fix the parser"))
		Expect(result.URL).To(Equal("https://github.com/acme/api/issues/7"))
	})

	It("records the issue as a github alias and a history comment", func() {
		result, err := pushTODO(provider, gh, Options{})

		Expect(err).ToNot(HaveOccurred())
		Expect(result.Alias).To(Equal("acme/api#7"))
		Expect(result.Updated).To(BeFalse())
		Expect(provider.aliases).To(ContainElement(todos.TodoAlias{Alias: "acme/api#7", Kind: AliasKind}))
		Expect(provider.comments).To(ConsistOf("Pushed to GitHub issue https://github.com/acme/api/issues/7"))
	})

	Context("the plan", func() {
		BeforeEach(func() {
			provider.plan = "1. Rewrite the tokenizer.\n2. Add a regression test."
			provider.todo.VerificationMarkdown = "```exec\ngo test ./parser\n```"
		})

		It("goes between the body and the verification fixture", func() {
			_, err := pushTODO(provider, gh, Options{Plan: true})

			Expect(err).ToNot(HaveOccurred())
			body := gh.calls[0].input.Body
			Expect(body).To(ContainSubstring("# Plan\n\n1. Rewrite the tokenizer."))
			Expect(strings.Index(body, "The parser drops trailing commas.")).
				To(BeNumerically("<", strings.Index(body, "# Plan")))
			Expect(strings.Index(body, "# Plan")).To(BeNumerically("<", strings.Index(body, "# Verification")))
		})

		It("reads the latest revision rather than only an approved one", func() {
			_, err := pushTODO(provider, gh, Options{Plan: true})

			Expect(err).ToNot(HaveOccurred())
			Expect(provider.planModes).To(ConsistOf(types.ModePlan))
		})

		It("is left out when the todo has none", func() {
			provider.plan = ""

			_, err := pushTODO(provider, gh, Options{Plan: true})

			Expect(err).ToNot(HaveOccurred())
			Expect(gh.calls[0].input.Body).ToNot(ContainSubstring("# Plan"))
		})

		It("is left out when the caller opts out", func() {
			_, err := pushTODO(provider, gh, Options{Plan: false})

			Expect(err).ToNot(HaveOccurred())
			Expect(gh.calls[0].input.Body).ToNot(ContainSubstring("# Plan"))
			Expect(provider.planModes).To(BeEmpty())
		})

		It("rejects a provider that cannot resolve durable plan content", func() {
			_, err := pushWithDeps(context.Background(), &planlessProvider{inner: provider},
				"5f0b1b", Options{Plan: true}, deps{saveIssue: gh.saveIssue})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not expose durable plan content"))
			Expect(err.Error()).To(ContainSubstring("--plan=false"))
			Expect(gh.calls).To(BeEmpty())
		})
	})

	Context("re-pushing onto an existing issue", func() {
		BeforeEach(func() {
			provider.aliases = []todos.TodoAlias{{Alias: "acme/api#3", Kind: AliasKind}}
		})

		It("edits the linked issue instead of opening a new one", func() {
			provider.todo.MarkdownBody = "Now with more detail."

			result, err := pushTODO(provider, gh, Options{Update: true})

			Expect(err).ToNot(HaveOccurred())
			Expect(gh.calls).To(HaveLen(1))
			Expect(gh.calls[0].input.Number).To(Equal(3))
			Expect(gh.calls[0].opts.Repo).To(Equal("acme/api"))
			Expect(gh.calls[0].input.Body).To(ContainSubstring("Now with more detail."))
			Expect(result.Updated).To(BeTrue())
			Expect(result.URL).To(Equal("https://github.com/acme/api/issues/3"))
		})

		It("keeps the single existing alias and records an update comment", func() {
			_, err := pushTODO(provider, gh, Options{Update: true})

			Expect(err).ToNot(HaveOccurred())
			Expect(provider.aliases).To(ConsistOf(todos.TodoAlias{Alias: "acme/api#3", Kind: AliasKind}))
			Expect(provider.comments).To(ConsistOf("Updated GitHub issue https://github.com/acme/api/issues/3"))
		})

		DescribeTable("targeting one issue explicitly",
			func(ref string, wantRepo string, wantNumber int) {
				provider.aliases = nil

				result, err := pushTODO(provider, gh, Options{Issue: ref})

				Expect(err).ToNot(HaveOccurred())
				Expect(gh.calls[0].input.Number).To(Equal(wantNumber))
				Expect(gh.calls[0].opts.Repo).To(Equal(wantRepo))
				Expect(result.Updated).To(BeTrue())
			},
			Entry("bare number falls back to the workspace remote", "412", "", 412),
			Entry("alias form", "acme/api#412", "acme/api", 412),
			Entry("issue URL", "https://github.com/acme/api/issues/412", "acme/api", 412),
		)

		It("links an issue gavel did not open", func() {
			provider.aliases = nil

			_, err := pushTODO(provider, gh, Options{Issue: "acme/api#412"})

			Expect(err).ToNot(HaveOccurred())
			Expect(provider.aliases).To(ConsistOf(todos.TodoAlias{Alias: "acme/api#412", Kind: AliasKind}))
		})

		It("rejects an unparseable issue reference before calling GitHub", func() {
			_, err := pushTODO(provider, gh, Options{Issue: "the-parser-one"})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("is not a GitHub issue reference"))
			Expect(gh.calls).To(BeEmpty())
		})

		It("asks which issue to rewrite when the todo is linked to several", func() {
			provider.aliases = append(provider.aliases, todos.TodoAlias{Alias: "acme/api#9", Kind: AliasKind})

			_, err := pushTODO(provider, gh, Options{Update: true})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("acme/api#3, acme/api#9"))
			Expect(err.Error()).To(ContainSubstring("--issue"))
			Expect(gh.calls).To(BeEmpty())
		})

		It("refuses to update a todo that has never been pushed", func() {
			provider.aliases = nil

			_, err := pushTODO(provider, gh, Options{Update: true})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not linked to a GitHub issue yet"))
			Expect(gh.calls).To(BeEmpty())
		})

		It("refuses to both open and rewrite in one push", func() {
			_, err := pushTODO(provider, gh, Options{Force: true, Update: true})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pick one"))
			Expect(gh.calls).To(BeEmpty())
		})
	})

	It("copies labels only when asked", func() {
		_, err := pushTODO(provider, gh, Options{Labels: true})
		Expect(err).ToNot(HaveOccurred())
		Expect(gh.calls[0].input.Labels).To(Equal([]string{"bug", "parser"}))

		gh.calls = nil
		provider.aliases = nil
		_, err = pushTODO(provider, gh, Options{Labels: false})
		Expect(err).ToNot(HaveOccurred())
		Expect(gh.calls[0].input.Labels).To(BeEmpty())
	})

	It("refuses a TODO that already carries a github alias", func() {
		provider.aliases = []todos.TodoAlias{{Alias: "acme/api#3", Kind: AliasKind}}

		_, err := pushTODO(provider, gh, Options{})

		Expect(err).To(MatchError(ErrAlreadyLinked))
		Expect(err.Error()).To(ContainSubstring("acme/api#3"))
		Expect(err.Error()).To(ContainSubstring("--force"))
		Expect(gh.calls).To(BeEmpty())
	})

	It("ignores non-github aliases when deciding whether the TODO is linked", func() {
		provider.aliases = []todos.TodoAlias{{Alias: "legacy-42", Kind: "import"}}

		_, err := pushTODO(provider, gh, Options{})

		Expect(err).ToNot(HaveOccurred())
		Expect(gh.calls).To(HaveLen(1))
	})

	It("appends a second alias under --force without dropping the first", func() {
		provider.aliases = []todos.TodoAlias{{Alias: "acme/api#3", Kind: AliasKind}}

		_, err := pushTODO(provider, gh, Options{Force: true})

		Expect(err).ToNot(HaveOccurred())
		Expect(provider.aliases).To(ConsistOf(
			todos.TodoAlias{Alias: "acme/api#3", Kind: AliasKind},
			todos.TodoAlias{Alias: "acme/api#7", Kind: AliasKind},
		))
	})

	It("rejects a provider that cannot record external references", func() {
		_, err := pushWithDeps(context.Background(), &bareProvider{}, "ref", Options{}, deps{saveIssue: gh.saveIssue})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("does not support external issue links"))
	})

	Context("attachments", func() {
		BeforeEach(func() {
			provider.todo.MarkdownBody = "See:\n\n![shot.png](" + todos.AttachmentURLPrefix + "abc.png)"
		})

		It("fails loudly when no base URL is configured", func() {
			_, err := pushTODO(provider, gh, Options{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no base URL is configured"))
			Expect(err.Error()).To(ContainSubstring("--base-url"))
			Expect(gh.calls).To(BeEmpty())
			Expect(provider.aliases).To(BeEmpty())
		})

		It("rewrites the links against the base URL", func() {
			_, err := pushTODO(provider, gh, Options{BaseURL: "https://gavel.example.com"})

			Expect(err).ToNot(HaveOccurred())
			Expect(gh.calls[0].input.Body).To(ContainSubstring(
				"![shot.png](https://gavel.example.com" + todos.AttachmentURLPrefix + "abc.png)"))
		})
	})

	Context("failure handling", func() {
		It("writes no alias when GitHub rejects the issue", func() {
			gh.err = errors.New("HTTP 422: Validation Failed")

			_, err := pushTODO(provider, gh, Options{})

			Expect(err).To(HaveOccurred())
			Expect(provider.aliases).To(BeEmpty())
			Expect(provider.comments).To(BeEmpty())
		})

		It("names the created issue when the local link cannot be recorded", func() {
			provider.addErr = fmt.Errorf("version conflict")

			_, err := pushTODO(provider, gh, Options{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("https://github.com/acme/api/issues/7"))
			Expect(err.Error()).To(ContainSubstring("DUPLICATE"))
		})
	})
})

var _ = Describe("ResolveBaseURL", func() {
	It("picks the first non-empty candidate", func() {
		got, err := ResolveBaseURL("", "  ", "https://gavel.example.com", "http://localhost:9092")
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal("https://gavel.example.com"))
	})

	It("trims a trailing slash", func() {
		got, err := ResolveBaseURL("https://gavel.example.com/")
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(Equal("https://gavel.example.com"))
	})

	It("returns empty when nothing is configured", func() {
		got, err := ResolveBaseURL("", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	DescribeTable("rejecting non-origins",
		func(candidate string) {
			_, err := ResolveBaseURL(candidate)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("absolute http(s) origin"))
		},
		Entry("no scheme", "gavel.example.com"),
		Entry("wrong scheme", "ftp://gavel.example.com"),
		Entry("scheme without host", "https://"),
	)

	DescribeTable("loopback detection",
		func(baseURL string, want bool) {
			Expect(IsLoopback(baseURL)).To(Equal(want))
		},
		Entry("localhost", "http://localhost:9092", true),
		Entry("127.0.0.1", "http://127.0.0.1:9092", true),
		Entry("public host", "https://gavel.example.com", false),
	)
})
