package query_test

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/flanksource/clicky/task"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/form"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("database progress examples", func() {
	It("publishes intermediate progress from the PostgreSQL function and retains ordinary output", func() {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "progress_example"})
		path, err := filepath.Abs("../../../examples/postgres-progress.fixture.yaml")
		Expect(err).NotTo(HaveOccurred())
		file, err := fixtures.ParseFixtureFile(path)
		Expect(err).NotTo(HaveOccurred())
		var mu sync.Mutex
		var progress []int
		var statuses []any
		sink := func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
			mu.Lock()
			defer mu.Unlock()
			var visit func(*fixtures.ExecutionNode)
			visit = func(node *fixtures.ExecutionNode) {
				if node.State == fixtures.ExecutionRunning && node.Total == 5 && node.Done > 0 && node.Done < 5 {
					progress = append(progress, node.Done)
					statuses = append(statuses, node.Metadata["status"])
				}
				for _, child := range node.Children {
					visit(child)
				}
			}
			visit(snapshot.Root)
			return nil
		}
		results, _, err := fixtures.RunNodes(fixtures.WithProgressSink(context.Background(), sink), []*fixtures.FixtureNode{file}, fixtures.RunOptions{
			WorkDir: filepath.Dir(path), Params: map[string]any{"connection": "progress-test"},
			Connections: func(_ dbcontext.Context, reference string) (*models.Connection, error) {
				Expect(reference).To(Equal("progress-test"))
				return &models.Connection{Name: reference, Type: models.ConnectionTypePostgres, URL: db.DSN()}, nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
		Expect(progress).To(ContainElement(1))
		Expect(statuses).To(ContainElement("Completed batch 1 of 5"))
		Expect(results[1].Stdout).To(And(ContainSubstring("Completed batch 5 of 5"), Not(ContainSubstring("::progress"))))
		Expect(results[1].Metadata).To(HaveKeyWithValue("description", "PostgreSQL batches complete"))
	})
})
