package database

import (
	"context"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/flanksource/gavel/todos/types"
)

const prLinkBackfillMigration = "160_backfill_todo_pr_links.sql"

type backfilledPRLink struct {
	IssueID  uuid.UUID
	Provider string
	Repo     string
	Number   int
	URL      string
	Relation string
	Source   string
}

var _ = Describe("todo PR links schema", Ordered, func() {
	var db *gorm.DB
	var workspaceID uuid.UUID

	BeforeAll(func() {
		ctx := context.Background()
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_pr_links"}).DSN()
		GinkgoT().Setenv(EnvDSN, dsn)
		GinkgoT().Setenv(EnvDisable, "")
		GinkgoT().Setenv(LegacyEnvDSN, "")
		GinkgoT().Setenv(LegacyEnvDisable, "")
		opened, err := Open(ctx, WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db = opened.Gorm()
		workspaceID = uuid.New()
		Expect(db.Exec(`INSERT INTO todo_workspaces (id, repo_key) VALUES (?, 'github.com/acme/pr-links')`, workspaceID).Error).To(Succeed())
	})

	insertIssue := func(title, prURL string) uuid.UUID {
		GinkgoHelper()
		id := uuid.New()
		if prURL == "" {
			Expect(db.Exec(`INSERT INTO todo_issues (id, workspace_id, title) VALUES (?, ?, ?)`, id, workspaceID, title).Error).To(Succeed())
			return id
		}
		Expect(db.Exec(`INSERT INTO todo_issues (id, workspace_id, title, pr) VALUES (?, ?, ?, jsonb_build_object('number', 1, 'url', ?::text))`,
			id, workspaceID, title, prURL).Error).To(Succeed())
		return id
	}

	rerunBackfill := func() {
		GinkgoHelper()
		Expect(db.Exec(`DELETE FROM schema_migration_scripts WHERE scope = 'gavel' AND path = ?`, prLinkBackfillMigration).Error).To(Succeed())
		reopened, err := Open(context.Background(), WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		Expect(reopened.Close()).To(Succeed())
	}

	It("backfills each primary PR as an implements link keyed exactly as PRRefFromURL keys it", func() {
		urls := []string{
			"https://github.com/Acme/Widgets/pull/12",
			"https://dev.azure.com/Org/Product/_git/Service/pullrequest/28",
			"https://dev.azure.com/org/_git/service/pullrequest/3",
			"https://org.visualstudio.com/DefaultCollection/product/_git/service/pullrequest/27",
			"https://org.visualstudio.com/product/_git/other/pullrequest/5",
		}
		want := map[uuid.UUID]backfilledPRLink{}
		for _, raw := range urls {
			ref, err := types.PRRefFromURL(raw)
			Expect(err).NotTo(HaveOccurred())
			id := insertIssue("linked "+raw, raw)
			want[id] = backfilledPRLink{IssueID: id, Provider: ref.Provider, Repo: ref.Repo, Number: ref.Number,
				URL: raw, Relation: "implements", Source: "manual"}
		}
		unlinked := insertIssue("no PR", "")

		rerunBackfill()

		var rows []backfilledPRLink
		Expect(db.Raw(`SELECT issue_id, provider, repo, number, url, relation, source FROM todo_pr_links WHERE workspace_id = ?`, workspaceID).Scan(&rows).Error).To(Succeed())
		got := map[uuid.UUID]backfilledPRLink{}
		for _, row := range rows {
			got[row.IssueID] = row
		}
		Expect(got).To(Equal(want))
		Expect(got).NotTo(HaveKey(unlinked))

		rerunBackfill()
		var count int64
		Expect(db.Table("todo_pr_links").Where("workspace_id = ?", workspaceID).Count(&count).Error).To(Succeed())
		Expect(count).To(Equal(int64(len(urls))), "re-running the backfill must not duplicate links")
	})

	It("rejects an unnormalized key, an unknown relation and an unknown source", func() {
		issue := insertIssue("constraints", "")
		insert := func(repo, relation, source string) error {
			return db.Exec(`INSERT INTO todo_pr_links (workspace_id, issue_id, provider, repo, number, url, relation, source)
				VALUES (?, ?, 'github', ?, 7, 'https://github.com/acme/widgets/pull/7', ?, ?)`, workspaceID, issue, repo, relation, source).Error
		}
		Expect(insert("Acme/Widgets", "related", "manual")).To(HaveOccurred())
		Expect(insert("acme/widgets", "fixes", "manual")).To(HaveOccurred())
		Expect(insert("acme/widgets", "related", "bot")).To(HaveOccurred())
		Expect(insert("acme/widgets", "related", "triage")).To(Succeed())
	})

	It("cascades links when an issue is deleted", func() {
		issue := insertIssue("deleted", "https://github.com/acme/widgets/pull/40")
		rerunBackfill()
		Expect(db.Exec(`DELETE FROM todo_issues WHERE id = ?`, issue).Error).To(Succeed())
		var count int64
		Expect(db.Table("todo_pr_links").Where("issue_id = ?", issue).Count(&count).Error).To(Succeed())
		Expect(count).To(BeZero())
	})

	It("stores duplicates once per pair, lower id first", func() {
		low, high := insertIssue("dup a", ""), insertIssue("dup b", "")
		if low.String() > high.String() {
			low, high = high, low
		}
		insert := func(issue, target uuid.UUID) error {
			return db.Exec(`INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
				VALUES (?, ?, ?, 'duplicates')`, workspaceID, issue, target).Error
		}
		Expect(insert(high, low)).To(HaveOccurred())
		Expect(insert(low, high)).To(Succeed())
	})
})
