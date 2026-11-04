package postgres_test

import (
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments/postgres"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fresh provisioning boundary", func() {
	It("refuses to convert an existing foreign comments table", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_foreign_schema"})
		Expect(db.Gorm().Exec("CREATE TABLE comments (id text PRIMARY KEY, body text NOT NULL)").Error).NotTo(HaveOccurred())
		Expect(postgres.Provision(ctx, db.DSN())).To(MatchError(ContainSubstring("existing comments table")))
		Expect(db.Gorm().Exec("INSERT INTO comments (id,body) VALUES ('original','preserved')").Error).NotTo(HaveOccurred())
	})
})
