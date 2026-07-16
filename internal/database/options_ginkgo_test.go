package database

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("Open options", func() {
	var (
		calls  []string
		opened *gorm.DB
		deps   openDependencies
	)

	BeforeEach(func() {
		calls = nil
		opened = &gorm.DB{}
		deps = openDependencies{
			disabled: func() (string, bool) { return "", false },
			resolve:  func() (string, string, error) { return "postgres://gavel", "test", nil },
			migrate: func(_ context.Context, dsn string) error {
				calls = append(calls, "migrate:"+dsn)
				return nil
			},
			open: func(dsn string, _ *gorm.Config) (*gorm.DB, error) {
				calls = append(calls, "open:"+dsn)
				return opened, nil
			},
		}
	})

	It("opens without migrating by default", func(ctx SpecContext) {
		db, err := open(ctx, deps)

		Expect(err).NotTo(HaveOccurred())
		Expect(db.Gorm()).To(BeIdenticalTo(opened))
		Expect(db.migrated).To(BeFalse())
		Expect(calls).To(Equal([]string{"open:postgres://gavel"}))
	})

	It("migrates before opening when explicitly requested", func(ctx SpecContext) {
		db, err := open(ctx, deps, WithMigrations())

		Expect(err).NotTo(HaveOccurred())
		Expect(db.Gorm()).To(BeIdenticalTo(opened))
		Expect(db.migrated).To(BeTrue())
		Expect(calls).To(Equal([]string{"migrate:postgres://gavel", "open:postgres://gavel"}))
	})

	It("does not open after migration fails", func(ctx SpecContext) {
		migrationErr := errors.New("migration failed")
		deps.migrate = func(context.Context, string) error { return migrationErr }

		_, err := open(ctx, deps, WithMigrations())

		Expect(err).To(MatchError(migrationErr))
		Expect(calls).To(BeEmpty())
	})
})

var _ = Describe("Shared migration policy", func() {
	BeforeEach(func() {
		processDB.Lock()
		processDB.db = nil
		processDB.Unlock()
	})

	AfterEach(func() {
		processDB.Lock()
		processDB.db = nil
		processDB.Unlock()
	})

	It("rejects migration after a non-migrating handle was installed", func(ctx SpecContext) {
		processDB.Lock()
		processDB.db = &DB{gorm: &gorm.DB{}, shared: true}
		processDB.Unlock()

		_, err := Shared(ctx, WithMigrations())

		Expect(err).To(MatchError("gavel serve cannot migrate after the process database was opened without migrations"))
	})
})
