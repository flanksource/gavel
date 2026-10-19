package fixtures

import (
	"encoding/json"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a failed fixture's source", func() {
	const source = "kind: Fixture\nsteps:\n  - id: build\n    kind: exec\n  - id: render\n    kind: exec\n    exec: ./render\n"
	const renderStep = "  - id: render\n    kind: exec\n    exec: ./render"

	attached := func(status task.Status) (*FixtureNode, string) {
		path := writeDocument("render.fixture.yaml", source)
		node := &FixtureNode{Name: "render", Type: TestNode, Origin: &FixtureOrigin{File: path, Kind: "step", Line: 5, EndLine: 7}}
		(&Runner{}).attachResult(node, FixtureResult{Name: "render", Status: status, Error: "expected exit code 0, got 1"})
		return node, path
	}

	It("is recorded on the result and carried in its detail", func() {
		node, path := attached(task.StatusFAIL)
		Expect(node.Results.Metadata["source"]).To(Equal(parsers.SourceSnippet{File: path, Line: 5, EndLine: 7, Language: "yaml", Text: renderStep}))
		var detail struct {
			Source parsers.SourceSnippet `json:"source"`
		}
		Expect(json.Unmarshal(ResultDetail(node.Results), &detail)).To(Succeed())
		Expect(detail.Source.Text).To(Equal(renderStep))
	})

	It("is shown under a failed step, in the node and in the tree", func() {
		node, _ := attached(task.StatusFAIL)
		for _, rendered := range []string{node.Pretty().String(), node.Tree().Pretty().String()} {
			Expect(rendered).To(ContainSubstring("- id: render"))
			Expect(rendered).To(ContainSubstring("exec: ./render"))
			Expect(rendered).NotTo(ContainSubstring("id: build"))
		}
	})

	It("is not recorded for a passing step", func() {
		node, _ := attached(task.StatusPASS)
		Expect(node.Results.Metadata).NotTo(HaveKey("source"))
	})
})
