package main

import (
	"context"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// parentTestProvider is the in-memory test provider with the parent link the
// native runtime stores: Create and SetParent resolve the ref and record what
// they were asked.
type parentTestProvider struct {
	*testTODOProvider
	created []todos.CreateRequest
}

func (p *parentTestProvider) Create(ctx context.Context, request todos.CreateRequest) (*types.TODO, error) {
	p.created = append(p.created, request)
	todo, err := p.testTODOProvider.Create(ctx, request)
	if err != nil || request.Parent == "" {
		return todo, err
	}
	return todo, p.SetParent(ctx, todo, request.Parent)
}

func (p *parentTestProvider) SetParent(ctx context.Context, todo *types.TODO, parentRef string) error {
	if parentRef == "" {
		todo.ParentID = ""
		return nil
	}
	parent, err := p.Get(ctx, parentRef)
	if err != nil {
		return err
	}
	todo.ParentID = parent.ID
	return nil
}

var _ = Describe("TODO parents on the command line", func() {
	const (
		runningID = "7e6d5c4b-5f56-4c2c-833f-83e9b7774657"
		sessionID = "session-5d0a"
	)
	var (
		provider *parentTestProvider
		parent   *types.TODO
	)

	titles := func(list types.TODOS) []string {
		names := make([]string, len(list))
		for i, todo := range list {
			names[i] = todo.Title
		}
		return names
	}
	list := func(opts TodosListOptions) []string {
		GinkgoHelper()
		out, err := runTodosList(opts)
		Expect(err).NotTo(HaveOccurred())
		return titles(out.(types.TODOS))
	}

	BeforeEach(func() {
		workDir := GinkgoT().TempDir()
		provider = &parentTestProvider{testTODOProvider: testProviderFor(workDir)}
		oldOpen, oldWorkingDir := openRuntimeTodosProvider, workingDir
		openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		workingDir = workDir
		DeferCleanup(func() { openRuntimeTodosProvider, workingDir = oldOpen, oldWorkingDir })
		GinkgoT().Setenv(commitpkg.EnvIssueID, "")
		GinkgoT().Setenv(commitpkg.EnvSessionID, "")

		var err error
		parent, err = provider.Create(context.Background(), todos.CreateRequest{Title: "Rework the importer"})
		Expect(err).NotTo(HaveOccurred())
		provider.created = nil
	})

	It("registers the parent flags on create, edit and list", func() {
		for _, name := range []string{"parent", "no-parent"} {
			Expect(todosCreateCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing todos create --%s", name)
			Expect(todosEditCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing todos edit --%s", name)
		}
		listCmd, _, err := todosCmd.Find([]string{"list"})
		Expect(err).NotTo(HaveOccurred())
		Expect(listCmd.Name()).To(Equal("list"))
		for _, name := range []string{"parent", "children"} {
			Expect(listCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing todos list --%s", name)
		}
	})

	Describe("create", func() {
		create := func(opts TodosCreateOptions) error {
			var err error
			captureStdout(GinkgoTB(), func() { err = runTodosCreate(opts) })
			return err
		}

		It("passes --parent to the provider", func() {
			Expect(create(TodosCreateOptions{Title: "Import the aliases", Parent: parent.ID})).To(Succeed())

			Expect(provider.created).To(HaveLen(1))
			Expect(provider.created[0].Parent).To(Equal(parent.ID))
			Expect(provider.created[0].Origin).To(BeNil())
		})

		It("hands the provider the run it was started inside", func() {
			GinkgoT().Setenv(commitpkg.EnvIssueID, runningID)
			GinkgoT().Setenv(commitpkg.EnvSessionID, sessionID)

			Expect(create(TodosCreateOptions{Title: "Found while running"})).To(Succeed())

			Expect(provider.created[0].Origin).To(Equal(&todos.CreateOrigin{IssueID: runningID, SessionID: sessionID}))
			Expect(provider.created[0].NoParent).To(BeFalse())
		})

		It("opts out of the parent with --no-parent while keeping the origin", func() {
			GinkgoT().Setenv(commitpkg.EnvIssueID, runningID)

			Expect(create(TodosCreateOptions{Title: "Unrelated follow-up", NoParent: true})).To(Succeed())

			Expect(provider.created[0].NoParent).To(BeTrue())
			Expect(provider.created[0].Origin).To(Equal(&todos.CreateOrigin{IssueID: runningID}))
		})

		It("refuses --parent together with --no-parent before creating anything", func() {
			err := create(TodosCreateOptions{Title: "Contradiction", Parent: parent.ID, NoParent: true})

			Expect(err).To(MatchError("--no-parent cannot be combined with --parent"))
			Expect(provider.created).To(BeEmpty())
		})

		It("documents the parent flags in its help", func() {
			plain := stripANSI(todosCreateHelp(todosCreateCmd).ANSI())

			for _, expected := range []string{"PARENT", "--parent", "--no-parent", "GAVEL_ISSUE_ID"} {
				Expect(plain).To(ContainSubstring(expected))
			}
		})
	})

	Context("with a child", func() {
		var child *types.TODO

		BeforeEach(func() {
			var err error
			child, err = provider.Create(context.Background(), todos.CreateRequest{Title: "Import the aliases", Parent: parent.ID})
			Expect(err).NotTo(HaveOccurred())
			_, err = provider.Create(context.Background(), todos.CreateRequest{Title: "Unrelated work"})
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("list",
			func(opts func() TodosListOptions, want ...string) {
				Expect(list(opts())).To(ConsistOf(want))
			},
			Entry("hides children by default", func() TodosListOptions { return TodosListOptions{} },
				"Rework the importer", "Unrelated work"),
			Entry("--children lists them beside their parents", func() TodosListOptions { return TodosListOptions{Children: true} },
				"Rework the importer", "Import the aliases", "Unrelated work"),
			Entry("--parent lists only that TODO's children", func() TodosListOptions { return TodosListOptions{Parent: parent.ID} },
				"Import the aliases"),
			Entry("--parent accepts a title", func() TodosListOptions { return TodosListOptions{Parent: "rework the importer"} },
				"Import the aliases"),
		)

		It("list reports a --parent that names nothing", func() {
			_, err := runTodosList(TodosListOptions{Parent: "no-such-todo"})

			Expect(err).To(MatchError(ContainSubstring(`"no-such-todo"`)))
		})

		It("edit --no-parent detaches and --parent attaches", func() {
			edit := func(opts TodosEditOptions) error {
				opts.IDs = []string{child.ID}
				var err error
				captureStdout(GinkgoTB(), func() { err = runTodosEdit(opts, nil) })
				return err
			}

			Expect(edit(TodosEditOptions{NoParent: true})).To(Succeed())
			Expect(child.ParentID).To(BeEmpty())
			Expect(list(TodosListOptions{})).To(ContainElement("Import the aliases"))

			Expect(edit(TodosEditOptions{Parent: parent.ID})).To(Succeed())
			Expect(child.ParentID).To(Equal(parent.ID))

			Expect(edit(TodosEditOptions{Parent: parent.ID, NoParent: true})).
				To(MatchError("--no-parent cannot be combined with --parent"))
		})

		It("get prints a child's parent", func() {
			out := captureStdout(GinkgoTB(), func() {
				Expect(runTodosGet(TodosGetOptions{TodoTargetOptions{IDs: []string{child.ID}}})).To(Succeed())
			})

			Expect(stripANSI(out)).To(ContainSubstring("Parent: " + parent.ID))
			Expect(stripANSI(out)).NotTo(ContainSubstring("Children"))
		})

		It("get lists a parent's children as rows with their phase columns", func() {
			child.Status = types.StatusVerified
			child.PhaseRuns = types.PhaseRuns{types.RunPhase: {State: "succeeded", DurationMS: 134_000}}

			out := stripANSI(captureStdout(GinkgoTB(), func() {
				Expect(runTodosGet(TodosGetOptions{TodoTargetOptions{IDs: []string{parent.ID}}})).To(Succeed())
			}))

			Expect(out).To(ContainSubstring("Children (1/1 done)"))
			Expect(out).To(ContainSubstring("Import the aliases"))
			Expect(out).To(ContainSubstring("2m"), "a child's run duration reaches the parent's view")
			Expect(out).NotTo(ContainSubstring("Unrelated work"))
		})

		It("get prints no children section for a TODO without any", func() {
			out := captureStdout(GinkgoTB(), func() {
				Expect(runTodosGet(TodosGetOptions{TodoTargetOptions{IDs: []string{"Unrelated work"}}})).To(Succeed())
			})

			Expect(stripANSI(out)).NotTo(ContainSubstring("Children"))
		})
	})
})
