package query_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/clicky/task"
	dbcontext "github.com/flanksource/commons-db/context"
	cdbmodels "github.com/flanksource/commons-db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/query"
	"github.com/flanksource/gavel/models"
)

func TestQueryKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Kind Suite")
}

// activities are the rows every spec's database starts with.
var activities = [][]any{{"a1", 10, "P1"}, {"a2", 20, "P1"}, {"a3", 30, "P2"}}

// seedDatabase creates an activities table at path holding rows.
func seedDatabase(path string, rows [][]any) {
	GinkgoHelper()
	db, err := sql.Open("sqlite", path)
	Expect(err).NotTo(HaveOccurred())
	defer db.Close()
	_, err = db.Exec("CREATE TABLE activities(guid TEXT, size INTEGER, policy TEXT)")
	Expect(err).NotTo(HaveOccurred())
	for _, row := range rows {
		_, err = db.Exec("INSERT INTO activities VALUES (?, ?, ?)", row...)
		Expect(err).NotTo(HaveOccurred())
	}
}

// fixtureDir is a fixture directory with data.db seeded and saved as the
// connection `lab` in its .connections.
func fixtureDir() (dir, dbPath string) {
	GinkgoHelper()
	GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
	dir = GinkgoT().TempDir()
	dbPath = filepath.Join(dir, "data.db")
	seedDatabase(dbPath, activities)
	writeFile(filepath.Join(dir, ".connections", "lab.yaml"), "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
		"metadata:\n  name: lab\nspec:\n  type: sqlite\n  url:\n    value: "+dbPath+"\n")
	return dir, dbPath
}

// viewOf is the clicky document a query step published, decoded as plain JSON.
func viewOf(result fixtures.FixtureResult) map[string]any {
	GinkgoHelper()
	raw, err := json.Marshal(result.Metadata[fixtures.ViewMetadataKey])
	Expect(err).NotTo(HaveOccurred())
	var view map[string]any
	Expect(json.Unmarshal(raw, &view)).To(Succeed())
	return view
}

func writeFile(path, body string) {
	GinkgoHelper()
	Expect(os.MkdirAll(filepath.Dir(path), 0o700)).To(Succeed())
	Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
}

func runYAML(dir, body string, opts fixtures.RunOptions) fixtures.FixtureResult {
	GinkgoHelper()
	path := filepath.Join(dir, "lint.fixture.yaml")
	writeFile(path, body)
	file, err := fixtures.ParseFixtureFile(path)
	Expect(err).NotTo(HaveOccurred())
	opts.WorkDir = dir
	results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{file}, opts)
	Expect(err).NotTo(HaveOccurred())
	Expect(results).To(HaveLen(1))
	return results[0]
}

func findingsOf(result fixtures.FixtureResult) []models.Violation {
	GinkgoHelper()
	outputs, ok := result.Metadata["outputs"].(map[string]any)
	Expect(ok).To(BeTrue(), "outputs are %T", result.Metadata["outputs"])
	raw, err := json.Marshal(outputs["findings"])
	Expect(err).NotTo(HaveOccurred())
	var violations []models.Violation
	Expect(json.Unmarshal(raw, &violations)).To(Succeed())
	for i := range violations {
		Expect(violations[i].Fingerprint).NotTo(BeEmpty())
		violations[i].Fingerprint = ""
	}
	return violations
}

func childStatuses(result fixtures.FixtureResult) []task.Status {
	var statuses []task.Status
	for _, child := range result.Children {
		statuses = append(statuses, child.Results.Status)
	}
	return statuses
}

var _ = Describe("query kind", func() {
	It("runs query text against a named connection, binding params as arguments", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: big
    kind: query
    connection: lab
    query: SELECT guid, size FROM activities WHERE size >= {{.params.min}} ORDER BY guid
    params: {min: 20}
    cel: "rowCount == 2 && row.guid == 'a2' && columns == ['guid', 'size'] && connection == {'name': 'lab', 'type': 'sqlite'}"
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("runs an inline profile body against an inline connection", func() {
		dir, dbPath := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: count
    kind: query
    connection: {type: sqlite, url: `+dbPath+`}
    query:
      query: SELECT count(*) AS n FROM activities WHERE policy = $(.params.policy)
      params: [{name: policy, type: string}]
      columns: [{name: n}]
    params: {policy: P1}
    cel: row.n == 2 && columns == ['n'] && connection.name == 'inline'
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("runs a stored profile from the fixture's .profiles with the reference's params", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, ".profiles", "large.yaml"), `profile: large-activities
query: SELECT guid FROM activities WHERE size >= {{.params.min}}
params: [{name: min, type: number}]
`)
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: large
    kind: query
    connection: connection://lab
    query: {profile: large-activities, params: {min: 25}}
    cel: rowCount == 1 && row.guid == 'a3'
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("asks the host's connection resolver before the fixture's .connections", func() {
		dir, _ := fixtureDir()
		hostDB := filepath.Join(GinkgoT().TempDir(), "host.db")
		seedDatabase(hostDB, activities[:1])
		host := func(_ dbcontext.Context, reference string) (*cdbmodels.Connection, error) {
			if reference != "lab" {
				return nil, nil
			}
			return &cdbmodels.Connection{Name: "lab", Type: cdbmodels.ConnectionTypeSQLite, URL: hostDB}, nil
		}
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: all
    kind: query
    connection: lab
    query: SELECT guid FROM activities
    cel: rowCount == 1
`, fixtures.RunOptions{Connections: host})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("fails the step on error findings from a CEL expression, each with every location", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: oversized
    kind: query
    connection: lab
    query: SELECT guid, size, policy FROM activities ORDER BY guid
    findings: >-
      outputs.rows.filter(r, r.size >= 20).map(r, {"severity": "error", "message": r.guid + " is " + string(r.size),
        "locations": ["queries/q.sql:" + string(r.size), {"entity": {"kind": "Policy", "id": r.policy}}]})
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusFAIL))
		Expect(childStatuses(result)).To(Equal([]task.Status{task.StatusFAIL, task.StatusFAIL}))
		Expect(findingsOf(result)).To(Equal([]models.Violation{
			{File: "queries/q.sql", Line: 20, Message: models.StringPtr("a2 is 20"), Source: "query", Severity: models.SeverityError,
				Rule: &models.Rule{Method: "oversized"}, Locations: []models.ViolationLocation{
					{File: "queries/q.sql", Line: 20}, {Entity: &models.EntityTarget{Kind: "Policy", ID: "P1"}}}},
			{File: "queries/q.sql", Line: 30, Message: models.StringPtr("a3 is 30"), Source: "query", Severity: models.SeverityError,
				Rule: &models.Rule{Method: "oversized"}, Locations: []models.ViolationLocation{
					{File: "queries/q.sql", Line: 30}, {Entity: &models.EntityTarget{Kind: "Policy", ID: "P2"}}}},
		}))
		// The tree row names the rule and message; the locations are the finding's detail.
		var names []string
		for _, child := range result.Children {
			names = append(names, child.Name)
			Expect(child.Results.Metadata[fixtures.ViolationMetadataKey]).To(HaveField("Locations", HaveLen(2)))
		}
		Expect(names).To(Equal([]string{"oversized: a2 is 20", "oversized: a3 is 30"}))
	})

	It("warns the step on findings a row mapper maps under when, with a CEL severity and templated locations", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: review
    kind: query
    connection: lab
    query: SELECT guid, size FROM activities ORDER BY guid
    findings:
      when: row.size >= 20
      severity: "row.size >= 30 ? 'warning' : 'info'"
      rule: large-{{.row.guid}}
      message: "{{.row.guid}} holds {{.row.size}} KB"
      locations:
        - "main:rules/{{.row.guid}}.xml:{{.row.size}}:2"
        - url: {url: "https://app.example/activity/{{.row.guid}}", css: "[data-id='{{.row.guid}}']"}
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusWarning), result.Error)
		Expect(childStatuses(result)).To(Equal([]task.Status{task.StatusPASS, task.StatusWarning}))
		Expect(findingsOf(result)).To(Equal([]models.Violation{
			{File: "rules/a2.xml", Line: 20, Column: 2, Message: models.StringPtr("a2 holds 20 KB"), Source: "query", Severity: models.SeverityInfo,
				Rule: &models.Rule{Method: "large-a2"}, Locations: []models.ViolationLocation{
					{Ref: "main", File: "rules/a2.xml", Line: 20, Column: 2},
					{URL: &models.URLTarget{URL: "https://app.example/activity/a2", CSS: "[data-id='a2']"}}}},
			{File: "rules/a3.xml", Line: 30, Column: 2, Message: models.StringPtr("a3 holds 30 KB"), Source: "query", Severity: models.SeverityWarning,
				Rule: &models.Rule{Method: "large-a3"}, Locations: []models.ViolationLocation{
					{Ref: "main", File: "rules/a3.xml", Line: 30, Column: 2},
					{URL: &models.URLTarget{URL: "https://app.example/activity/a3", CSS: "[data-id='a3']"}}}},
		}))
	})

	DescribeTable("errors loudly on a malformed findings element, naming it",
		func(findings, problem string) {
			dir, _ := fixtureDir()
			result := runYAML(dir, `
kind: Fixture
steps:
  - id: bad
    kind: query
    connection: lab
    query: SELECT guid FROM activities ORDER BY guid
    findings: `+findings+`
`, fixtures.RunOptions{})
			Expect(result.Status).To(Equal(task.StatusERR))
			Expect(result.Error).To(ContainSubstring(problem))
		},
		Entry("a list of strings", `"outputs.rows.map(r, r.guid)"`, "findings[0] is string, want a map"),
		Entry("not a list", `"size(rows)"`, "want a list of maps"),
		Entry("an unknown severity", `"outputs.rows.map(r, {'severity': 'fatal', 'message': r.guid})"`, `findings[0]: unknown severity "fatal"`),
		Entry("two URL selectors", `"outputs.rows.map(r, {'severity': 'error', 'message': r.guid, 'locations': [{'url': {'url': 'https://x.example', 'css': '.a', 'xpath': '//a'}}]})"`,
			"findings[0]: locations[0]: url target https://x.example sets css and xpath"),
		Entry("an unparsable shortcut", `"outputs.rows.map(r, {'severity': 'error', 'message': r.guid, 'locations': ['entity://Policy']})"`,
			"entity://<kind>/<id>"),
	)

	It("runs DML with bound params as a SQL program and reports the rows it changed", func() {
		dir, dbPath := fixtureDir()
		const purgedPolicy = "P1"
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: purge
    kind: query
    connection: lab
    query: DELETE FROM activities WHERE policy = $(.params.policy)
    params: {policy: `+purgedPolicy+`}
    cel: rowsAffected == 2 && size(batches) == 1
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		db, err := sql.Open("sqlite", dbPath)
		Expect(err).NotTo(HaveOccurred())
		defer db.Close()
		var remaining int
		Expect(db.QueryRow("SELECT count(*) FROM activities WHERE policy = ?", purgedPolicy).Scan(&remaining)).To(Succeed())
		Expect(remaining).To(BeZero())
	})

	It("publishes the response as commons-db's clicky table over the profile's columns", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: inspect
    kind: query
    connection: lab
    query: {query: "SELECT guid, size FROM activities ORDER BY guid", columns: [{name: guid}, {name: size, type: number}]}
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		view := viewOf(result)
		Expect(view).To(HaveKeyWithValue("version", BeNumerically("==", 1)))
		node := view["node"].(map[string]any)
		Expect(node).To(HaveKeyWithValue("kind", "table"))
		Expect(node["rows"]).To(HaveLen(len(activities)))
		Expect(fmt.Sprint(node["columns"])).To(And(ContainSubstring("guid"), ContainSubstring("size")))
	})

	It("records the query as authored, the SQL it issued and the params it bound", func() {
		dir, _ := fixtureDir()
		const template = "SELECT guid FROM activities WHERE size >= $(.params.min) ORDER BY guid"
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: big
    kind: query
    connection: lab
    query: "`+template+`"
    params: {min: 20}
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		raw, err := json.Marshal(result.Metadata[fixtures.QueryMetadataKey])
		Expect(err).NotTo(HaveOccurred())
		var record map[string]any
		Expect(json.Unmarshal(raw, &record)).To(Succeed())
		Expect(record).To(HaveKeyWithValue("template", template))
		Expect(record).To(HaveKeyWithValue("params", map[string]any{"min": float64(20)}))
		Expect(record["sql"]).To(And(ContainSubstring("FROM activities"), Not(ContainSubstring("$(.params.min)"))))
		Expect(record["arguments"]).To(ContainElement(BeNumerically("==", 20)))
	})

	It("runs an inline profile's GO-separated queryFile relative to the fixture", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "stage.sql"), "CREATE TABLE stage(guid TEXT);\nGO\n"+
			"INSERT INTO stage SELECT guid FROM activities WHERE size >= 20;\nGO\nSELECT count(*) AS staged FROM stage;\n")
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: stage
    kind: query
    connection: lab
    query: {queryFile: stage.sql}
    cel: size(batches) == 3 && row.staged == 2
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Command).To(Equal("stage.sql"))
		rendered, err := json.Marshal(viewOf(result))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(rendered)).To(And(ContainSubstring("batch 1 (lines 1-1)"), ContainSubstring("batch 3 (lines 5-5)"), ContainSubstring("staged")))
	})

	// sqlite is the only driver here and has no databases to switch between, so
	// its refusal proves database: reaches the provider; the SQL Server switch
	// is exercised by oipa-cli's purge fixture integration.
	It("hands database: to the SQL provider, which refuses to switch a sqlite connection", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
env: {TARGET: uat}
steps:
  - id: other
    kind: query
    connection: lab
    database: "{{.env.TARGET}}"
    query: SELECT 1 AS one
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(ContainSubstring("failed to select sql database: sqlite snapshots do not support database switching"))
	})

	It("fails loudly when no layer defines the connection", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: missing
    kind: query
    connection: connection://nowhere
    query: SELECT 1 AS one
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(ContainSubstring(`connection "connection://nowhere": connection not found`))
	})

	It("streams each batch's end to the running step's output", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "stage.sql"), "CREATE TABLE stage(guid TEXT);\nGO\nINSERT INTO stage SELECT guid FROM activities WHERE size >= 20;\n")
		path := filepath.Join(dir, "stream.fixture.yaml")
		writeFile(path, "kind: Fixture\nsteps:\n  - id: stage\n    kind: query\n    connection: lab\n    query: {queryFile: stage.sql}\n")
		file, err := fixtures.ParseFixtureFile(path)
		Expect(err).NotTo(HaveOccurred())
		var running []string
		sink := func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
			var walk func(*fixtures.ExecutionNode)
			walk = func(node *fixtures.ExecutionNode) {
				if node.State == fixtures.ExecutionRunning && node.Output != "" {
					running = append(running, node.Output)
				}
				for _, child := range node.Children {
					walk(child)
				}
			}
			walk(snapshot.Root)
			return nil
		}
		results, _, err := fixtures.RunNodes(fixtures.WithProgressSink(context.Background(), sink), []*fixtures.FixtureNode{file}, fixtures.RunOptions{WorkDir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		// The first line publishes at once; the rest of the burst lands with the result.
		Expect(running).To(Equal([]string{"batch 1 done (0 rows affected)\n"}))
	})

	It("waits by re-running the query until its outputs satisfy until", func() {
		dir, dbPath := fixtureDir()
		go func() {
			defer GinkgoRecover()
			time.Sleep(200 * time.Millisecond)
			db, err := sql.Open("sqlite", dbPath)
			Expect(err).NotTo(HaveOccurred())
			defer db.Close()
			_, err = db.Exec("INSERT INTO activities VALUES ('a4', 40, 'P2')")
			Expect(err).NotTo(HaveOccurred())
		}()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: arrived
    kind: query
    connection: lab
    query: SELECT guid FROM activities
    wait: {until: "rowCount == 4", interval: 50ms, timeout: 10s}
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("runs a markdown Query row with its Connection, CEL Validation and Findings columns", func() {
		dir, _ := fixtureDir()
		node, err := fixtures.ParseMarkdownDocument("lint.md", "# Lint\n\n"+
			"| Name | Query | Connection | CEL Validation | Findings |\n"+
			"|------|-------|------------|----------------|----------|\n"+
			"| big rows | SELECT guid, size FROM activities WHERE size >= 20 ORDER BY guid | lab | rowCount == 2 | outputs.rows.map(r, {'severity': 'warning', 'message': r.guid}) |\n", dir)
		Expect(err).NotTo(HaveOccurred())
		results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{node}, fixtures.RunOptions{WorkDir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Status).To(Equal(task.StatusWarning), results[0].Error)
		Expect(results[0].CELResult).To(BeTrue())
		Expect(childStatuses(results[0])).To(Equal([]task.Status{task.StatusWarning, task.StatusWarning}))
	})
})
