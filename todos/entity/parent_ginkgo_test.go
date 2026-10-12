package entity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	clickyentity "github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// parentStore is one workspace holding one TODO. It records what create and
// edit hand the provider, which is all these specs assert on.
type parentStore struct {
	todos.Provider
	todo       *types.TODO
	created    []todos.CreateRequest
	parentRefs []string
}

func (s *parentStore) GetGlobal(context.Context, string) (*types.TODO, error) { return s.todo, nil }

func (s *parentStore) Get(context.Context, string) (*types.TODO, error) { return s.todo, nil }

func (s *parentStore) Create(_ context.Context, request todos.CreateRequest) (*types.TODO, error) {
	s.created = append(s.created, request)
	return s.todo, nil
}

func (s *parentStore) SetParent(_ context.Context, todo *types.TODO, parentRef string) error {
	s.parentRefs = append(s.parentRefs, parentRef)
	todo.ParentID = parentRef
	return nil
}

// refListProvider is a workspace that lists its TODOs and resolves a reference
// the way native storage does: by id, short id or alias. A title is not a ref.
type refListProvider struct {
	fixedListProvider
	aliases map[string]string
	// ambiguous is a prefix this workspace's ids share.
	ambiguous string
}

func (p refListProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	if p.ambiguous != "" && ref == p.ambiguous {
		return nil, fmt.Errorf("%w: issue reference %q", native.ErrAmbiguousReference, ref)
	}
	if id, ok := p.aliases[ref]; ok {
		ref = id
	}
	for _, item := range p.items {
		if item.ID == ref || item.ShortID == ref {
			copied := *item
			return &copied, nil
		}
	}
	if len(ref) < native.MinShortReferenceLength {
		return nil, fmt.Errorf("%w: short issue reference must contain at least %d characters",
			native.ErrInvalidInput, native.MinShortReferenceLength)
	}
	return nil, fmt.Errorf("%w: issue reference %q", native.ErrNotFound, ref)
}

var _ = Describe("TODO parents through the entity", func() {
	const (
		workDir   = "/work/gavel"
		parentID  = "0a1b2c3d-5f56-4c2c-833f-83e9b7774657"
		parentRef = "0a1b2c3d"
		runningID = "7e6d5c4b-5f56-4c2c-833f-83e9b7774657"
		sessionID = "session-2b9f"
	)
	cli := clickyentity.ContextWithOperationSurface(context.Background(), "cli")
	dashboard := clickyentity.ContextWithOperationSurface(context.Background(), "http")

	var (
		store *parentStore
		deps  Deps
	)

	BeforeEach(func() {
		existing := listedTodo("33e85357", "Tidy the projects bar", types.PriorityLow)
		existing.CWD = workDir
		store = &parentStore{todo: existing}
		deps = testDeps()
		deps.OpenProvider = func(context.Context, string) (todos.Provider, error) { return store, nil }
		deps.OpenGlobal = func(context.Context) (todos.GlobalReferenceProvider, error) { return store, nil }
		deps.DefaultDir = func(context.Context) (string, error) { return workDir, nil }
		GinkgoT().Setenv(commit.EnvIssueID, "")
		GinkgoT().Setenv(commit.EnvSessionID, "")
	})

	Describe("list", func() {
		const (
			clickyDir = "/work/clicky-ui"
			emptyDir  = "/work/empty"
			alias     = "jira-42"
		)

		BeforeEach(func() {
			child := listedTodo("9f8e7d6c", "Split out the alias import", types.PriorityMedium)
			child.ParentID = parentID
			routerChild := listedTodo("b7a6c5d4", "Document the adapter", types.PriorityMedium)
			routerChild.ParentID = "e28bd740-5f56-4c2c-833f-83e9b7774657"
			providers := map[string]todos.Provider{
				workDir: refListProvider{
					fixedListProvider: fixedListProvider{items: types.TODOS{
						listedTodo(parentRef, "Rework the importer", types.PriorityHigh), child,
						listedTodo("c1d2e3f4", "Shared title", types.PriorityLow),
					}},
					aliases: map[string]string{alias: parentID},
				},
				clickyDir: refListProvider{fixedListProvider: fixedListProvider{items: types.TODOS{
					listedTodo("e28bd740", "Document route.Router", types.PriorityMedium), routerChild,
					listedTodo("d4c3b2a1", "shared TITLE", types.PriorityLow),
				}}, ambiguous: "0a1b2c3d"},
				emptyDir: refListProvider{},
			}
			deps.OpenProvider = func(_ context.Context, dir string) (todos.Provider, error) {
				provider, ok := providers[dir]
				if !ok {
					return nil, errors.New("no workspace at " + dir)
				}
				return provider, nil
			}
			deps.Workspaces = func(context.Context) ([]query.Workspace, error) {
				return []query.Workspace{{Name: "Gavel", Dir: workDir}, {Name: "Clicky UI", Dir: clickyDir}}, nil
			}
		})

		DescribeTable("resolves the parent ref to one TODO before listing its children",
			func(ref string) {
				page, err := deps.list(context.Background(), query.ListOpts{Dir: workDir, Parent: ref, Limit: 50})
				Expect(err).NotTo(HaveOccurred())
				Expect(titles(page)).To(Equal([]string{"Split out the alias import"}))
			},
			Entry("a full id", parentID),
			Entry("a short id", parentRef),
			Entry("an alias", alias),
			Entry("a title", "rework the IMPORTER"),
		)

		It("finds the parent in whichever registered project owns it", func() {
			page, err := deps.list(context.Background(), query.ListOpts{Parent: "e28bd740", Limit: 50})

			Expect(err).NotTo(HaveOccurred())
			Expect(page.Data).To(ConsistOf(
				SatisfyAll(HaveField("Title", "Document the adapter"), HaveField("Project", "clicky-ui")),
			))
		})

		It("rejects a parent ref that names no TODO", func() {
			_, err := deps.list(context.Background(), query.ListOpts{Dir: workDir, Parent: "0a1b2c3e", Limit: 50})

			expectBadRequest(err, `parent "0a1b2c3e" does not name a TODO in /work/gavel`)
		})

		It("rejects a bad parent ref in an empty workspace rather than listing nothing", func() {
			_, err := deps.list(context.Background(), query.ListOpts{Dir: emptyDir, Parent: "0a1b", Limit: 50})

			expectBadRequest(err, `parent "0a1b" does not name a TODO in /work/empty`)
		})

		// The short id resolves in gavel, but clicky-ui has two ids under the same
		// prefix: picking gavel's would be guessing which TODO the caller meant.
		It("rejects a short id that another listed project finds ambiguous", func() {
			_, err := deps.list(context.Background(), query.ListOpts{Parent: parentRef, Limit: 50})

			expectBadRequest(err, `parent "0a1b2c3d" in clicky-ui: native todo reference is ambiguous`)
		})

		It("rejects a parent ref that names TODOs in two projects, naming both", func() {
			_, err := deps.list(context.Background(), query.ListOpts{Parent: "Shared Title", Limit: 50})

			expectBadRequest(err, `parent "Shared Title" names 2 TODOs`)
			expectBadRequest(err, `c1d2e3f4 "Shared title" in gavel`)
			expectBadRequest(err, `d4c3b2a1 "shared TITLE" in clicky-ui`)
		})

		DescribeTable("hides children unless the request asks for them",
			func(opts query.ListOpts, want ...string) {
				opts.Dir, opts.Limit = workDir, 50
				page, err := deps.list(context.Background(), opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(titles(page)).To(ConsistOf(want))
			},
			Entry("by default only top-level TODOs are listed", query.ListOpts{}, "Rework the importer", "Shared title"),
			Entry("children lists them beside their parents", query.ListOpts{Children: true},
				"Rework the importer", "Shared title", "Split out the alias import"),
			Entry("parent lists that TODO's children", query.ListOpts{Parent: parentRef}, "Split out the alias import"),
			Entry("a search finds a child", query.ListOpts{Search: "alias"}, "Split out the alias import"),
		)

		It("names a child's parent by short id on the row", func() {
			page, err := deps.list(context.Background(), query.ListOpts{Dir: workDir, Parent: parentRef, Limit: 50})
			Expect(err).NotTo(HaveOccurred())

			encoded, err := json.Marshal(page.Data)

			Expect(err).NotTo(HaveOccurred())
			Expect(encoded).To(MatchJSON(`[{
				"id": "9f8e7d6c",
				"title": "Split out the alias import",
				"status": "pending",
				"priority": "medium",
				"parent": "0a1b2c3d"
			}]`))
		})
	})

	Describe("create", func() {
		It("passes an explicit parent to the provider", func() {
			_, err := deps.create(cli, "", CreateFlags{Title: "Split out the alias import", Parent: parentRef})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.created).To(HaveLen(1))
			Expect(store.created[0].Parent).To(Equal(parentRef))
			Expect(store.created[0].NoParent).To(BeFalse())
			Expect(store.created[0].Origin).To(BeNil())
		})

		It("hands a terminal create the run it was started inside", func() {
			GinkgoT().Setenv(commit.EnvIssueID, runningID)
			GinkgoT().Setenv(commit.EnvSessionID, sessionID)

			_, err := deps.create(cli, "", CreateFlags{Title: "Found while running"})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.created[0].Origin).To(Equal(&todos.CreateOrigin{IssueID: runningID, SessionID: sessionID}))
		})

		It("keeps the origin and opts out of the parent with no-parent", func() {
			GinkgoT().Setenv(commit.EnvIssueID, runningID)

			_, err := deps.create(cli, "", CreateFlags{Title: "Unrelated follow-up", NoParent: true})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.created[0].NoParent).To(BeTrue())
			Expect(store.created[0].Origin).To(Equal(&todos.CreateOrigin{IssueID: runningID}))
		})

		// The server's environment describes whoever started the server, not the
		// request it is serving.
		It("does not read the server's environment for a dashboard create", func() {
			GinkgoT().Setenv(commit.EnvIssueID, runningID)

			_, err := deps.create(dashboard, "", CreateFlags{Title: "Created in the dashboard"})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.created[0].Origin).To(BeNil())
		})

		It("rejects parent together with no-parent before creating anything", func() {
			_, err := deps.create(cli, "", CreateFlags{Title: "Contradiction", Parent: parentRef, NoParent: true})

			expectBadRequest(err, "--no-parent cannot be combined with --parent")
			Expect(store.created).To(BeEmpty())
		})
	})

	Describe("edit", func() {
		It("attaches the TODO to the named parent", func() {
			edited, err := deps.edit(cli, "33e85357", EditFlags{Parent: parentRef})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.parentRefs).To(Equal([]string{parentRef}))
			Expect(edited.ParentID).To(Equal(parentRef))
		})

		It("detaches the TODO with no-parent", func() {
			store.todo.ParentID = parentID

			edited, err := deps.edit(cli, "33e85357", EditFlags{NoParent: true})

			Expect(err).NotTo(HaveOccurred())
			Expect(store.parentRefs).To(Equal([]string{""}))
			Expect(edited.ParentID).To(BeEmpty())
		})

		It("rejects parent together with no-parent before writing anything", func() {
			_, err := deps.edit(cli, "33e85357", EditFlags{Parent: parentRef, NoParent: true})

			expectBadRequest(err, "--no-parent cannot be combined with --parent")
			Expect(store.parentRefs).To(BeEmpty())
		})
	})
})
