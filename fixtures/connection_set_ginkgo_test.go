package fixtures_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

// declaredConnections is the ConnectionSet a markdown document's front matter
// declares, read off its first test.
func declaredConnections(frontMatter string) (*fixtures.ConnectionSet, error) {
	node, err := fixtures.ParseMarkdownDocument("conns.md", "---\n"+frontMatter+"---\n# Conns\n\n```bash\ntrue\n```\n", ".")
	if err != nil {
		return nil, err
	}
	var found *fixtures.FixtureTest
	node.Walk(func(n *fixtures.FixtureNode) {
		if found == nil && n.Test != nil {
			found = n.Test
		}
	})
	Expect(found).NotTo(BeNil())
	Expect(found.Metadata).NotTo(HaveKey("connections"))
	Expect(found.Metadata).NotTo(HaveKey("connection"))
	return found.Connections, nil
}

var _ = Describe("ConnectionSet", func() {
	const inlineURL = "postgres://u:p@localhost:5432/scratch"

	DescribeTable("reads the front-matter list and map forms",
		func(frontMatter string, want *fixtures.ConnectionSet) {
			got, err := declaredConnections(frontMatter)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("list of defaults", "connections:\n  - connection://pg/main\n  - sqlserver-dev\n",
			&fixtures.ConnectionSet{Defaults: []string{"connection://pg/main", "sqlserver-dev"}}),
		Entry("map of string refs and inline objects", "connections:\n  warehouse: connection://pg/warehouse\n  scratch: {type: postgres, url: \""+inlineURL+"\"}\n",
			&fixtures.ConnectionSet{Named: map[string]any{
				"warehouse": "connection://pg/warehouse",
				"scratch":   map[string]any{"type": "postgres", "url": inlineURL},
			}}),
		Entry("legacy connection: key as a one-entry default list", "connection: lab\n",
			&fixtures.ConnectionSet{Defaults: []string{"lab"}}),
	)

	DescribeTable("rejects a malformed declaration, naming the shape it got",
		func(frontMatter, want string) {
			_, err := declaredConnections(frontMatter)
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("scalar", "connections: lab\n", "got string lab"),
		Entry("list entry that is not a string", "connections:\n  - {type: sqlite}\n", "connections[0] must be a connection name"),
		Entry("map entry that is a number", "connections:\n  lab: 3\n", "connections.lab must be"),
		Entry("legacy key with an inline object", "connection: {type: sqlite}\n", "declare an inline connection under connections:"),
		Entry("legacy key alongside connections:", "connection: lab\nconnections: [other]\n", "both connection: and connections:"),
	)

	It("round-trips through JSON in the form it was declared in", func() {
		doc := []byte(`{"warehouse":"connection://pg/warehouse","scratch":{"type":"postgres","url":"` + inlineURL + `"}}`)
		var set fixtures.ConnectionSet
		Expect(json.Unmarshal(doc, &set)).To(Succeed())
		encoded, err := json.Marshal(set)
		Expect(err).NotTo(HaveOccurred())
		Expect(encoded).To(MatchJSON(doc))
	})

	It("carries a YAML fixture document's connections onto each step", func() {
		node, err := fixtures.ParseFixtureDocuments("conns.fixture.yaml", []byte("kind: Fixture\nconnections: [lab]\nsteps:\n  - kind: exec\n    exec: 'true'\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(node.Children).To(HaveLen(1))
		Expect(node.Children[0].Test.Connections).To(Equal(&fixtures.ConnectionSet{Defaults: []string{"lab"}}))
	})
})
