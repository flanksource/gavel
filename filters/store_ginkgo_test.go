// Tests the named filter store against a migrated database: well-known defaults,
// idempotent edits, per-owner isolation (including the local null owner) and
// rejected names and values.
package filters_test

import (
	"context"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/filters"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("filters.Store", Ordered, func() {
	var (
		store    *filters.Store
		ctx      context.Context
		org      uuid.UUID
		ada, bob tenant.Scope
		local    = tenant.Local()
	)

	BeforeAll(func() {
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_filters"}).DSN())
		opened, err := database.Open(context.Background(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db := opened.Gorm()
		org = uuid.New()
		Expect(db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id) VALUES (?, 'https://clerk.example', 'org_acme')`, org).Error).To(Succeed())
		adaID, bobID := uuid.New(), uuid.New()
		for id, clerk := range map[uuid.UUID]string{adaID: "user_ada", bobID: "user_bob"} {
			Expect(db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, clerk).Error).To(Succeed())
		}
		ada, bob = tenant.Scope{OrgID: org, UserID: &adaID}, tenant.Scope{OrgID: org, UserID: &bobID}
		store, ctx = filters.NewStore(db), context.Background()
	})

	It("returns an empty filter of its kind for a well-known name with no row", func() {
		got, err := store.Get(ctx, local, filters.HiddenGitHubOrgs)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(filters.Filter{Name: filters.HiddenGitHubOrgs, Kind: filters.KindGitHubOrg, Values: []string{}}))
	})

	It("adds and removes values idempotently, keeping them sorted and trimmed", func() {
		for _, value := range []string{"/src/zeta", " /src/alpha ", "/src/zeta", "/src/beta/../alpha"} {
			_, err := store.Add(ctx, local, filters.HiddenProjects, value)
			Expect(err).NotTo(HaveOccurred())
		}
		got, err := store.Remove(ctx, local, filters.HiddenProjects, "/src/zeta")
		Expect(err).NotTo(HaveOccurred())
		again, err := store.Remove(ctx, local, filters.HiddenProjects, "/src/zeta")
		Expect(err).NotTo(HaveOccurred())

		want := filters.Filter{Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{"/src/alpha"}}
		Expect([]filters.Filter{got, again}).To(Equal([]filters.Filter{want, want}))
	})

	It("keeps each owner's filters apart, the local null owner included", func() {
		_, err := store.Add(ctx, ada, filters.HiddenGitHubOrgs, "acme-legacy")
		Expect(err).NotTo(HaveOccurred())
		_, err = store.Add(ctx, bob, filters.HiddenGitHubOrgs, "example-sandbox")
		Expect(err).NotTo(HaveOccurred())
		_, err = store.Add(ctx, local, filters.HiddenGitHubOrgs, "tenant-x")
		Expect(err).NotTo(HaveOccurred())

		lists := map[string][]filters.Filter{}
		for name, scope := range map[string]tenant.Scope{"ada": ada, "bob": bob, "local": local} {
			lists[name], err = store.List(ctx, scope)
			Expect(err).NotTo(HaveOccurred())
		}
		projects := func(values ...string) filters.Filter {
			return filters.Filter{Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: append([]string{}, values...)}
		}
		orgs := func(values ...string) filters.Filter {
			return filters.Filter{Name: filters.HiddenGitHubOrgs, Kind: filters.KindGitHubOrg, Values: values}
		}
		Expect(lists).To(Equal(map[string][]filters.Filter{
			"ada":   {orgs("acme-legacy"), projects()},
			"bob":   {orgs("example-sandbox"), projects()},
			"local": {orgs("tenant-x"), projects("/src/alpha")},
		}))
	})

	It("rejects unknown names and invalid values", func() {
		_, getErr := store.Get(ctx, local, "favourite-projects")
		_, addErr := store.Add(ctx, local, "favourite-projects", "/src/alpha")
		_, blankErr := store.Add(ctx, local, filters.HiddenGitHubOrgs, "  ")
		_, relativeErr := store.Add(ctx, local, filters.HiddenProjects, "src/alpha")

		Expect(getErr).To(MatchError(filters.ErrUnknownFilter))
		Expect(getErr).To(MatchError(ContainSubstring("favourite-projects")))
		Expect(addErr).To(MatchError(filters.ErrUnknownFilter))
		Expect(blankErr).To(MatchError(filters.ErrInvalidValue))
		Expect(relativeErr).To(MatchError(filters.ErrInvalidValue))
	})

	It("accepts a hosted repo:<id> workspace value as given", func() {
		value := "repo:" + uuid.NewString()
		got, err := store.Add(ctx, ada, filters.HiddenProjects, value)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values).To(Equal([]string{value}))
	})
})
