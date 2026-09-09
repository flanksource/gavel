package testrunner

import (
	"path/filepath"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture test projection", func() {
	It("carries the inherited Markdown source location onto leaves", func() {
		file := filepath.Join(GinkgoT().TempDir(), "checks.fixture.md")
		node := &fixtures.FixtureNode{
			Name:   "checks.fixture.md",
			Type:   fixtures.FileNode,
			Origin: &fixtures.FixtureOrigin{File: file},
			Children: []*fixtures.FixtureNode{{
				Name: "health",
				Type: fixtures.TestNode,
				Origin: &fixtures.FixtureOrigin{
					File: filepath.Base(file),
					Line: 12,
				},
				Results: &fixtures.FixtureResult{Status: task.StatusPASS},
			}},
		}

		projected := fixtureNodeToTests(node)
		Expect(projected).To(HaveLen(1))
		Expect(projected[0].Children).To(HaveLen(1))
		Expect(projected[0].Children[0].Framework).To(Equal(parsers.Fixture))
		Expect(projected[0].Children[0].File).To(Equal(file))
		Expect(projected[0].Children[0].Line).To(Equal(12))
	})

	It("projects native CEL traces into fixture context", func() {
		trace := "cel: actual == expected\n     │         │"
		node := &fixtures.FixtureNode{
			Name: "trace",
			Type: fixtures.TestNode,
			Results: &fixtures.FixtureResult{
				Status:        task.StatusFAIL,
				CELExpression: "actual == expected",
				CELTrace:      trace,
			},
		}

		projected := fixtureNodeToTests(node)

		Expect(projected).To(HaveLen(1))
		Expect(projected[0].Context).To(Equal(parsers.FixtureContext{
			CELExpression: "actual == expected",
			CELTrace:      trace,
		}))
	})
})
