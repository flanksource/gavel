package native_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/todos/native"
	. "github.com/onsi/ginkgo/v2"
	ginkgotypes "github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

type testServerIdentity struct {
	Port      int
	StartedAt time.Time
	Directory string
	Database  string
}

var _ = Describe("shared PostgreSQL test server", func() {
	It("reuses one server while keeping coordinator databases isolated", func(ctx SpecContext) {
		first := openCoordinatorFixture(ctx, "gavel_server_reuse_first")
		second := openCoordinatorFixture(ctx, "gavel_server_reuse_second")
		identities := make([]testServerIdentity, 2)
		for index, fixture := range []*coordinatorFixture{first, second} {
			identities[index] = readTestServerIdentity(ctx, fixture.db)
		}
		AddReportEntry("leased databases", identities, ReportEntryVisibilityNever)
		Expect(identities[1].Port).To(Equal(identities[0].Port))
		Expect(identities[1].StartedAt).To(Equal(identities[0].StartedAt))
		Expect(identities[1].Directory).To(Equal(identities[0].Directory))
		Expect(identities[1].Database).NotTo(Equal(identities[0].Database))

		issue := first.createIssue(ctx, "Visible only in the first database")
		stored, err := first.repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.ID).To(Equal(issue.ID))
		_, err = second.repository.GetIssue(ctx, issue.ID)
		Expect(err).To(MatchError(native.ErrNotFound))
	})

	It("releases leases and reuses the server across separate test processes", func(ctx SpecContext) {
		first, cleanup, err := dbtest.Open(dbtest.Options{Name: "gavel_reuse_released"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(cleanup()).To(Succeed()) })
		before := readTestServerIdentity(ctx, first.Gorm())
		observer := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_reuse_observer"})
		Expect(cleanup()).To(Succeed())
		var exists bool
		Expect(observer.SQL().QueryRowContext(ctx,
			`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, before.Database,
		).Scan(&exists)).To(Succeed())
		Expect(exists).To(BeFalse(), "released databases are dropped without stopping the server")

		var binariesBefore map[string]time.Time
		if os.Getenv(dbtest.EnvURL) == "" {
			binariesBefore = postgresBinaryModificationTimes(before.Directory)
		}
		databaseNames := []string{before.Database}
		for range 2 {
			for _, identity := range runCoordinatorReuseProcess(ctx) {
				Expect(identity).To(SatisfyAll(
					HaveField("Port", before.Port), HaveField("StartedAt", before.StartedAt),
					HaveField("Directory", before.Directory),
				))
				Expect(databaseNames).NotTo(ContainElement(identity.Database))
				databaseNames = append(databaseNames, identity.Database)
				Expect(observer.SQL().QueryRowContext(ctx,
					`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, identity.Database,
				).Scan(&exists)).To(Succeed())
				Expect(exists).To(BeFalse(), "child-process consumer pools close and leases are released")
			}
		}
		if os.Getenv(dbtest.EnvURL) == "" {
			Expect(postgresBinaryModificationTimes(before.Directory)).To(Equal(binariesBefore),
				"warm test processes reuse the extracted binaries")
		}
	})
})

func readTestServerIdentity(ctx context.Context, db *gorm.DB) testServerIdentity {
	GinkgoHelper()
	var identity testServerIdentity
	Expect(db.WithContext(ctx).Raw(`
		SELECT inet_server_port() AS port, pg_postmaster_start_time() AS started_at,
		       current_setting('data_directory') AS directory, current_database() AS database
	`).Scan(&identity).Error).To(Succeed())
	return identity
}

func runCoordinatorReuseProcess(ctx context.Context) []testServerIdentity {
	GinkgoHelper()
	executable, err := os.Executable()
	Expect(err).NotTo(HaveOccurred())
	reportPath := filepath.Join(GinkgoT().TempDir(), "reuse.json")
	command := exec.CommandContext(ctx, executable, "-test.run=^TestNativePlanCreation$",
		"-ginkgo.focus=reuses one server while keeping coordinator databases isolated$",
		"-ginkgo.json-report="+reportPath, "-ginkgo.no-color", "-ginkgo.fail-on-empty")
	output, err := command.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "%s", output)
	data, err := os.ReadFile(reportPath)
	Expect(err).NotTo(HaveOccurred())
	var reports []ginkgotypes.Report
	Expect(json.Unmarshal(data, &reports)).To(Succeed())
	var identities []testServerIdentity
	for _, report := range reports {
		Expect(report.SuiteSucceeded).To(BeTrue())
		for _, spec := range report.SpecReports {
			for _, entry := range spec.ReportEntries {
				if entry.Name == "leased databases" {
					Expect(json.Unmarshal([]byte(entry.Value.AsJSON), &identities)).To(Succeed())
				}
			}
		}
	}
	Expect(identities).To(HaveLen(2), "the child process must execute the real coordinator reuse regression")
	return identities
}

func postgresBinaryModificationTimes(dataDirectory string) map[string]time.Time {
	GinkgoHelper()
	paths, err := filepath.Glob(filepath.Join(filepath.Dir(dataDirectory), "bin", "*", "bin", "postgres"))
	Expect(err).NotTo(HaveOccurred())
	Expect(paths).NotTo(BeEmpty())
	modified := make(map[string]time.Time, len(paths))
	for _, path := range paths {
		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		modified[path] = info.ModTime()
	}
	return modified
}
