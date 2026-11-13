package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/filters"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// orgListedTodo is the part of a flat todo list item the org list specs assert on.
type orgListedTodo struct {
	Title string        `json:"title"`
	Dir   string        `json:"dir"`
	Diff  *todoDiffStat `json:"diff"`
}

type orgTodoList struct {
	Counts todoCounts      `json:"counts"`
	Items  []orgListedTodo `json:"items"`
}

var _ = Describe("todo org list API", func() {
	const alphaRoot, betaRoot, hiddenRoot = "/work/alpha", "/work/beta", "/work/hidden"
	var server *Server

	BeforeEach(func() {
		ctx := context.Background()
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "todo_org_list"}).DSN())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		repository, err := native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		issues := map[string]uuid.UUID{}
		workspaces := map[string]uuid.UUID{}
		for title, root := range map[string]string{"Alpha task": alphaRoot, "Beta task": betaRoot, "Hidden task": hiddenRoot} {
			workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{RootPath: root})
			Expect(err).NotTo(HaveOccurred())
			issue, err := repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: title, Status: native.StatusOpen})
			Expect(err).NotTo(HaveOccurred())
			workspaces[root], issues[title] = workspace.ID, issue.ID
		}
		store := filters.NewStore(opened.Gorm())
		_, err = store.Add(ctx, tenant.Local(), filters.HiddenProjects, hiddenRoot)
		Expect(err).NotTo(HaveOccurred())
		server = &Server{}
		server.SetFilterStore(store)
		// Only alpha's diffs are prepared: beta's run range has never been compared.
		server.workspaceData = &workspaceData{db: opened.Gorm(), org: tenant.LocalOrgID, diffs: map[uuid.UUID]map[string]gavelgit.DiffStat{
			workspaces[alphaRoot]: {issues["Alpha task"].String(): {Commits: 2, Files: 3, Adds: 10, Dels: 4}},
		}}
		// No git on PATH: the list must not compare a range, or track a repository, to answer.
		GinkgoT().Setenv("PATH", GinkgoT().TempDir())
	})

	get := func(query string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		server.handleTodos(recorder, httptest.NewRequest(http.MethodGet, "/api/todos?"+query, nil))
		return recorder
	}

	It("lists every workspace's todos except the hidden ones, with each item's dir and prepared diff", func() {
		recorder := get("exclude=" + filters.HiddenProjects)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var list orgTodoList
		Expect(json.Unmarshal(recorder.Body.Bytes(), &list)).To(Succeed())
		Expect(list).To(Equal(orgTodoList{
			Counts: todoCounts{Total: 2, Open: 2, Pending: 2},
			Items: []orgListedTodo{
				{Title: "Alpha task", Dir: alphaRoot, Diff: &todoDiffStat{Commits: 2, Files: 3, Adds: 10, Dels: 4}},
				{Title: "Beta task", Dir: betaRoot},
			},
		}))
	})

	DescribeTable("refuses an exclude that does not name a workspace filter",
		func(exclude, message string) {
			recorder := get("exclude=" + exclude)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(message))
		},
		Entry("an unknown filter", "no-such-filter", "unknown filter"),
		Entry("a filter of GitHub orgs", filters.HiddenGitHubOrgs, "holds github-org values"),
	)
})
