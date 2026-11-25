package database

import (
	"io/fs"

	"github.com/flanksource/commons-db/dbtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("line count cache migration", func() {
	It("invalidates derived counts without deleting repositories or worktrees", func() {
		script, err := fs.ReadFile(schemaFS, "schema/182_reset_line_count_caches.sql")
		Expect(err).NotTo(HaveOccurred())
		db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_line_count_cache_migration"}).Gorm()
		Expect(db.Exec(`
			CREATE TABLE git_range_stats (adds integer, file_list jsonb);
			CREATE TABLE git_repos (id integer, refs_fingerprint text);
			CREATE TABLE git_worktrees (path text, status_fingerprint text);
			INSERT INTO git_range_stats VALUES (100, '[{"path":"image.svg","adds":100}]');
			INSERT INTO git_repos VALUES (1, 'unchanged');
			INSERT INTO git_worktrees VALUES ('repo', 'unchanged');
		`).Error).To(Succeed())
		Expect(db.Exec(string(script)).Error).To(Succeed())
		var ranges int64
		Expect(db.Table("git_range_stats").Count(&ranges).Error).To(Succeed())
		Expect(ranges).To(BeZero())
		var repos []struct {
			ID              int
			RefsFingerprint string
		}
		Expect(db.Table("git_repos").Scan(&repos).Error).To(Succeed())
		Expect(repos).To(Equal([]struct {
			ID              int
			RefsFingerprint string
		}{{ID: 1}}))
		var worktrees []struct {
			Path              string
			StatusFingerprint string
		}
		Expect(db.Table("git_worktrees").Scan(&worktrees).Error).To(Succeed())
		Expect(worktrees).To(Equal([]struct {
			Path              string
			StatusFingerprint string
		}{{Path: "repo"}}))
		Expect(db.Exec(string(script)).Error).To(Succeed())
	})
})
