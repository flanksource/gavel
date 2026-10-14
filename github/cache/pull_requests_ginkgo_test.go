package cache

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCacheSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "github/cache")
}

func indexedPR(repo string, number int, title, state string, updated time.Time) PullRequest {
	return PullRequest{
		Provider: "github", Repo: repo, Number: number, Title: title, State: state, Author: "octocat",
		URL: "https://github.com/" + repo + "/pull/" + strconv.Itoa(number), UpdatedAt: updated,
	}
}

var _ = Describe("the persisted PR index", Ordered, func() {
	var store *Store
	base := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

	BeforeAll(func() {
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_pr_index"}).DSN()
		GinkgoT().Setenv(EnvDatabaseDSN, dsn)
		GinkgoT().Setenv(EnvDatabaseDisable, "")
		GinkgoT().Setenv(EnvDSN, "")
		GinkgoT().Setenv(EnvDisable, "")
		opened, err := database.Open(context.Background(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		store = &Store{db: opened}

		Expect(store.UpsertPRs([]PullRequest{
			indexedPR("Acme/Widgets", 12, "Fix parser panic on unterminated string", "OPEN", base.Add(3*time.Hour)),
			indexedPR("acme/gadgets", 7, "Parser speedup", "MERGED", base.Add(2*time.Hour)),
			indexedPR("acme/widgets", 30, "Add dark mode", "CLOSED", base.Add(1*time.Hour)),
		})).To(Succeed())
	})

	numbers := func(rows []PullRequest) []int {
		out := make([]int, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Number)
		}
		return out
	}

	It("lists the most recently updated PRs when the text is empty", func() {
		rows, err := store.SearchPRs(PRQuery{Limit: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{12, 7}))
	})

	It("matches title words as prefixes", func() {
		rows, err := store.SearchPRs(PRQuery{Text: "pars", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{12, 7}))

		rows, err = store.SearchPRs(PRQuery{Text: "parser unterm", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{12}))
	})

	It("matches a PR number", func() {
		rows, err := store.SearchPRs(PRQuery{Text: "#30", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{30}))
	})

	It("filters by repository case-insensitively and by state", func() {
		rows, err := store.SearchPRs(PRQuery{Repo: "ACME/widgets", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{12, 30}))

		rows, err = store.SearchPRs(PRQuery{State: "merged", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(numbers(rows)).To(Equal([]int{7}))
	})

	It("rejects an unknown state and an out-of-range limit", func() {
		_, err := store.SearchPRs(PRQuery{State: "draft", Limit: 10})
		Expect(err).To(MatchError(ContainSubstring(`unknown PR state "draft"`)))
		Expect(err).To(MatchError(ErrInvalidPRQuery))
		_, err = store.SearchPRs(PRQuery{})
		Expect(err).To(MatchError(ContainSubstring("limit 0")))
		Expect(err).To(MatchError(ErrInvalidPRQuery))
	})

	It("refreshes a re-synced PR in place", func() {
		Expect(store.UpsertPRs([]PullRequest{
			indexedPR("acme/gadgets", 7, "Parser speedup v2", "CLOSED", base.Add(4*time.Hour)),
		})).To(Succeed())
		rows, err := store.SearchPRs(PRQuery{Text: "#7", Limit: 10})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Title).To(Equal("Parser speedup v2"))
		Expect(rows[0].State).To(Equal("CLOSED"))
		Expect(rows[0].UpdatedAt).To(BeTemporally("==", base.Add(4*time.Hour)))
	})

	It("loads rows by case-insensitive key and omits unsynced keys", func() {
		got, err := store.GetPRs([]PRKey{
			{Provider: "github", Repo: "acme/widgets", Number: 12},
			{Provider: "github", Repo: "acme/widgets", Number: 99},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
		Expect(got[PRKey{Provider: "github", Repo: "acme/widgets", Number: 12}].Repo).To(Equal("Acme/Widgets"))
	})

	It("refuses reads but accepts writes on a disabled store", func() {
		disabled := &Store{disabled: true}
		Expect(disabled.UpsertPRs([]PullRequest{indexedPR("acme/widgets", 1, "x", "OPEN", base)})).To(Succeed())
		_, err := disabled.SearchPRs(PRQuery{Limit: 1})
		Expect(err).To(MatchError(ErrPRIndexUnavailable))
		_, err = disabled.GetPRs(nil)
		Expect(err).To(MatchError(ErrPRIndexUnavailable))
	})
})
