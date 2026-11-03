package services_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func databaseConnection(dsn string) *sql.DB {
	GinkgoHelper()
	connection, err := sql.Open("postgres", dsn)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(connection.Close()).To(Succeed()) })
	return connection
}

func databaseText(db *sql.DB, query string) string {
	GinkgoHelper()
	var result string
	Expect(db.QueryRow(query).Scan(&result)).To(Succeed())
	return result
}

var _ = Describe("services kind: database integration", Label("integration"), func() {
	BeforeEach(func() {
		GinkgoT().Setenv("COMMONS_DB_URL", "")
		GinkgoT().Setenv("COMMONS_DB_CREATE", "true")
	})

	It("exports automatic connections that PostgreSQL and dbtest children actually use", func() {
		step := newStep("stack", `{"services":{"db":{"dbtest":{}}}}`)
		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		db := entryOutputs(outcome, "db")
		for _, key := range []string{"DATABASE_URL", "GAVEL_DB_DSN", "COMMONS_DB_URL", "DB_URL"} {
			Expect(envValue(step, key)).To(Equal(db["url"]))
		}
		Expect(envValue(step, "COMMONS_DB_CREATE")).To(Equal("false"))
		Expect(os.Getenv("COMMONS_DB_URL")).To(BeEmpty())
		for _, mode := range []string{"env", "dbtest"} {
			cmd := exec.Command(dbClientBin, mode)
			cmd.Env = step.Environ()
			data, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(data))
			var actual map[string]string
			Expect(json.Unmarshal(data, &actual)).To(Succeed(), string(data))
			Expect(actual["database"]).To(Equal(db["database"]))
		}
	})

	It("leases isolated databases on the same server and releases only its own database", func() {
		step := newStep("stack", `{"services":{"one":{"dbtest":{"default":true}},"two":{"dbtest":{}}}}`)
		outcome, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		one, two := entryOutputs(outcome, "one"), entryOutputs(outcome, "two")
		Expect(one["database"]).NotTo(Equal(two["database"]))
		first, second := databaseConnection(one["url"].(string)), databaseConnection(two["url"].(string))
		Expect(databaseText(first, "SELECT pg_postmaster_start_time()::text")).To(Equal(databaseText(second, "SELECT pg_postmaster_start_time()::text")))
		_, err = first.Exec("CREATE TABLE only_in_first(value text)")
		Expect(err).NotTo(HaveOccurred())
		Expect(databaseText(second, "SELECT to_regclass('only_in_first') IS NULL")).To(Equal("true"))
		Expect(envValue(step, "DATABASE_URL")).To(Equal(one["url"]))
		Expect(first.Close()).To(Succeed())
		Expect(second.Close()).To(Succeed())
		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		Expect(databaseConnection(one["url"].(string)).Ping()).NotTo(Succeed())
		Expect(outcome.Cleanup(context.Background())).To(Succeed())
		other, err := execute(newStep("other", `{"services":{"db":{"dbtest":{}}}}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(entryOutputs(other, "db")["port"]).To(Equal(one["port"]))
	})

	It("reuses prepared templates, runs instance SQL, and invalidates changed migrations", func() {
		step := newStep("stack", `{"services":{"db":{"dbtest":{"migrations":{"dir":"schema","name":"fixture-template"}}}}}`)
		dir := filepath.Join(step.WorkDir, "schema")
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "schema.hcl"), []byte(`schema "public" {}
table "instance_events" {
  schema = schema.public
  column "value" { type = text }
}`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "seed.sql"), []byte("CREATE TABLE template_token AS SELECT md5(random()::text) AS value;"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "instance.sql"), []byte("-- runs: always\nINSERT INTO instance_events VALUES (current_database());"), 0o600)).To(Succeed())
		first, err := execute(step)
		Expect(err).NotTo(HaveOccurred())
		secondStep := newStep("stack", string(step.Spec.Config))
		secondStep.WorkDir = step.WorkDir
		second, err := execute(secondStep)
		Expect(err).NotTo(HaveOccurred())
		one := databaseConnection(entryOutputs(first, "db")["url"].(string))
		two := databaseConnection(entryOutputs(second, "db")["url"].(string))
		token := databaseText(one, "SELECT value FROM template_token")
		Expect(databaseText(two, "SELECT value FROM template_token")).To(Equal(token))
		Expect(databaseText(two, "SELECT count(*) FROM instance_events WHERE value = current_database()")).To(Equal("1"))
		Expect(os.WriteFile(filepath.Join(dir, "changed.sql"), []byte("CREATE TABLE added(value text);"), 0o600)).To(Succeed())
		thirdStep := newStep("stack", string(step.Spec.Config))
		thirdStep.WorkDir = step.WorkDir
		third, err := execute(thirdStep)
		Expect(err).NotTo(HaveOccurred())
		three := databaseConnection(entryOutputs(third, "db")["url"].(string))
		Expect(databaseText(three, "SELECT value FROM template_token")).NotTo(Equal(token))
		Expect(databaseText(three, "SELECT to_regclass('added') IS NOT NULL")).To(Equal("true"))
	})

	It("releases an earlier database when a later entry fails", func() {
		step := newStep("stack", `{"services":{"db":{"dbtest":{}},"broken":{"exec":"exit 1","ready":{"tcp":"127.0.0.1:1"},"wait":{"timeout":"1s"}}}}`)
		outcome, err := execute(step)
		Expect(err).To(HaveOccurred())
		dsn := entryOutputs(outcome, "db")["url"].(string)
		Expect(databaseConnection(dsn).Ping()).NotTo(Succeed())
	})

	It("uses an external disposable database directly and refuses provisioning in direct mode", func() {
		owner, err := execute(newStep("owner", `{"services":{"db":{"dbtest":{}}}}`))
		Expect(err).NotTo(HaveOccurred())
		dsn := entryOutputs(owner, "db")["url"].(string)
		GinkgoT().Setenv("COMMONS_DB_URL", dsn)
		GinkgoT().Setenv("COMMONS_DB_CREATE", "false")
		borrowed, err := execute(newStep("borrowed", `{"services":{"db":{"dbtest":{}}}}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(entryOutputs(borrowed, "db")["url"]).To(Equal(dsn))
		Expect(borrowed.Cleanup(context.Background())).To(Succeed())
		Expect(databaseConnection(dsn).Ping()).To(Succeed())
		step := newStep("direct", fmt.Sprintf(`{"services":{"db":{"dbtest":{"migrations":{"dir":%q}}}}}`, GinkgoT().TempDir()))
		_, err = execute(step)
		Expect(err).To(MatchError(ContainSubstring("COMMONS_DB_CREATE=false")))
	})

	It("connects later services and commands to a default declared in another services step", func() {
		data := fmt.Sprintf(`kind: Fixture
steps:
  - id: other
    kind: services
    services: {aux: {dbtest: {}}, audit: {dbtest: {}}}
  - id: primary
    kind: services
    services: {app: {dbtest: {default: true}}}
  - id: query
    kind: exec
    exec: %q
    args: [dbtest]
    cel: json.database == steps.primary.app.database
`, dbClientBin)
		tree, err := fixtures.ParseFixtureDocuments("database.fixture.yaml", []byte(data))
		Expect(err).NotTo(HaveOccurred())
		results, _, err := fixtures.RunNodes(context.Background(), tree.Children, fixtures.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(3))
		for _, result := range results {
			Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		}
	})

	It("passes automatic connections into a test runner child", func() {
		data := `kind: Fixture
steps:
  - id: resources
    kind: services
    services:
      db: {dbtest: {}}
      workspace:
        git:
          files:
            go.mod: |
              module example.com/fixture-env
              go 1.26.1
            connection_test.go: |
              package fixtureenv
              import ("net"; "os"; "testing"; "time")
              func TestConnection(t *testing.T) {
                if os.Getenv("PGDATABASE") != os.Getenv("EXPECTED_DATABASE") { t.Fatal("wrong database environment") }
                if os.Getenv("COMMONS_DB_CREATE") != "false" { t.Fatal("child would create another database") }
                connection, err := net.DialTimeout("tcp", net.JoinHostPort(os.Getenv("PGHOST"), os.Getenv("PGPORT")), time.Second)
                if err != nil { t.Fatal(err) }
                if err := connection.Close(); err != nil { t.Fatal(err) }
              }
  - id: tests
    kind: test
    paths: [.]
    framework: [go]
    work-dir: '{{.steps.resources.workspace.dir}}'
    runtime:
      env:
        GOWORK: 'off'
        EXPECTED_DATABASE: '{{.steps.resources.db.database}}'
`
		tree, err := fixtures.ParseFixtureDocuments("database.fixture.yaml", []byte(data))
		Expect(err).NotTo(HaveOccurred())
		results, _, err := fixtures.RunNodes(context.Background(), tree.Children, fixtures.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		Expect(results[1].Status).To(Equal(task.StatusPASS), results[1].Error)
	})
})
