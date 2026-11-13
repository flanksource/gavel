package todoprojection

import (
	"context"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO overview batches", func() {
	It("preserves execution-state precedence across run changes", func() {
		db := openMigratedDatabase()
		fixture := newActiveRunFixture(db, time.Now().UTC())
		for _, state := range []string{"running", "waiting", "succeeded", "failed", "cancelled"} {
			Expect(db.Exec(`UPDATE captain_prompt_runs SET state = ? WHERE id = ?`, state, fixture.runID).Error).To(Succeed())
			var row struct {
				Actual   string
				Expected string
			}
			Expect(db.Raw(`SELECT execution_state AS actual, gavel_todo_issue_execution_state(issue_id) AS expected FROM todo_issue_runtime WHERE issue_id = ?`, fixture.issueID).Scan(&row).Error).To(Succeed())
			Expect(row.Actual).To(Equal(row.Expected), state)
		}
	})
	It("combines global and workspace labels with visible usage counts", func() {
		db := openMigratedDatabase()
		repository, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		ctx := context.Background()
		first, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{RootPath: "/work/first", DisplayName: "First"})
		Expect(err).NotTo(HaveOccurred())
		second, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{RootPath: "/work/second", DisplayName: "Second"})
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "bug", Color: "red"})
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.SetLabelDefinition(ctx, native.LabelDefinitionInput{WorkspaceID: &first.ID, Name: "bug", Color: "blue"})
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: second.ID, Title: "Counted", Labels: []string{"bug"}})
		Expect(err).NotTo(HaveOccurred())
		indexes, err := todoruntime.ReadLabelIndexes(ctx, db, []todoruntime.WorkspaceOptions{{Name: "First", RootPath: "/work/first"}, {Name: "Second", RootPath: "/work/second"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(indexes[0].Err).NotTo(HaveOccurred())
		Expect(indexes[1].Counts).To(Equal(map[string]int{"bug": 1}))
		global, err := todoruntime.NewGlobal(db)
		Expect(err).NotTo(HaveOccurred())
		tracked := len(task.RunsRaw(task.RunFilter{}))
		listed, err := global.ListOrg(ctx, todos.DiscoveryFilters{}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(task.RunsRaw(task.RunFilter{})).To(HaveLen(tracked), "routine workspace reads must not register background tasks")
		Expect(listedTodos(listed)).To(Equal([]listedTodo{{Title: "Counted", CWD: "/work/second", Labels: "bug=red"}}))
		for i, color := range []string{"blue", "red"} {
			found := false
			for _, definition := range indexes[i].Definitions {
				if definition.Name == "bug" {
					Expect(string(definition.Color)).To(Equal(color))
					found = true
				}
			}
			Expect(found).To(BeTrue())
		}
	})
	It("lists the org's todos outside the excluded workspaces as the viewer sees them", func() {
		db := openMigratedDatabase()
		repository, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		ctx := context.Background()
		creator, member := uuid.New(), uuid.New()
		for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
			Expect(db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error).To(Succeed())
		}
		asCreator := tenant.WithScope(ctx, tenant.Scope{OrgID: tenant.LocalOrgID, UserID: &creator})
		asMember := tenant.WithScope(ctx, tenant.Scope{OrgID: tenant.LocalOrgID, UserID: &member})
		roots := map[string]string{"alpha": "/work/alpha", "beta": "/work/beta", "hidden": "/work/hidden"}
		for name, root := range roots {
			workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{RootPath: root, DisplayName: name})
			Expect(err).NotTo(HaveOccurred())
			for _, visibility := range []native.Visibility{native.VisibilityPublic, native.VisibilityPrivate} {
				_, err = repository.CreateIssue(asCreator, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: name + " " + string(visibility), Visibility: visibility})
				Expect(err).NotTo(HaveOccurred())
			}
		}
		global, err := todoruntime.NewGlobal(db)
		Expect(err).NotTo(HaveOccurred())

		listed, err := global.ListOrg(asMember, todos.DiscoveryFilters{}, []string{roots["hidden"] + "/"})

		Expect(err).NotTo(HaveOccurred())
		Expect(listedTodos(listed)).To(Equal([]listedTodo{
			{Title: "alpha public", CWD: roots["alpha"]},
			{Title: "beta public", CWD: roots["beta"]},
		}))
	})
})

// listedTodo is the part of a listed TODO the org list specs assert on.
type listedTodo struct {
	Title, CWD, Labels string
}

func listedTodos(items types.TODOS) []listedTodo {
	out := make([]listedTodo, 0, len(items))
	for _, item := range items {
		var labels []string
		for _, definition := range item.LabelDefinitions {
			labels = append(labels, definition.Name+"="+string(definition.Color))
		}
		out = append(out, listedTodo{Title: item.Title, CWD: item.CWD, Labels: strings.Join(labels, ",")})
	}
	return out
}
