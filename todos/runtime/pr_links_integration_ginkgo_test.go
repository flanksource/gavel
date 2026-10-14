package runtime

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("runtime PR links and duplicates", Ordered, func() {
	const (
		primaryURL = "https://github.com/acme/links/pull/5"
		relatedURL = "https://github.com/acme/links/pull/9"
	)
	var (
		ctx      = context.Background()
		provider *Provider
	)

	BeforeAll(func() {
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_runtime_pr_links"}).DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		root := filepath.Join(GinkgoT().TempDir(), "workspace")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{Name: "PR links", RootPath: root, Repositories: []string{"acme/links"}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("links, lists, reverse-looks-up and unlinks PRs, keeping todo.PR in step", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Fix links", PR: &types.PR{Number: 5, URL: primaryURL}})
		Expect(err).NotTo(HaveOccurred())

		link, err := provider.LinkPR(ctx, todo, todos.PRLinkRequest{URL: relatedURL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage})
		Expect(err).NotTo(HaveOccurred())
		Expect(link.PRRef).To(Equal(types.PRRef{Provider: types.PRProviderGitHub, Repo: "acme/links", Number: 9, URL: relatedURL}))

		links, err := provider.PRLinks(ctx, todo)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(2))
		Expect(links[0].URL).To(Equal(primaryURL))
		Expect(links[0].Relation).To(Equal(types.PRRelationImplements))

		linked, err := provider.TodosForPR(ctx, relatedURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(linked).To(Equal([]todos.PRLinkedTODO{{
			ID: todo.ID, ShortID: todo.ID[:8], Title: "Fix links", Status: todo.Status, Priority: todo.Priority,
			Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage,
		}}))

		Expect(provider.UnlinkPR(ctx, todo, primaryURL)).To(Succeed())
		Expect(todo.PR).To(BeNil(), "removing the primary PR's link clears todo.PR on the reloaded TODO")
		_, err = provider.LinkPR(ctx, todo, todos.PRLinkRequest{URL: "not a url", Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual})
		Expect(err).To(MatchError(ContainSubstring("PR link requires an HTTPS pull request URL")))
	})

	It("writes and reports duplicates like related_to", func() {
		first, err := provider.Create(ctx, todos.CreateRequest{Title: "Parser crash"})
		Expect(err).NotTo(HaveOccurred())
		second, err := provider.Create(ctx, todos.CreateRequest{Title: "Parser panic"})
		Expect(err).NotTo(HaveOccurred())

		_, err = provider.Link(ctx, second, first.ID, types.RelationDuplicates)
		Expect(err).NotTo(HaveOccurred())
		links, err := provider.Links(ctx, first)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(1))
		Expect(links[0].Relation).To(Equal(types.RelationDuplicates))
		Expect(links[0].TargetID).To(Equal(second.ID))
		Expect(links[0].TargetTitle).To(Equal("Parser panic"))
	})
})
