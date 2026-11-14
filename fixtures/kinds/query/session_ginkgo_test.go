package query_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/dbtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// queryRecord is the step's query record decoded as plain JSON.
func queryRecord(result fixtures.FixtureResult) map[string]any {
	GinkgoHelper()
	raw, err := json.Marshal(result.Metadata[fixtures.QueryMetadataKey])
	Expect(err).NotTo(HaveOccurred())
	var record map[string]any
	Expect(json.Unmarshal(raw, &record)).To(Succeed())
	return record
}

// sessionStep is a fixture with one query step against the postgres at dsn.
func sessionStep(dsn, statement, extra string) string {
	return fmt.Sprintf(`
kind: Fixture
steps:
  - id: probe
    kind: query
    connection: {type: postgres, url: %q}
    query: %q
%s`, dsn, statement, extra)
}

var _ = Describe("query step server session", Ordered, func() {
	var dsn string
	BeforeAll(func() {
		dsn = dbtest.ForGinkgo(dbtest.Options{Name: "gavel_query_session"}).DSN()
	})

	It("records the backend pid the statement ran on and exposes it as the session output", func() {
		result := runYAML(GinkgoT().TempDir(), sessionStep(dsn, "SELECT pg_backend_pid() AS pid", `    cel: session.engine == 'postgres' && session.sessionId == row.pid`), fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		session, ok := queryRecord(result)["session"].(map[string]any)
		Expect(ok).To(BeTrue(), "session is %T", queryRecord(result)["session"])
		Expect(session).To(HaveKeyWithValue("engine", "postgres"))
		Expect(session["sessionId"]).To(BeNumerically(">", 0))
		Expect(queryRecord(result)).NotTo(HaveKey("diagnose"))
	})

	It("streams AutoDiagnose samples of the session once the statement outlives autoDiagnose", func() {
		result := runYAML(GinkgoT().TempDir(), sessionStep(dsn, "SELECT pg_backend_pid() AS pid, pg_sleep(1)", `    autoDiagnose: 200ms`), fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		record := queryRecord(result)
		session := record["session"].(map[string]any)
		Expect(record).NotTo(HaveKey("diagnoseError"))
		Expect(record["diagnose"]).To(ContainElement(And(
			HaveKeyWithValue("engine", "postgres"),
			HaveKeyWithValue("sessionId", session["sessionId"]),
			HaveKeyWithValue("state", "active"),
		)))
	})

	// 500ms ends before the default 10s and before the sampler's 2s interval,
	// so a sample proves -1 started sampling at once.
	It("samples from the start when autoDiagnose is -1", func() {
		result := runYAML(GinkgoT().TempDir(), sessionStep(dsn, "SELECT pg_sleep(0.5)", `    autoDiagnose: -1`), fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(queryRecord(result)["diagnose"]).To(HaveLen(1))
	})

	It("refuses a negative autoDiagnose other than -1 when the fixture is parsed", func() {
		path := filepath.Join(GinkgoT().TempDir(), "lint.fixture.yaml")
		writeFile(path, sessionStep(dsn, "SELECT 1", `    autoDiagnose: -5s`))
		_, err := fixtures.ParseFixtureFile(path)
		Expect(err).To(MatchError(ContainSubstring("/autoDiagnose: '-5s' does not match pattern")))
	})

	It("takes no samples when autoDiagnose is 0s", func() {
		result := runYAML(GinkgoT().TempDir(), sessionStep(dsn, "SELECT pg_sleep(0.5)", `    autoDiagnose: 0s`), fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(queryRecord(result)).To(And(HaveKey("session"), Not(HaveKey("diagnose"))))
	})
})

var _ = Describe("query step on sqlite", func() {
	It("records no server session: sqlite runs in process", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: count
    kind: query
    connection: lab
    query: "SELECT count(*) AS n FROM activities"
    autoDiagnose: 10ms
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(queryRecord(result)).To(And(Not(HaveKey("session")), Not(HaveKey("diagnose"))))
		Expect(result.Metadata["outputs"]).NotTo(HaveKey("session"))
	})
})
