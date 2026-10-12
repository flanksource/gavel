package query

import (
	"context"
	"errors"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQuery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO Query Suite")
}

// listProvider answers List with a fixed page or failure; nothing else on the
// Provider is reachable from a list.
type listProvider struct {
	todos.Provider
	items types.TODOS
	err   error
}

func (p *listProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	return p.items, p.err
}

// opener serves one provider per directory and counts how often each was opened.
type opener struct {
	providers map[string]todos.Provider
	failures  map[string]error
	opened    map[string]int
}

func (o *opener) open(_ context.Context, dir string) (todos.Provider, error) {
	if o.opened == nil {
		o.opened = map[string]int{}
	}
	o.opened[dir]++
	if err := o.failures[dir]; err != nil {
		return nil, err
	}
	provider, ok := o.providers[dir]
	if !ok {
		return nil, errors.New("no provider for " + dir)
	}
	return provider, nil
}

var _ = DescribeTable("ShortProjectName",
	func(name, short string) {
		Expect(ShortProjectName(name)).To(Equal(short))
	},
	Entry("lowercases and dashes spaces", "OM Digital Frontend", "om-digital-frontend"),
	Entry("keeps an existing slug", "mission-control-oipa", "mission-control-oipa"),
	Entry("collapses runs of punctuation", "Clicky  UI (v2)", "clicky-ui-v2"),
	Entry("trims leading and trailing separators", "  _Docs_ ", "docs"),
	Entry("keeps non-ASCII letters", "Café Ops", "café-ops"),
	Entry("has nothing to slug", "--", ""),
)

var _ = Describe("IndexByShortName", func() {
	It("keys each project by its short name", func() {
		index, err := IndexByShortName([]Workspace{{Name: "Clicky UI", Dir: "/a"}, {Name: "gavel", Dir: "/b"}})

		Expect(err).NotTo(HaveOccurred())
		Expect(index).To(Equal(map[string]Workspace{
			"clicky-ui": {Name: "Clicky UI", Dir: "/a"},
			"gavel":     {Name: "gavel", Dir: "/b"},
		}))
	})

	It("rejects two projects that share a short name, naming both", func() {
		_, err := IndexByShortName([]Workspace{{Name: "Clicky UI", Dir: "/a"}, {Name: "clicky-ui", Dir: "/b"}})

		Expect(err).To(MatchError(`projects "Clicky UI" and "clicky-ui" share the short name "clicky-ui"; rename one of them`))
	})

	It("rejects a project whose name has nothing to slug", func() {
		_, err := IndexByShortName([]Workspace{{Name: "--", Dir: "/a"}})

		Expect(err).To(MatchError(ContainSubstring(`project "--" has no letters or digits`)))
	})
})

var _ = Describe("ListWorkspaces", func() {
	const firstDir, secondDir = "/work/first", "/work/second"

	It("reads each distinct directory once and tags rows with their project", func() {
		o := &opener{providers: map[string]todos.Provider{
			firstDir:  &listProvider{items: types.TODOS{todo("First")}},
			secondDir: &listProvider{items: types.TODOS{todo("Second")}},
		}}

		listed, err := ListWorkspaces(context.Background(), []Workspace{
			{Name: "first", Dir: firstDir},
			{Name: "second", Dir: secondDir},
			{Name: "alias of first", Dir: firstDir + "/"},
		}, o.open, todos.DiscoveryFilters{})

		Expect(err).NotTo(HaveOccurred())
		Expect(o.opened).To(Equal(map[string]int{firstDir: 1, secondDir: 1}))
		Expect(listed).To(ConsistOf(
			SatisfyAll(
				HaveField("Title", "First"),
				HaveField("Workspace", "first"),
				HaveField("CWD", firstDir),
			),
			SatisfyAll(
				HaveField("Title", "Second"),
				HaveField("Workspace", "second"),
				HaveField("CWD", secondDir),
			),
		))
	})

	It("returns every open and list failure, each naming its project", func() {
		openFailure := errors.New("database unavailable")
		listFailure := errors.New("list query failed")
		o := &opener{
			failures:  map[string]error{firstDir: openFailure},
			providers: map[string]todos.Provider{secondDir: &listProvider{err: listFailure}},
		}

		_, err := ListWorkspaces(context.Background(), []Workspace{
			{Name: "first", Dir: firstDir},
			{Name: "second", Dir: secondDir},
		}, o.open, todos.DiscoveryFilters{})

		Expect(errors.Is(err, openFailure)).To(BeTrue(), "%v", err)
		Expect(errors.Is(err, listFailure)).To(BeTrue(), "%v", err)
		Expect(err.Error()).To(SatisfyAll(
			ContainSubstring(`open native TODO workspace for project "first"`),
			ContainSubstring(`list native TODOs for project "second"`),
		))
	})

	It("fails loudly on a workspace without a directory but still lists the rest", func() {
		o := &opener{providers: map[string]todos.Provider{
			secondDir: &listProvider{items: types.TODOS{todo("Second")}},
		}}

		listed, err := ListWorkspaces(context.Background(), []Workspace{
			{Name: "blank", Dir: "  "},
			{Name: "second", Dir: secondDir},
		}, o.open, todos.DiscoveryFilters{})

		Expect(err).To(MatchError(ContainSubstring(`project "blank": workspace directory is empty`)))
		Expect(listed).To(HaveLen(1))
		Expect(o.opened).To(Equal(map[string]int{secondDir: 1}))
	})

	It("keeps the stored workspace and cwd for a directory named without a project", func() {
		stored := todo("Stored", func(t *types.TODO) {
			t.Workspace = "recorded"
			t.CWD = "/elsewhere"
		})
		o := &opener{providers: map[string]todos.Provider{firstDir: &listProvider{items: types.TODOS{stored}}}}

		listed, err := ListWorkspaces(context.Background(), []Workspace{{Dir: firstDir}}, o.open, todos.DiscoveryFilters{})

		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ConsistOf(SatisfyAll(
			HaveField("Workspace", "recorded"),
			HaveField("CWD", "/elsewhere"),
		)))
	})
})
