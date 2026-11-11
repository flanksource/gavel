package query_test

import (
	"path/filepath"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/models"
)

// A detector program returns one result set per check, each its own shape:
// large activities, then the policies holding them.
const detectorProgram = "SELECT 'large' AS detector, guid, size FROM activities WHERE size >= 20 ORDER BY guid;\nGO\n" +
	"SELECT 'policies' AS detector, policy, count(*) AS n FROM activities GROUP BY policy ORDER BY policy;\n"

var _ = Describe("query result sets", func() {
	It("exposes every result set of a program in order, with rows the last", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "detect.sql"), detectorProgram)
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: detect
    kind: query
    connection: lab
    query: {queryFile: detect.sql}
    cel: >-
      size(resultSets) == 2 &&
      resultSets[0].map(r, r.guid) == ['a2', 'a3'] &&
      resultSets[1].map(r, r.policy) == ['P1', 'P2'] &&
      rows == resultSets[1] && rowCount == 2
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("exposes a plain query's rows as its one result set", func() {
		dir, _ := fixtureDir()
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: all
    kind: query
    connection: lab
    query: SELECT guid FROM activities ORDER BY guid
    cel: size(resultSets) == 1 && resultSets[0] == rows && rowCount == 3
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
	})

	It("maps findings from every result set, telling them apart by .resultSet", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "detect.sql"), detectorProgram)
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: detect
    kind: query
    connection: lab
    query: {queryFile: detect.sql}
    findings:
      when: "resultSet == 0 || row.n > 1"
      severity: warning
      rule: "{{.row.detector}}-{{.resultSet}}"
      message: "{{ if eq .row.detector \"large\" }}{{.row.guid}} holds {{.row.size}} KB{{ else }}{{.row.policy}} holds {{.row.n}} activities{{ end }}"
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusWarning), result.Error)
		Expect(findingsOf(result)).To(Equal([]models.Violation{
			{Message: models.StringPtr("a2 holds 20 KB"), Source: "query", Severity: models.SeverityWarning, Rule: &models.Rule{Method: "large-0"}},
			{Message: models.StringPtr("a3 holds 30 KB"), Source: "query", Severity: models.SeverityWarning, Rule: &models.Rule{Method: "large-0"}},
			{Message: models.StringPtr("P1 holds 2 activities"), Source: "query", Severity: models.SeverityWarning, Rule: &models.Rule{Method: "policies-1"}},
		}))
	})

	It("names the result set and row of a finding that fails to render", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "detect.sql"), detectorProgram)
		result := runYAML(dir, `
kind: Fixture
steps:
  - id: detect
    kind: query
    connection: lab
    query: {queryFile: detect.sql}
    findings:
      when: "resultSet == 1 ? row.policy : false"
      severity: warning
      message: "{{.row.policy}}"
`, fixtures.RunOptions{})
		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(ContainSubstring("findings result set 1 row 0:"))
	})
})
