package entity

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/flanksource/clicky"
	clickyentity "github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fixedListProvider answers List with a fixed set of TODOs; a list reaches
// nothing else on the Provider.
type fixedListProvider struct {
	todos.Provider
	items types.TODOS
}

func (p fixedListProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	// Fresh copies, so tagging rows with their project in one spec cannot leak
	// into the next.
	out := make(types.TODOS, len(p.items))
	for i, item := range p.items {
		copied := *item
		out[i] = &copied
	}
	return out, nil
}

func listedTodo(id, title string, priority types.Priority, labels ...string) *types.TODO {
	todo := &types.TODO{ID: id + "-5f56-4c2c-833f-83e9b7774657", ShortID: id, Labels: labels}
	todo.Title = title
	todo.Status = types.StatusPending
	todo.Priority = priority
	todo.MarkdownBody = "A body far longer than any list row should carry."
	return todo
}

func titles(page clicky.PagedResult[Summary]) []string {
	out := make([]string, len(page.Data))
	for i, summary := range page.Data {
		out[i] = summary.Title
	}
	return out
}

func expectBadRequest(err error, message string) {
	var status *clickyentity.StatusError
	Expect(errors.As(err, &status)).To(BeTrue(), "want a status error, got %v", err)
	Expect(status.Status).To(Equal(http.StatusBadRequest))
	Expect(status.Message).To(ContainSubstring(message))
}

var _ = Describe("Listing TODOs across registered projects", func() {
	const gavelDir, clickyDir, looseDir = "/work/gavel", "/work/clicky-ui", "/work/loose"

	var deps Deps
	var registered []query.Workspace

	BeforeEach(func() {
		providers := map[string]todos.Provider{
			gavelDir: fixedListProvider{items: types.TODOS{
				listedTodo("6408dc66", "Wire the chat picker", types.PriorityHigh, "ui", "chat"),
				listedTodo("33e85357", "Tidy the projects bar", types.PriorityLow),
			}},
			clickyDir: fixedListProvider{items: types.TODOS{listedTodo("e28bd740", "Document route.Router", types.PriorityMedium)}},
			looseDir:  fixedListProvider{items: types.TODOS{listedTodo("d4bb0afb", "Loose checkout work", types.PriorityMedium)}},
		}
		registered = []query.Workspace{{Name: "Gavel", Dir: gavelDir}, {Name: "Clicky UI", Dir: clickyDir}}
		deps = testDeps()
		deps.OpenProvider = func(_ context.Context, dir string) (todos.Provider, error) {
			provider, ok := providers[dir]
			if !ok {
				return nil, errors.New("no workspace at " + dir)
			}
			return provider, nil
		}
		deps.Workspaces = func(context.Context) ([]query.Workspace, error) { return registered, nil }
	})

	It("reads every registered project when the request names no scope", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Limit: 50})

		Expect(err).NotTo(HaveOccurred())
		Expect(page.Data).To(ConsistOf(
			SatisfyAll(HaveField("Title", "Wire the chat picker"), HaveField("Project", "gavel")),
			SatisfyAll(HaveField("Title", "Tidy the projects bar"), HaveField("Project", "gavel")),
			SatisfyAll(HaveField("Title", "Document route.Router"), HaveField("Project", "clicky-ui")),
		))
		Expect(page.Page.Total).To(BeEquivalentTo(3))
	})

	It("returns each row as its short id, title, status, priority, labels and short project name only", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Project: "gavel", Search: "chat picker", Limit: 50})
		Expect(err).NotTo(HaveOccurred())

		encoded, err := json.Marshal(page.Data)

		Expect(err).NotTo(HaveOccurred())
		Expect(encoded).To(MatchJSON(`[{
			"id": "6408dc66",
			"title": "Wire the chat picker",
			"status": "pending",
			"priority": "high",
			"labels": ["ui", "chat"],
			"project": "gavel"
		}]`))
	})

	It("reads only the project named by its short name", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Project: "clicky-ui", Limit: 50})

		Expect(err).NotTo(HaveOccurred())
		Expect(titles(page)).To(Equal([]string{"Document route.Router"}))
	})

	It("does not accept a project's full name in place of its short name", func() {
		_, err := deps.list(context.Background(), query.ListOpts{Project: "Clicky UI", Limit: 50})

		expectBadRequest(err, `unknown project "Clicky UI"; registered projects: gavel, clicky-ui`)
	})

	It("reads only the named directory, registered or not", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Dir: looseDir, Limit: 50})

		Expect(err).NotTo(HaveOccurred())
		Expect(titles(page)).To(Equal([]string{"Loose checkout work"}))
	})

	It("rejects naming both a directory and a project", func() {
		_, err := deps.list(context.Background(), query.ListOpts{Dir: gavelDir, Project: "gavel", Limit: 50})

		expectBadRequest(err, "either dir or project")
	})

	It("refuses to list when two projects share a short name", func() {
		registered = append(registered, query.Workspace{Name: "clicky-ui", Dir: looseDir})

		_, err := deps.list(context.Background(), query.ListOpts{Limit: 50})

		Expect(err).To(MatchError(ContainSubstring(`projects "Clicky UI" and "clicky-ui" share the short name "clicky-ui"`)))
	})

	It("still narrows by severity and title after reading every project", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Priority: []string{"high", "medium"}, Search: "o", Limit: 50})

		Expect(err).NotTo(HaveOccurred())
		Expect(titles(page)).To(ConsistOf("Document route.Router"))
	})

	It("windows the matches and reports how many matched in total", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Limit: 1, Offset: 1})

		Expect(err).NotTo(HaveOccurred())
		Expect(page.Data).To(HaveLen(1))
		Expect(page.Page.Limit).To(Equal(1))
		Expect(page.Page.Offset).To(Equal(1))
		Expect(page.Page.Total).To(BeEquivalentTo(3))
	})

	It("returns an empty window past the last match", func() {
		page, err := deps.list(context.Background(), query.ListOpts{Limit: 10, Offset: 10})

		Expect(err).NotTo(HaveOccurred())
		Expect(page.Data).To(BeEmpty())
		Expect(page.Page.Total).To(BeEquivalentTo(3))
	})

	It("rejects a window that cannot hold a row", func() {
		_, err := deps.list(context.Background(), query.ListOpts{Limit: 0})

		expectBadRequest(err, "limit must be at least 1")
	})

	// clicky reads a table's columns off a zero row when a page is empty.
	It("renders a zero summary without panicking", func() {
		Expect(Summary{}.PrettyRow(nil)).To(HaveKey("Project"))
	})
})
