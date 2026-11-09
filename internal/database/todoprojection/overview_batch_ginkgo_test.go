package todoprojection

import (
	"context"
	"github.com/flanksource/clicky/task"
	"time"

	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
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
		firstProvider, err := todoruntime.New(ctx, db, todoruntime.WorkspaceOptions{Name: "First", RootPath: "/work/first"})
		Expect(err).NotTo(HaveOccurred())
		secondProvider, err := todoruntime.New(ctx, db, todoruntime.WorkspaceOptions{Name: "Second", RootPath: "/work/second"})
		Expect(err).NotTo(HaveOccurred())
		tracked := len(task.RunsRaw(task.RunFilter{}))
		lists, err := todoruntime.ListWorkspaces(ctx, []*todoruntime.Provider{firstProvider, secondProvider})
		Expect(err).NotTo(HaveOccurred())
		Expect(task.RunsRaw(task.RunFilter{})).To(HaveLen(tracked), "routine workspace reads must not register background tasks")
		Expect(lists[first.ID].Err).NotTo(HaveOccurred())
		Expect(lists[first.ID].Items).To(BeEmpty())
		Expect(lists[second.ID].Err).NotTo(HaveOccurred())
		Expect(lists[second.ID].Items).To(HaveLen(1))
		Expect(lists[second.ID].Items[0].Title).To(Equal("Counted"))
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
})
