package agents_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Comment agents of a project's dashboard pages", func() {
	It("routes a /pages/ thread through the page entry saved in the project's layer, in its delivery group", func(ctx SpecContext) {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comment_agent_pages"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir, err := projects.CanonicalPath(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(dir, verify.GavelConfigFileName), []byte("comments:\n  agents:\n    default:\n      prompt: file-agent\n      edits: propose\n"), 0o600)).To(Succeed())
		catalog := projects.NewDB(db.Gorm(), filepath.Join(dir, "no-projects.json"))
		_, err = catalog.Create(ctx, projects.Project{Name: "acme", Kind: projects.KindDir, Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		configured, err := settings.NewDB(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		routes := agents.ProjectRoutes(catalog, configured)

		file, err := routes.Agent(ctx, comments.Target{ScopeURI: scope, URI: scope + "/files/" + verify.GavelConfigFileName})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{file.Agent.Agent.Prompt, file.Agent.Agent.Group, file.Agent.Agent.Edits}).To(Equal([]any{"file-agent", "", verify.CommentEditsPropose}))
		Expect(file.Agent.Provenance["edits"]).To(Equal(file.Agent.Provenance["prompt"]), "edits is traced to the .gavel.yaml layer that set it")

		key := url.PathEscape("https://example.com/docs#intro")
		target := comments.Target{ScopeURI: scope, URI: scope + "/pages/" + key, Kind: "web-page"}
		layer, err := json.Marshal(settings.Ref{Kind: settings.KindProject})
		Expect(err).NotTo(HaveOccurred())
		enabled := true
		agent, err := json.Marshal(verify.CommentAgent{Enabled: &enabled, Prompt: "page-agent", Group: "docs-review", Edits: verify.CommentEditsPropose})
		Expect(err).NotTo(HaveOccurred())
		_, err = routes.SaveCommentAgent(ctx, comments.SaveCommentAgentRequest{Target: target, Layer: layer, Page: true, Agent: agent})
		Expect(err).NotTo(HaveOccurred())

		route, ok, err := routes.RouteFor(ctx, comments.Comment{Target: target, WorkflowID: comments.HumanReview().ID})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{ok, route}).To(Equal([]any{true, comments.Route{Source: "db:project:", Target: comments.DeliverTarget{Kind: "run", Prompt: "page-agent", GroupKey: "docs-review", Edits: comments.EditsPropose}}}))
		var synced int64
		Expect(db.Gorm().Raw("SELECT count(*) FROM gavel_settings WHERE file IS NOT NULL").Scan(&synced).Error).To(Succeed())
		Expect(synced).To(Equal(int64(1)), "resolving a page keeps the project's synced .gavel.yaml row")
	})
})
