package workspaceaccess

import (
	"context"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("durable workspace grants", func() {
	It("persists grants across repository instances and scopes revocation", func() {
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "workspace_access"}).DSN()
		GinkgoT().Setenv(database.EnvDSN, dsn)
		db, err := database.Shared(context.Background(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Disabled()).To(BeFalse())
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		root, err := Canonical(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		first := durable{}
		added, err := first.Grant(context.Background(), root)
		Expect(err).NotTo(HaveOccurred())
		roots, err := (durable{}).List(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(roots).To(ContainElement(HaveField("ID", added.ID)))
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		other, err := (durable{}).List(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(other).To(BeEmpty())
		Expect(first.Revoke(context.Background(), added.ID)).To(MatchError("saved grant does not exist"))
		GinkgoT().Setenv("HOME", home)
		Expect(first.Revoke(context.Background(), added.ID)).To(Succeed())
		roots, err = first.List(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(roots).To(BeEmpty())
	})
})
