package fixtures

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// TestingT is the minimal testing interface supported by RunTesting.
// *testing.T satisfies this interface, and so does GinkgoT().
type TestingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

type testingRunner interface {
	Run(name string, f func(t *testing.T)) bool
}

// RunTesting runs fixtures and reports failures through testing.T-style APIs.
//
// When t is a *testing.T, markdown files, sections, tables, and rows are
// mirrored as nested t.Run nodes. For GinkgoT and other testing-like adapters
// that do not expose t.Run, this falls back to one aggregate failure.
func (r *Runner) RunTesting(t TestingT) *FixtureNode {
	t.Helper()

	tree, err := r.Run()
	if tree == nil {
		if err != nil {
			t.Fatalf("%s", formatRunFailure(nil, err))
		}
		return nil
	}

	if runner, ok := t.(testingRunner); ok {
		for _, child := range tree.Children {
			child := child
			runner.Run(testNodeName(child), func(st *testing.T) {
				reportTestingNode(st, child)
			})
		}
		if err != nil && !hasFailureResults(tree) {
			t.Fatalf("%s", formatRunFailure(tree, err))
		}
		return tree
	}

	if err != nil {
		t.Fatalf("%s", formatRunFailure(tree, err))
	}
	return tree
}

// RunGomega runs fixtures and reports failures through a Gomega instance.
func (r *Runner) RunGomega(g types.Gomega) *FixtureNode {
	tree, err := r.Run()
	g.Expect(err).ToNot(gomega.HaveOccurred(), formatRunFailure(tree, err))
	return tree
}

// RunGinkgo runs fixtures as a single aggregate Ginkgo spec helper.
//
// Use RegisterGinkgoSpecs when each markdown section/table row should be
// registered as its own Ginkgo node.
func (r *Runner) RunGinkgo() *FixtureNode {
	ginkgo.GinkgoHelper()
	tree, err := r.Run()
	if err != nil {
		ginkgo.Fail(formatRunFailure(tree, err), 1)
	}
	return tree
}

// RegisterGinkgoSpecs registers markdown files, sections, tables, and rows as
// Ginkgo Describe/It nodes. Call it during spec registration, not inside an It.
func (r *Runner) RegisterGinkgoSpecs() {
	tree, err := r.prepareFixtureTree()
	if err != nil {
		ginkgo.It("parse fixtures", func() {
			ginkgo.Fail(formatRunFailure(tree, err), 1)
		})
		return
	}

	ginkgo.Describe("Fixtures", ginkgo.Ordered, func() {
		ginkgo.BeforeAll(func() {
			ginkgo.GinkgoHelper()
			if err := r.setupGinkgoRun(); err != nil {
				ginkgo.Fail(err.Error(), 1)
			}
		})
		ginkgo.AfterAll(func() {
			r.stopDaemon()
		})

		for _, child := range tree.Children {
			registerGinkgoNode(r, child)
		}
	})
}

func registerGinkgoNode(r *Runner, node *FixtureNode) {
	if node == nil {
		return
	}

	if node.Test != nil {
		fixture := *node.Test
		ginkgo.It(testNodeName(node), func() {
			ginkgo.GinkgoHelper()
			result, err := r.runSingleFixture(fixture)
			node.Results = &result
			if err != nil {
				ginkgo.Fail(formatNodeFailure(node, &result, err), 1)
			}
			if isFailureResult(&result) {
				ginkgo.Fail(formatNodeFailure(node, &result, nil), 1)
			}
		})
		return
	}

	ginkgo.Describe(testNodeName(node), func() {
		for _, child := range node.Children {
			child := child
			registerGinkgoNode(r, child)
		}
	})
}

func (r *Runner) runSingleFixture(fixture FixtureTest) (FixtureResult, error) {
	ctx := flanksourceContext.NewContext(context.Background())
	return r.executeFixture(ctx, fixture)
}

func (r *Runner) setupGinkgoRun() error {
	if r.options.WorkDir == "" {
		r.options.WorkDir, _ = os.Getwd()
	}

	ctx := flanksourceContext.NewContext(context.Background())
	buildCmd := r.getBuildCommand()
	if buildCmd != "" {
		if err := r.executeBuildCommand(ctx, buildCmd); err != nil {
			return fmt.Errorf("build failed, skipping all fixtures: %w", err)
		}
	}

	daemonCmd := r.getDaemonCommand()
	if daemonCmd != "" {
		if err := r.startDaemon(ctx, daemonCmd); err != nil {
			return fmt.Errorf("daemon failed to start: %w", err)
		}
	}

	return nil
}

func reportTestingNode(t *testing.T, node *FixtureNode) {
	t.Helper()
	if node == nil {
		return
	}

	if node.Test != nil {
		if node.Results == nil {
			t.Fatalf("fixture did not produce a result: %s", node.Name)
		}
		if node.Results.Status == task.StatusSKIP {
			t.Skipf("%s", node.Results.Error)
		}
		if isFailureResult(node.Results) {
			t.Fatalf("%s", formatNodeFailure(node, node.Results, nil))
		}
		return
	}

	for _, child := range node.Children {
		child := child
		t.Run(testNodeName(child), func(st *testing.T) {
			reportTestingNode(st, child)
		})
	}
}

func hasFailureResults(tree *FixtureNode) bool {
	if tree == nil {
		return false
	}
	found := false
	tree.Walk(func(node *FixtureNode) {
		if node.Results != nil && isFailureResult(node.Results) {
			found = true
		}
	})
	return found
}

func isFailureResult(result *FixtureResult) bool {
	if result == nil {
		return false
	}
	switch result.Status {
	case task.StatusFailed, task.StatusFAIL, task.StatusERR, task.StatusCancelled:
		return true
	default:
		return false
	}
}

func testNodeName(node *FixtureNode) string {
	if node == nil || strings.TrimSpace(node.Name) == "" {
		return "fixture"
	}
	if node.Origin != nil && node.Origin.Kind == "table-row" && node.Origin.RowIndex > 0 {
		return fmt.Sprintf("row %d: %s", node.Origin.RowIndex, node.Name)
	}
	return node.Name
}

func formatRunFailure(tree *FixtureNode, err error) string {
	var b strings.Builder
	if err != nil {
		fmt.Fprintf(&b, "gavel fixtures failed: %v", err)
	} else {
		b.WriteString("gavel fixtures failed")
	}

	if tree != nil {
		stats := tree.GetStats()
		fmt.Fprintf(&b, "\nsummary: total=%d passed=%d failed=%d skipped=%d errors=%d", stats.Total, stats.Passed, stats.Failed, stats.Skipped, stats.Error)

		var failures []*FixtureNode
		tree.Walk(func(node *FixtureNode) {
			if node.Results != nil && isFailureResult(node.Results) {
				failures = append(failures, node)
			}
		})
		for _, node := range failures {
			b.WriteString("\n\n")
			b.WriteString(formatNodeFailure(node, node.Results, nil))
		}
	}

	return b.String()
}

func formatNodeFailure(node *FixtureNode, result *FixtureResult, err error) string {
	var b strings.Builder
	name := ""
	if node != nil {
		name = node.GetSectionPath()
	}
	if name == "" && result != nil {
		name = result.Name
	}
	if name == "" {
		name = "fixture"
	}
	fmt.Fprintf(&b, "fixture: %s", name)

	if node != nil && node.Origin != nil {
		if origin := formatOrigin(node.Origin); origin != "" {
			fmt.Fprintf(&b, "\norigin: %s", origin)
		}
	}
	if err != nil {
		fmt.Fprintf(&b, "\nerror: %v", err)
	}
	if result != nil {
		if result.Error != "" {
			fmt.Fprintf(&b, "\nresult: %s", result.Error)
		}
		if result.Command != "" {
			fmt.Fprintf(&b, "\ncommand: %s", result.Command)
		}
		if result.Stderr != "" {
			fmt.Fprintf(&b, "\nstderr:\n%s", truncateFixtureOutput(result.Stderr))
		}
		if result.Stdout != "" {
			fmt.Fprintf(&b, "\nstdout:\n%s", truncateFixtureOutput(result.Stdout))
		}
		if result.CELExpression != "" {
			fmt.Fprintf(&b, "\ncel: %s", result.CELExpression)
		}
	}
	return b.String()
}

func formatOrigin(origin *FixtureOrigin) string {
	if origin == nil {
		return ""
	}
	var parts []string
	if origin.File != "" {
		if origin.Line > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d", origin.File, origin.Line))
		} else {
			parts = append(parts, origin.File)
		}
	} else if origin.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d", origin.Line))
	}
	if origin.Kind != "" {
		parts = append(parts, origin.Kind)
	}
	if origin.SectionPath != "" {
		parts = append(parts, origin.SectionPath)
	}
	if origin.TableIndex > 0 {
		parts = append(parts, fmt.Sprintf("table=%d", origin.TableIndex))
	}
	if origin.RowIndex > 0 {
		parts = append(parts, fmt.Sprintf("row=%d", origin.RowIndex))
	}
	return strings.Join(parts, " ")
}

func truncateFixtureOutput(s string) string {
	const max = 1200
	s = strings.TrimRight(s, "\n")
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... truncated ..."
}
