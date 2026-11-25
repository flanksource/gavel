package settings_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/settings"
	gaveltenant "github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("database settings", Ordered, func() {
	org := uuid.MustParse("5e771265-0000-4000-8000-000000000001")
	tenant := org.String()
	var (
		db      *gorm.DB
		service *settings.DB
	)
	ctx := gaveltenant.WithScope(context.Background(), gaveltenant.Scope{OrgID: org})

	type stored struct {
		Path   *string
		File   *string
		GitSHA *string
	}
	rowsOf := func(projectID uuid.UUID) []stored {
		GinkgoHelper()
		var rows []stored
		Expect(db.Raw(`SELECT path, file, git_sha FROM gavel_settings WHERE tenant_id = ? AND (project_id = ? OR project_id IS NULL) ORDER BY path NULLS FIRST, file NULLS FIRST`, tenant, projectID).Scan(&rows).Error).To(Succeed())
		return rows
	}
	newTarget := func() (settings.Target, string, string) {
		GinkgoHelper()
		_, root, page := projectTree()
		id := uuid.New()
		Expect(db.Exec(`INSERT INTO projects (id, tenant_id, name, dir) VALUES (?, ?, ?, ?)`, id, tenant, "acme-"+id.String()[:8], root).Error).To(Succeed())
		return settings.Target{Path: page, Project: &settings.Project{ID: id, Dir: root}}, root, page
	}
	text := func(value string) *string { return &value }

	BeforeAll(func(ctx SpecContext) {
		pg := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_settings"})
		Expect(database.Migrate(ctx, pg.DSN())).To(Succeed())
		db = pg.Gorm()
		var err error
		service, err = settings.NewDB(db)
		Expect(err).NotTo(HaveOccurred())
	})

	It("syncs the chain's files into rows at the HEAD they were read at, and drops a row whose file is gone", func() {
		target, root, _ := newTarget()
		head := git(root, "rev-parse", "HEAD")
		writeConfig(root, verify.GavelConfigFileName, "pr:\n  base: origin/main\n")
		writeConfig(filepath.Join(root, "slides"), ".gavel.agents.yaml", "comments:\n  agents:\n    default:\n      enabled: true\n")

		_, err := service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		Expect(rowsOf(target.Project.ID)).To(Equal([]stored{
			{File: text(".gavel.yaml"), GitSHA: &head},
			{Path: text("slides"), File: text(".gavel.agents.yaml"), GitSHA: &head},
		}))

		Expect(os.Remove(filepath.Join(root, "slides", ".gavel.agents.yaml"))).To(Succeed())
		_, err = service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		Expect(rowsOf(target.Project.ID)).To(Equal([]stored{{File: text(".gavel.yaml"), GitSHA: &head}}))
	})

	It("keeps the HEAD a file was first synced at until its config changes", func() {
		target, root, _ := newTarget()
		writeConfig(root, verify.GavelConfigFileName, "pr:\n  draft: true\n")
		first := git(root, "rev-parse", "HEAD")
		_, err := service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())

		git(root, "-c", "user.email=dev@example.com", "-c", "user.name=dev", "commit", "--quiet", "--allow-empty", "-m", "next")
		_, err = service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		unchanged := *rowsOf(target.Project.ID)[0].GitSHA

		writeConfig(root, verify.GavelConfigFileName, "pr:\n  draft: false\n")
		_, err = service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{unchanged, *rowsOf(target.Project.ID)[0].GitSHA}).To(Equal([]string{first, git(root, "rev-parse", "HEAD")}))
	})

	It("authors a database layer at a sub-folder with no file, merges it in its place, and traces fields to it", func() {
		target, root, _ := newTarget()
		writeConfig(root, verify.GavelConfigFileName, "comments:\n  agents:\n    default:\n      enabled: true\n      prompt: editor-comments\n")
		folder := settings.Ref{Kind: settings.KindPath, Path: "slides/arb"}

		Expect(service.Save(ctx, target, folder, func(cfg *verify.GavelConfig) error {
			cfg.Comments.Agents.Default = &verify.CommentAgent{Prompt: "slides-review"}
			return nil
		})).To(Succeed())
		resolved, err := service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		var takesDB bool
		for _, position := range resolved.Positions {
			if position.Kind == settings.KindPath && position.Path == "slides/arb" {
				takesDB = position.DB
			}
		}
		agent := resolved.Merged.Comments.Agents.Default
		Expect([]any{takesDB, *agent.Enabled, agent.Prompt, resolved.Provenance["/comments/agents/default/prompt"], resolved.Provenance["/comments/agents/default/enabled"]}).
			To(Equal([]any{true, true, "slides-review", "db:path:slides/arb", filepath.Join(root, ".gavel.yaml")}))
		Expect(rowsOf(target.Project.ID)).To(ContainElement(stored{Path: text("slides/arb")}))
		var payload string
		Expect(db.Raw(`SELECT payload::text FROM gavel_settings WHERE project_id = ? AND path = 'slides/arb'`, target.Project.ID).Scan(&payload).Error).To(Succeed())
		Expect(payload).To(Equal(`{"comments": {"agents": {"default": {"prompt": "slides-review"}}}}`))
	})

	It("refuses a database layer where a file exists, and saving to that file instead re-syncs it", func() {
		target, root, _ := newTarget()
		writeConfig(filepath.Join(root, "slides"), verify.GavelConfigFileName, "{}\n")
		draft := func(cfg *verify.GavelConfig) error { cfg.PR.Draft = true; return nil }

		Expect(service.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "slides"}, draft)).
			To(MatchError(ContainSubstring("already has [.gavel.yaml]")))
		Expect(service.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "slides", File: ".gavel.pr.yaml"}, draft)).To(Succeed())
		head := git(root, "rev-parse", "HEAD")
		Expect(rowsOf(target.Project.ID)).To(ContainElement(stored{Path: text("slides"), File: text(".gavel.pr.yaml"), GitSHA: &head}))
	})

	It("layers a web project's database layers by URL path segment, and keeps files out of it", func() {
		home, _, _ := projectTree()
		writeConfig(home, verify.GavelConfigFileName, "comments:\n  agents:\n    default:\n      prompt: editor-comments\n")
		id := uuid.New()
		Expect(db.Exec(`INSERT INTO projects (id, tenant_id, name, kind) VALUES (?, ?, ?, 'web')`, id, tenant, "web-"+id.String()[:8]).Error).To(Succeed())
		target := settings.Target{Key: "pages/team/members", Project: &settings.Project{ID: id}}
		enable := func(cfg *verify.GavelConfig) error {
			cfg.Comments.Agents.Default = &verify.CommentAgent{Enabled: ptr(true)}
			return nil
		}

		Expect(service.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "pages/team"}, enable)).To(Succeed())
		resolved, err := service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		refs := []settings.Ref{}
		for _, layer := range resolved.Layers {
			refs = append(refs, layer.Ref)
		}
		agent := resolved.Merged.Comments.Agents.Default
		Expect([]any{refs, *agent.Enabled, agent.Prompt}).To(Equal([]any{
			[]settings.Ref{{Kind: settings.KindUser, File: ".gavel.yaml"}, {Kind: settings.KindPath, Path: "pages/team"}},
			true, "editor-comments",
		}))
		Expect(service.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "pages", File: ".gavel.yaml"}, enable)).
			To(MatchError(ContainSubstring("has no directory")))
	})

	It("ignores a sub-folder's database layer once a file appears there", func() {
		target, root, _ := newTarget()
		Expect(service.Save(ctx, target, settings.Ref{Kind: settings.KindPath, Path: "slides"}, func(cfg *verify.GavelConfig) error {
			cfg.PR.Base = "origin/db"
			return nil
		})).To(Succeed())
		writeConfig(filepath.Join(root, "slides"), verify.GavelConfigFileName, "pr:\n  base: origin/file\n")

		resolved, err := service.Resolve(ctx, target)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Merged.PR.Base).To(Equal("origin/file"))
	})
})
