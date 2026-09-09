package fixtures_test

import (
	"context"
	"sync"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/record"

	// Registers the exec/query fixture types and the ai/runner step hooks.
	_ "github.com/flanksource/gavel/fixtures/types"
)

const passingDocument = `---
codeBlocks: [bash]
---

# Run nodes

## Shell checks

### command: echoes

` + "```bash\necho hello\n```" + `

- cel: exitCode == 0

### command: also echoes

` + "```bash\necho again\n```" + `

- cel: exitCode == 0
`

const failingDocument = `---
codeBlocks: [bash]
---

# Run nodes

### command: exits non-zero

` + "```bash\nexit 3\n```" + `

- cel: exitCode == 0
`

func parseDocument(markdown string) []*fixtures.FixtureNode {
	GinkgoHelper()
	tree, err := fixtures.ParseMarkdownDocument("verification", markdown, GinkgoT().TempDir())
	Expect(err).ToNot(HaveOccurred())
	return tree.Children
}

// firstTest is the first executable node in tree order, however deep the
// headings above it nest.
func firstTest(nodes []*fixtures.FixtureNode) fixtures.FixtureTest {
	GinkgoHelper()
	for _, node := range nodes {
		if node.Test != nil {
			return *node.Test
		}
		if len(node.Children) > 0 {
			return firstTest(node.Children)
		}
	}
	Fail("document declares no test node")
	return fixtures.FixtureTest{}
}

func namesOf(results []fixtures.FixtureResult) []string {
	names := make([]string, 0, len(results))
	for _, r := range results {
		names = append(names, r.Name)
	}
	return names
}

var _ = Describe("fixtures.RunNodes", func() {
	It("runs every test node in tree order and reports a passing snapshot", func() {
		results, snapshot, err := fixtures.RunNodes(context.Background(), parseDocument(passingDocument),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(namesOf(results)).To(Equal([]string{"echoes", "also echoes"}))
		for _, result := range results {
			Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		}
		Expect(snapshot).ToNot(BeNil())
		Expect(snapshot.State).To(Equal(fixtures.ExecutionPassed))
		Expect(snapshot.Summary.Total).To(Equal(2))
		Expect(snapshot.Summary.Passed).To(Equal(2))
	})

	It("reports a failing node as a failed result and a failed snapshot", func() {
		results, snapshot, err := fixtures.RunNodes(context.Background(), parseDocument(failingDocument),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].IsOK()).To(BeFalse())
		Expect(snapshot.State).To(Equal(fixtures.ExecutionFailed))
		Expect(snapshot.Summary.Failed).To(Equal(1))
	})

	It("streams execution snapshots to the sink registered on the context", func() {
		var mu sync.Mutex
		var states []fixtures.ExecutionState
		ctx := fixtures.WithProgressSink(context.Background(),
			func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
				mu.Lock()
				defer mu.Unlock()
				states = append(states, snapshot.State)
				return nil
			})

		_, _, err := fixtures.RunNodes(ctx, parseDocument(passingDocument),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		mu.Lock()
		defer mu.Unlock()
		Expect(states).ToNot(BeEmpty())
		Expect(states[0]).To(Equal(fixtures.ExecutionQueued))
		Expect(states).To(ContainElement(fixtures.ExecutionRunning))
		Expect(states[len(states)-1]).To(Equal(fixtures.ExecutionPassed))
	})

	It("returns an errored result for a node no fixture type can dispatch", func() {
		node := &fixtures.FixtureNode{
			Name: "undispatchable", Type: fixtures.TestNode,
			Test: &fixtures.FixtureTest{Name: "undispatchable"},
		}

		results, snapshot, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{node},
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Status).To(Equal(task.StatusERR))
		Expect(results[0].Error).To(ContainSubstring("unable to determine fixture type"))
		Expect(snapshot.State).To(Equal(fixtures.ExecutionErrored))
	})

	It("honours a fixture's skip declaration without running it", func() {
		node := &fixtures.FixtureNode{
			Name: "skipped", Type: fixtures.TestNode,
			Test: &fixtures.FixtureTest{Name: "skipped", TestOS: "plan9"},
		}

		results, _, err := fixtures.RunNodes(context.Background(), []*fixtures.FixtureNode{node},
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Status).To(Equal(task.StatusSKIP))
		Expect(results[0].Error).To(ContainSubstring("requires os plan9"))
	})
})

// RunNode is the one dispatcher: what `gavel fixtures run` used to decide on
// its own — a recorder nothing implements, the daemon's port — is decided here
// for every caller.
var _ = Describe("fixtures.RunNode", func() {
	const portDocument = `---
codeBlocks: [bash]
---

### command: prints the daemon port

` + "```bash\necho port={{.port}}\n```" + `

- cel: stdout.contains('port=4242')
`

	It("refuses a recorder that has no implementation before the fixture runs", func() {
		previous := record.Implemented[record.KindSQL]
		record.Implemented[record.KindSQL] = false
		DeferCleanup(func() { record.Implemented[record.KindSQL] = previous })

		result := fixtures.RunNode(context.Background(), firstTest(parseDocument(passingDocument)),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir(), Record: &record.Spec{SQL: &record.SQLOptions{}}})

		Expect(result.Status).To(Equal(task.StatusERR))
		Expect(result.Error).To(ContainSubstring("record: [sql] is not implemented yet"))
	})

	It("exposes the daemon port to exec fixtures as {{.port}}", func() {
		result := fixtures.RunNode(context.Background(), firstTest(parseDocument(portDocument)),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir(), DaemonPort: 4242})

		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Stdout).To(ContainSubstring("port=4242"))
	})

	It("records the changed files a node was scoped to on its result", func() {
		changed := []string{"pkg/a.go", "pkg/b.go"}

		result := fixtures.RunNode(context.Background(), firstTest(parseDocument(passingDocument)),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir(), Changed: changed})

		Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
		Expect(result.Metadata).To(HaveKeyWithValue("changed_files", changed))
	})

	It("records nothing about scope on an unscoped run", func() {
		result := fixtures.RunNode(context.Background(), firstTest(parseDocument(passingDocument)),
			fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(result.Metadata).ToNot(HaveKey("changed_files"))
	})
})

var _ = Describe("fixtures.SplitFrontMatter", func() {
	It("returns the document unchanged when it opens with no delimiter", func() {
		fm, body, err := fixtures.SplitFrontMatter("# Heading\n\ntext\n")
		Expect(err).ToNot(HaveOccurred())
		Expect(fm).To(BeNil())
		Expect(body).To(Equal("# Heading\n\ntext\n"))
	})

	It("fails loudly on front matter that is never closed", func() {
		_, _, err := fixtures.SplitFrontMatter("---\ncodeBlocks: [bash]\n\n# Heading\n")
		Expect(err).To(MatchError(ContainSubstring("unterminated YAML front matter")))
	})

	It("separates parsed front matter from the body", func() {
		fm, body, err := fixtures.SplitFrontMatter("---\ncodeBlocks: [bash]\n---\n\n# Heading\n")
		Expect(err).ToNot(HaveOccurred())
		Expect(fm).ToNot(BeNil())
		Expect(fm.CodeBlocks).To(Equal([]string{"bash"}))
		Expect(body).To(Equal("\n# Heading"))
	})
})
