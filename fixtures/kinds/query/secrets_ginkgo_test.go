package query_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons/properties"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

var _ = Describe("A fixture connection's secret", func() {
	// Expiring the process-wide value cache at once stands in for a fixture
	// that outlives its one minute.
	const cacheTimeout = "envvar.1password.cache.timeout"
	const document = `kind: Fixture
connections: [vaulted]
steps:
  - id: named
    kind: query
    connection: vaulted
    query: SELECT guid FROM activities
    cel: rowCount == 3
  - id: referenced
    kind: query
    connection: connection://vaulted
    query: SELECT guid FROM activities WHERE size > 10
    cel: rowCount == 2
  - id: defaulted
    kind: query
    query: SELECT count(*) AS n FROM activities
    cel: row.n == 3
`
	var dir, path, calls string

	BeforeEach(func() {
		properties.Set(cacheTimeout, "1ns")
		dbcontext.New().ClearCache()
		DeferCleanup(func() {
			properties.Set(cacheTimeout, "")
			dbcontext.New().ClearCache()
		})
		var dbPath string
		dir, dbPath = fixtureDir()
		bin := GinkgoT().TempDir()
		calls = filepath.Join(bin, "calls")
		op := filepath.Join(bin, "op")
		writeFile(op, "#!/bin/sh\necho \"$@\" >> '"+calls+"'\nprintf %s '"+dbPath+"'\n")
		Expect(os.Chmod(op, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("OP_SERVICE_ACCOUNT_TOKEN", "")
		writeFile(filepath.Join(dir, ".connections", "vaulted.yaml"), "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
			"metadata:\n  name: vaulted\nspec:\n  type: sqlite\n  url:\n    valueFrom:\n      onePassword: op://example/lab/url\n")
		path = filepath.Join(dir, "vaulted.fixture.yaml")
		writeFile(path, document)
	})

	reads := func() []string {
		GinkgoHelper()
		raw, err := os.ReadFile(calls)
		Expect(err).NotTo(HaveOccurred())
		return strings.Split(strings.TrimSpace(string(raw)), "\n")
	}

	It("is read once across the steps of RunNodes", func() {
		file, err := fixtures.ParseFixtureFile(path)
		Expect(err).NotTo(HaveOccurred())
		results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{file}, fixtures.RunOptions{WorkDir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(3))
		for _, result := range results {
			Expect(result.Error).To(BeEmpty())
		}
		Expect(reads()).To(Equal([]string{"read --no-newline -- op://example/lab/url"}))
	})

	It("is read once across the steps of a Runner", func() {
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).NotTo(HaveOccurred())
		Expect(reads()).To(Equal([]string{"read --no-newline -- op://example/lab/url"}))
	})
})
