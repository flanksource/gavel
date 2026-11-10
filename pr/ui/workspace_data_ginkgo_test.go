package ui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("workspace startup and invalidation", func() {
	It("bounds batch reads independently of workspace count without writes", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		oldPath := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = oldPath })
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "workspace_batch_reads"}).DSN())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		var projects []Project
		var dirs []string
		for _, name := range []string{"first", "second", "third", "fourth"} {
			dir := GinkgoT().TempDir()
			projects = append(projects, Project{Name: name, Dir: dir})
			dirs = append(dirs, dir)
		}
		Expect(SaveProjects(projects)).To(Succeed())
		server := &Server{}
		Expect(server.StartWorkspaceData(ctx, opened.Gorm())).To(Succeed())
		DeferCleanup(func() { cancel(); server.workspaceData.workers.Wait() })
		previousProvider := openTodoProvider
		openTodoProvider = func(context.Context, string) (todos.Provider, error) {
			return nil, errors.New("HTTP read attempted workspace initialization")
		}
		DeferCleanup(func() { openTodoProvider = previousProvider })
		for _, dir := range dirs {
			_, _, err := server.todoProviderContext(ctx, todoSource{Dir: dir})
			Expect(err).NotTo(HaveOccurred())
			provider, err := server.workspaceReadProvider(ctx, dir)
			Expect(err).NotTo(HaveOccurred())
			_, err = provider.Repository().CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: provider.Workspace().ID, Title: "Independent backlog", Labels: []string{"bug"}})
			Expect(err).NotTo(HaveOccurred())
		}
		params := url.Values{"dir": {dirs[0]}, "sessionId": {uuid.NewString()}}
		response := httptest.NewRecorder()
		server.handleTodoSessionStats(response, httptest.NewRequest(http.MethodGet, "/api/todos/session/stats?"+params.Encode(), nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		type requestKey struct{}
		request := context.WithValue(ctx, requestKey{}, true)
		var mu sync.Mutex
		var statements []string
		record := func(tx *gorm.DB) {
			if tx.Statement.Context.Value(requestKey{}) == true {
				mu.Lock()
				statements = append(statements, tx.Statement.SQL.String())
				mu.Unlock()
			}
		}
		callbacks := opened.Gorm().Callback()
		Expect(callbacks.Row().After("gorm:row").Register("test:workspace-reads", record)).To(Succeed())
		Expect(callbacks.Raw().After("gorm:raw").Register("test:workspace-reads", record)).To(Succeed())
		Expect(callbacks.Create().After("gorm:create").Register("test:workspace-reads", record)).To(Succeed())
		Expect(callbacks.Update().After("gorm:update").Register("test:workspace-reads", record)).To(Succeed())
		Expect(callbacks.Delete().After("gorm:delete").Register("test:workspace-reads", record)).To(Succeed())
		results := server.loadTodoBatch(request, dirs)
		Expect(results).To(HaveLen(len(dirs)))
		for _, result := range results {
			Expect(result.Error).To(BeNil())
			Expect(result.Items).To(HaveLen(1))
		}
		Expect(len(statements)).To(BeNumerically("<=", 8), "%v", statements)
		for _, statement := range statements {
			Expect(statement).To(MatchRegexp(`(?i)^\s*(SELECT|WITH)\b`))
		}
		added := GinkgoT().TempDir()
		projects = append(projects, Project{Name: "added", Dir: added})
		Expect(SaveProjects(projects)).To(Succeed())
		Eventually(func() int {
			rows, err := server.catalogProjects()
			Expect(err).NotTo(HaveOccurred())
			return len(rows)
		}, 5*time.Second).Should(Equal(len(projects)))
		before := server.workspaceData.snapshotCopy().Workspaces[added].Config
		Expect(os.WriteFile(filepath.Join(added, ".gavel.yaml"), []byte("ai:\n  model: example-model\n"), 0o600)).To(Succeed())
		Eventually(func() uint64 { return server.workspaceData.snapshotCopy().Workspaces[added].Config }, 5*time.Second).Should(BeNumerically(">", before))
		Expect(SaveProjects(nil)).To(Succeed())
		Eventually(func() int { return len(server.workspaceData.snapshotCopy().Workspaces) }, 5*time.Second).Should(BeZero())
	})
	It("keeps reads immutable and publishes only changed workspaces after commit", func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		oldPath := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = oldPath })
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "workspace_startup"}).DSN())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		first, second := GinkgoT().TempDir(), GinkgoT().TempDir()
		Expect(SaveProjects([]Project{{Name: "first", Dir: first}, {Name: "second", Dir: second}})).To(Succeed())
		server := &Server{}
		Expect(server.StartWorkspaceData(ctx, opened.Gorm())).To(Succeed())
		DeferCleanup(func() { cancel(); server.workspaceData.workers.Wait() })
		provider, err := server.workspaceReadProvider(ctx, first)
		Expect(err).NotTo(HaveOccurred())
		before := server.workspaceData.snapshotCopy()
		unlinked, err := server.workspaceData.affectedWorkspaces(ctx, workspaceChange{RunIDs: []uuid.UUID{uuid.New()}})
		Expect(err).NotTo(HaveOccurred())
		Expect(unlinked).NotTo(BeNil())
		Expect(unlinked).To(BeEmpty())
		_, err = provider.List(ctx, todos.DiscoveryFilters{})
		Expect(err).NotTo(HaveOccurred())
		Expect(server.workspaceData.snapshotCopy()).To(Equal(before))
		rollback := errors.New("rollback fixture")
		Expect(opened.Gorm().Transaction(func(tx *gorm.DB) error {
			repository, err := native.NewRepository(tx)
			Expect(err).NotTo(HaveOccurred())
			_, err = repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: provider.Workspace().ID, Title: "Rolled back"})
			Expect(err).NotTo(HaveOccurred())
			return rollback
		})).To(MatchError(rollback))
		Consistently(server.workspaceData.snapshotCopy, 400*time.Millisecond).Should(Equal(before))
		issue, err := provider.Repository().CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: provider.Workspace().ID, Title: "Committed"})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() uint64 { return server.workspaceData.snapshotCopy().Workspaces[first].Revision }, 5*time.Second).Should(BeNumerically(">", before.Workspaces[first].Revision))
		Expect(server.workspaceData.snapshotCopy().Workspaces[second].Revision).To(Equal(before.Workspaces[second].Revision))
		current := server.workspaceData.snapshotCopy()
		Expect(server.workspaceData.refresh(ctx)).To(Succeed())
		Expect(server.workspaceData.snapshotCopy()).To(Equal(current))
		_, err = provider.Repository().SetLabelDefinition(ctx, native.LabelDefinitionInput{Name: "bug", Color: "red"})
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() uint64 { return server.workspaceData.snapshotCopy().Workspaces[second].Labels }, 5*time.Second).Should(BeNumerically(">", current.Workspaces[second].Labels))
		session, err := provider.Captain().CreateOrGetSession(ctx, captaindb.CreateSessionInput{ID: uuid.New(), Source: "gavel"})
		Expect(err).NotTo(HaveOccurred())
		run, err := provider.Captain().CreatePromptRun(ctx, captaindb.CreatePromptRunInput{SessionID: session.ID})
		Expect(err).NotTo(HaveOccurred())
		_, err = provider.Repository().LinkPromptRun(ctx, issue.ID, run.ID, native.StepRun, 0, issue.Version, "fixture")
		Expect(err).NotTo(HaveOccurred())
		for _, change := range []workspaceChange{{RunIDs: []uuid.UUID{run.ID}}, {SessionIDs: []uuid.UUID{session.ID}}} {
			Expect(server.workspaceData.affectedWorkspaces(ctx, change)).To(Equal([]uuid.UUID{provider.Workspace().ID}))
		}
	})
})
