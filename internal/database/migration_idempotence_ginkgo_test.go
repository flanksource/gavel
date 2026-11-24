package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/flanksource/commons-db/dbtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("shared database migrations", func() {
	It("preserves partial indexes when run again while plans are being read", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_partial_index_migrations"})
		Expect(Migrate(ctx, db.DSN())).To(Succeed())
		before := partialIndexOIDs(ctx, db.SQL())
		Expect(before).To(HaveKey("captain_plans_prompt_variant_key"))
		Expect(before).To(HaveKey("comment_operations_active"))

		reader, err := db.SQL().BeginTx(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		defer func() { Expect(reader.Rollback()).To(Succeed()) }()
		_, err = reader.ExecContext(ctx, `LOCK TABLE public.captain_plans IN ACCESS SHARE MODE`)
		Expect(err).NotTo(HaveOccurred())

		Expect(Migrate(ctx, db.DSN())).To(Succeed())
		Expect(partialIndexOIDs(ctx, db.SQL())).To(Equal(before))
	})
})

func partialIndexOIDs(ctx context.Context, db *sql.DB) map[string]int64 {
	GinkgoHelper()
	var raw []byte
	Expect(db.QueryRowContext(ctx, `
		SELECT jsonb_object_agg(idx.relname, idx.oid::bigint)
		FROM pg_index i
		JOIN pg_class idx ON idx.oid = i.indexrelid
		JOIN pg_namespace ns ON ns.oid = idx.relnamespace
		WHERE ns.nspname = 'public' AND i.indpred IS NOT NULL
	`).Scan(&raw)).To(Succeed())
	var oids map[string]int64
	Expect(json.Unmarshal(raw, &oids)).To(Succeed())
	return oids
}
