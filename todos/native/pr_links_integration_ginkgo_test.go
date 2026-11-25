package native_test

import (
	"context"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("issue PR links", Ordered, func() {
	const (
		primaryURL = "https://github.com/acme/widgets/pull/12"
		relatedURL = "https://github.com/acme/widgets/pull/30"
	)
	var (
		ctx       = context.Background()
		repo      *native.Repository
		workspace *native.Workspace
	)

	BeforeAll(func() {
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_issue_pr_links"}).DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		repo, err = native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		workspace, err = repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/widgets"})
		Expect(err).NotTo(HaveOccurred())
	})

	type linkView struct {
		URL      string
		Relation types.PRRelation
		Source   types.PRLinkSource
	}
	links := func(issueID uuid.UUID) []linkView {
		GinkgoHelper()
		stored, err := repo.ListIssuePRLinks(ctx, issueID)
		Expect(err).NotTo(HaveOccurred())
		out := []linkView{}
		for _, link := range stored {
			out = append(out, linkView{URL: link.URL, Relation: link.Relation, Source: link.Source})
		}
		return out
	}
	create := func(title string, pr *types.PR) *native.Issue {
		GinkgoHelper()
		issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: title, PR: pr})
		Expect(err).NotTo(HaveOccurred())
		return issue
	}

	It("mirrors the primary PR as an implements link from creation to clearing", func() {
		issue := create("Fix the parser", &types.PR{Number: 12, URL: primaryURL})
		Expect(links(issue.ID)).To(Equal([]linkView{{URL: primaryURL, Relation: types.PRRelationImplements, Source: types.PRLinkSourceManual}}))

		next := &types.PR{Number: 30, URL: relatedURL}
		updated, err := repo.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{PR: &next})
		Expect(err).NotTo(HaveOccurred())
		Expect(links(issue.ID)).To(Equal([]linkView{{URL: relatedURL, Relation: types.PRRelationImplements, Source: types.PRLinkSourceManual}}))

		var cleared *types.PR
		_, err = repo.UpdateIssue(ctx, issue.ID, updated.Version, native.IssuePatch{PR: &cleared})
		Expect(err).NotTo(HaveOccurred())
		Expect(links(issue.ID)).To(BeEmpty())
	})

	It("links further PRs, refuses a second link to the same PR and lists the issues per PR", func() {
		issue := create("Speed up parsing", &types.PR{Number: 12, URL: primaryURL})
		link, err := repo.LinkIssuePR(ctx, issue.ID, issue.Version, native.PRLinkInput{
			URL: relatedURL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage, Actor: "gavel",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(link.Repo).To(Equal("acme/widgets"))
		Expect(link.Number).To(Equal(30))
		Expect(links(issue.ID)).To(Equal([]linkView{
			{URL: primaryURL, Relation: types.PRRelationImplements, Source: types.PRLinkSourceManual},
			{URL: relatedURL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage},
		}))

		reloaded, err := repo.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.Version).To(Equal(issue.Version+1), "a link is a recorded mutation")
		_, err = repo.LinkIssuePR(ctx, issue.ID, reloaded.Version, native.PRLinkInput{
			URL: "https://github.com/ACME/Widgets/pull/30", Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual,
		})
		Expect(err).To(MatchError(native.ErrPRLinkExists))

		linked, err := repo.ListPRLinkedIssues(ctx, native.PRLinkedIssuesQuery{WorkspaceID: workspace.ID, URL: "https://github.com/Acme/widgets/pull/30"})
		Expect(err).NotTo(HaveOccurred())
		Expect(linked).To(HaveLen(1))
		Expect(linked[0].Issue.ID).To(Equal(issue.ID))
		Expect(linked[0].Relation).To(Equal(types.PRRelationRelated))
		Expect(linked[0].Source).To(Equal(types.PRLinkSourceTriage))
	})

	It("lists the issues linking a PR across every workspace when no workspace is given", func() {
		const sharedURL = "https://github.com/acme/widgets/pull/77"
		local := create("Local work", &types.PR{Number: 77, URL: sharedURL})
		other, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/other"})
		Expect(err).NotTo(HaveOccurred())
		remote, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: other.ID, Title: "Remote work", PR: &types.PR{Number: 77, URL: sharedURL}})
		Expect(err).NotTo(HaveOccurred())

		issueIDs := func(query native.PRLinkedIssuesQuery) []uuid.UUID {
			GinkgoHelper()
			linked, err := repo.ListPRLinkedIssues(ctx, query)
			Expect(err).NotTo(HaveOccurred())
			ids := []uuid.UUID{}
			for _, entry := range linked {
				ids = append(ids, entry.Issue.ID)
			}
			return ids
		}
		Expect(issueIDs(native.PRLinkedIssuesQuery{WorkspaceID: workspace.ID, URL: sharedURL})).To(Equal([]uuid.UUID{local.ID}))
		Expect(issueIDs(native.PRLinkedIssuesQuery{URL: sharedURL})).To(ConsistOf(local.ID, remote.ID))
	})

	It("rejects a URL that is not a hosted PR", func() {
		issue := create("Bad link", nil)
		_, err := repo.LinkIssuePR(ctx, issue.ID, issue.Version, native.PRLinkInput{
			URL: "https://github.com/acme/widgets/issues/4", Relation: types.PRRelationRelated, Source: types.PRLinkSourceManual,
		})
		Expect(err).To(MatchError(native.ErrInvalidInput))
	})

	It("promotes a triage link when the PR becomes primary", func() {
		issue := create("Promoted", nil)
		_, err := repo.LinkIssuePR(ctx, issue.ID, issue.Version, native.PRLinkInput{
			URL: relatedURL, Relation: types.PRRelationRelated, Source: types.PRLinkSourceTriage,
		})
		Expect(err).NotTo(HaveOccurred())
		reloaded, err := repo.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		primary := &types.PR{Number: 30, URL: relatedURL}
		_, err = repo.UpdateIssue(ctx, issue.ID, reloaded.Version, native.IssuePatch{PR: &primary})
		Expect(err).NotTo(HaveOccurred())
		Expect(links(issue.ID)).To(Equal([]linkView{{URL: relatedURL, Relation: types.PRRelationImplements, Source: types.PRLinkSourceManual}}))
	})

	It("clears the primary PR when its link is removed and refuses an unknown link", func() {
		issue := create("Unlink primary", &types.PR{Number: 12, URL: primaryURL})
		Expect(repo.UnlinkIssuePR(ctx, issue.ID, issue.Version, primaryURL, "gavel")).To(Succeed())
		reloaded, err := repo.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.PR).To(BeNil())
		Expect(links(issue.ID)).To(BeEmpty())

		Expect(repo.UnlinkIssuePR(ctx, issue.ID, reloaded.Version, primaryURL, "gavel")).To(MatchError(native.ErrPRLinkNotFound))
	})

	It("records duplicates once per pair and reports it from both ends", func() {
		first, second := create("Parser crash", nil), create("Parser panics", nil)
		_, err := repo.AddRelationship(ctx, second.ID, first.ID, native.RelationshipDuplicates, second.Version, "gavel")
		Expect(err).NotTo(HaveOccurred())
		for _, pair := range [][2]uuid.UUID{{first.ID, second.ID}, {second.ID, first.ID}} {
			relationships, err := repo.ListRelationships(ctx, pair[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(relationships).To(HaveLen(1))
			Expect(relationships[0].Relation).To(Equal(native.RelationshipDuplicates))
			Expect(relationships[0].TargetIssueID).To(Equal(pair[1]))
		}
		reloaded, err := repo.GetIssue(ctx, first.ID)
		Expect(err).NotTo(HaveOccurred())
		_, err = repo.AddRelationship(ctx, first.ID, second.ID, native.RelationshipDuplicates, reloaded.Version, "gavel")
		Expect(err).To(MatchError(native.ErrRelationshipExists))
	})
})
