package main

import (
	"context"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel pr search", func() {
	isolate := func(dsn string) {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
	}

	It("lists matching PRs from the persisted index", func() {
		isolate(dbtest.ForGinkgo(dbtest.Options{Name: "gavel_cli_pr_search"}).DSN())
		opened, err := database.Open(context.Background(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		store, err := cache.Open()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(store.Close()).To(Succeed()) })
		Expect(store.UpsertPRs([]cache.PullRequest{
			{Provider: "github", Repo: "acme/widgets", Number: 4, URL: "https://github.com/acme/widgets/pull/4", Title: "Parser panic fix", State: "OPEN", UpdatedAt: time.Now()},
			{Provider: "github", Repo: "acme/widgets", Number: 5, URL: "https://github.com/acme/widgets/pull/5", Title: "Dark mode", State: "OPEN", UpdatedAt: time.Now()},
		})).To(Succeed())

		rows, err := runPRSearch(PRSearchOptions{Args: []string{"parser"}, State: "all", Limit: 20})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(Equal([]prSearchRow{{Number: 4, Repo: "acme/widgets", State: "OPEN", Title: "Parser panic fix", URL: "https://github.com/acme/widgets/pull/4"}}))
	})

	It("fails rather than reporting no matches when there is no database", func() {
		isolate("")
		_, err := runPRSearch(PRSearchOptions{State: "all", Limit: 20})
		Expect(err).To(MatchError(cache.ErrPRIndexUnavailable))
	})
})
