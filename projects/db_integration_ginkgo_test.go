package projects

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var localTenant = tenant.LocalOrgID.String()

var _ = Describe("DB catalog", Ordered, func() {
	var (
		gormDB     *gorm.DB
		importPath string
	)

	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "projects_catalog"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		gormDB = db.Gorm()
	})

	// emptyCatalog clears every tenant's projects and import record, so each
	// spec starts from a first use.
	emptyCatalog := func() {
		Expect(gormDB.Exec(`TRUNCATE projects, project_imports CASCADE`).Error).To(Succeed())
		importPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	}

	Context("as a catalog", func() {
		catalogContract(func() Service {
			emptyCatalog()
			return NewDB(gormDB, importPath)
		})
	})

	Context("importing projects.json", func() {
		BeforeEach(emptyCatalog)

		It("imports the local catalog once, keeping each project's file-mode id", func(ctx SpecContext) {
			Expect(os.WriteFile(importPath, []byte(`[{"name":"gavel","dir":"/work/gavel","repos":["acme/gavel"]},{"name":"notes","dir":"/work/notes"}]`), 0o600)).To(Succeed())
			fromFile, err := NewFile(importPath).List(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(NewDB(gormDB, importPath).List(ctx)).To(Equal(fromFile))

			// Neither a later edit of the file nor an emptied table re-imports it.
			Expect(NewDB(gormDB, importPath).Delete(ctx, "gavel")).To(Succeed())
			Expect(NewDB(gormDB, importPath).Delete(ctx, "notes")).To(Succeed())
			Expect(os.WriteFile(importPath, []byte(`[{"name":"late","dir":"/work/late"}]`), 0o600)).To(Succeed())
			Expect(NewDB(gormDB, importPath).List(ctx)).To(BeEmpty())
		})

		It("fails the first use on a malformed file instead of starting empty", func(ctx SpecContext) {
			Expect(os.WriteFile(importPath, []byte("{"), 0o600)).To(Succeed())

			_, err := NewDB(gormDB, importPath).List(ctx)

			Expect(err).To(MatchError(ContainSubstring(importPath)))
			var imports int64
			Expect(gormDB.Raw(`SELECT count(*) FROM project_imports`).Scan(&imports).Error).To(Succeed())
			Expect(imports).To(BeZero())
		})
	})

	It("keeps names unique per tenant, not across tenants", func(ctx SpecContext) {
		emptyCatalog()
		catalog := NewDB(gormDB, importPath)
		other := tenant.WithScope(ctx, tenant.Scope{OrgID: uuid.New()})
		local, err := catalog.Create(ctx, Project{Name: "shared", Dir: "/local/shared"})
		Expect(err).NotTo(HaveOccurred())

		hosted, err := catalog.Create(other, Project{Name: "shared", Dir: "/hosted/shared"})

		Expect(err).NotTo(HaveOccurred())
		Expect(hosted.ID).NotTo(Equal(local.ID))
		Expect(catalog.List(other)).To(ConsistOf(HaveField("Dir", "/hosted/shared")))
		Expect(catalog.Get(context.Background(), "shared")).To(HaveField("Dir", "/local/shared"))
	})
})
