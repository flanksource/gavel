package query_test

import (
	"encoding/json"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("query batch response", func() {
	It("puts each batch's result under its own collapsed first-line heading", func() {
		dir, _ := fixtureDir()
		writeFile(filepath.Join(dir, "batches.sql"), "SELECT 'first' AS value;\nGO\nSELECT 'second' AS value;\n")
		result := runYAML(dir, "kind: Fixture\nsteps:\n  - id: batches\n    kind: query\n    connection: lab\n    query: {queryFile: batches.sql}\n", fixtures.RunOptions{})
		var sections []map[string]any
		var walk func(map[string]any)
		walk = func(node map[string]any) {
			if node["kind"] == "collapsed" {
				sections = append(sections, node)
				return
			}
			for _, key := range []string{"children", "items"} {
				children, _ := node[key].([]any)
				for _, child := range children {
					walk(child.(map[string]any))
				}
			}
		}
		walk(viewOf(result)["node"].(map[string]any))
		Expect(sections).To(HaveLen(2))
		for index, section := range sections {
			Expect(section["label"]).To(HaveKeyWithValue("text", ContainSubstring([]string{"batch 1 (lines 1-1)", "batch 2 (lines 3-3)"}[index])))
			content, err := json.Marshal(section["content"])
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(And(ContainSubstring([]string{"first", "second"}[index]), Not(ContainSubstring([]string{"second", "first"}[index]))))
		}
	})
})
