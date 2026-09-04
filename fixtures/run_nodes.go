package fixtures

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/goccy/go-yaml"
)

// RunNodes executes every test node in a fixture tree and returns one result per
// test, in tree order, plus the execution snapshot the run ended on.
//
// It is the one node runner: `gavel fixtures run` drives whole files through
// Runner, and every caller that already holds a parsed tree — a TODO's
// `## Verification` section, the definition-of-done verifier, the reproduction
// steps — goes through here instead of re-implementing the walk and the
// node-type dispatch. Two dispatchers meant a fixture kind added to one of them
// silently never ran in the other.
//
// Progress is reported to the ProgressSink on ctx (see WithProgressSink); the
// returned snapshot is the same tree the sink last saw, so a caller that wants
// only the final state need not register one.
//
// A node that could not be dispatched is a result with an error status, not a
// dropped node: the caller's verdict has to see it. An error return means the
// walk itself could not continue (a progress sink that failed), and the results
// collected so far are returned with it.
func RunNodes(ctx context.Context, nodes []*FixtureNode, opts RunOptions) ([]FixtureResult, *ExecutionSnapshot, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("fixtures: RunNodes requires a context")
	}
	if opts.Context == nil {
		opts.Context = ctx
	}
	if opts.Evaluator == nil {
		evaluator, err := NewCELEvaluator()
		if err != nil {
			return nil, nil, fmt.Errorf("fixtures: create CEL evaluator: %w", err)
		}
		opts.Evaluator = evaluator
	}

	reporter := NewExecutionReporter(nodes, opts.WorkDir, nil, ProgressSinkFromContext(ctx))
	if err := reporter.Publish(ctx); err != nil {
		return nil, nil, fmt.Errorf("fixtures: publish queued fixture progress: %w", err)
	}
	results, err := runNodeTree(ctx, nodes, opts, reporter)
	snapshot := reporter.Snapshot()
	return results, &snapshot, err
}

func runNodeTree(ctx context.Context, nodes []*FixtureNode, opts RunOptions, reporter *ExecutionReporter) ([]FixtureResult, error) {
	var results []FixtureResult
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if node.Test != nil {
			if err := reporter.StartFixture(ctx, node); err != nil {
				return results, err
			}
			nodeOpts := opts
			nodeOpts.Progress = func(done, total int) error {
				return reporter.UpdateFixture(ctx, node, done, total)
			}
			result := RunNode(ctx, *node.Test, nodeOpts)
			results = append(results, result)
			if err := reporter.CompleteFixture(ctx, node, result); err != nil {
				return results, err
			}
		}
		childResults, err := runNodeTree(ctx, node.Children, opts, reporter)
		results = append(results, childResults...)
		if err != nil {
			return results, err
		}
	}
	return results, nil
}

// RunNode runs one fixture test through the node-type dispatch every caller
// shares — `gavel fixtures run` and RunNodes alike, so a fixture kind added
// here runs everywhere: a recorder the fixture declares but nothing implements
// is refused, a skip declaration honoured, then `ai` steps go via the
// AIStepRunner hook, `yaml test` / `yaml lint` fences via the runner-step
// hooks, and everything else via the type registry.
func RunNode(ctx context.Context, test FixtureTest, opts RunOptions) FixtureResult {
	result := dispatchNode(ctx, test, opts)
	if len(opts.Changed) > 0 {
		if result.Metadata == nil {
			result.Metadata = map[string]interface{}{}
		}
		result.Metadata["changed_files"] = opts.Changed
	}
	return result
}

func dispatchNode(ctx context.Context, test FixtureTest, opts RunOptions) FixtureResult {
	if missing := opts.Record.Missing(); len(missing) > 0 {
		return nodeError(test, fmt.Sprintf("record: %v is not implemented yet", missing))
	}
	if reason := test.ShouldSkip(); reason != "" {
		return FixtureResult{Name: test.Name, Status: task.StatusSKIP, Test: test, Error: reason}
	}
	if test.IsAIStep() {
		if AIStepRunner == nil {
			return nodeError(test, "AI step runner not registered; import _ \"github.com/flanksource/gavel/fixtures/types\"")
		}
		return AIStepRunner(test, opts)
	}
	if test.IsRunnerStep() {
		runner := TestStepRunner
		if test.IsLintStep() {
			runner = LintStepRunner
		}
		if runner == nil {
			return nodeError(test, "runner step hook not registered; import _ \"github.com/flanksource/gavel/fixtures/types\"")
		}
		return runner(test, opts)
	}
	fixtureType, err := DefaultRegistry.GetForFixture(test)
	if err != nil {
		return nodeError(test, err.Error())
	}
	if opts.DaemonPort > 0 {
		if test.TemplateVars == nil {
			test.TemplateVars = map[string]any{}
		}
		test.TemplateVars["port"] = strconv.Itoa(opts.DaemonPort)
	}
	return fixtureType.Run(ctx, test, opts)
}

func nodeError(test FixtureTest, msg string) FixtureResult {
	return FixtureResult{Name: test.Name, Status: task.StatusERR, Test: test, Error: msg}
}

// ParseMarkdownDocument parses a complete fixture markdown document — YAML front
// matter included — into a fixture tree. It is the string twin of
// ParseMarkdownFixturesWithTree, for callers whose document never touched disk
// (a workflow's `verify.fixture`, an issue body, a document on stdin).
func ParseMarkdownDocument(name, markdown, sourceDir string) (*FixtureNode, error) {
	frontMatter, body, err := SplitFrontMatter(markdown)
	if err != nil {
		return nil, err
	}
	if frontMatter != nil {
		frontMatter.CleanMetadata()
	}
	return ParseMarkdownContentWithTree(name, body, sourceDir, frontMatter)
}

// SplitFrontMatter separates a markdown document's leading `---` YAML block from
// its body. A document that does not open with a delimiter has no front matter,
// which is not an error; malformed YAML inside one is.
func SplitFrontMatter(markdown string) (*FrontMatter, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	scanner.Buffer(make([]byte, 0, 64*1024), maxFrontMatterLineBytes)
	if !scanner.Scan() {
		return nil, markdown, nil
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return nil, markdown, nil
	}

	var frontMatterLines []string
	closed := false
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "---" {
			closed = true
			break
		}
		frontMatterLines = append(frontMatterLines, scanner.Text())
	}
	if !closed {
		return nil, "", fmt.Errorf("unterminated YAML front matter: no closing --- delimiter")
	}

	var frontMatter FrontMatter
	if err := yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &frontMatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML front-matter: %w", err)
	}

	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return &frontMatter, strings.Join(contentLines, "\n"), nil
}

// maxFrontMatterLineBytes bounds one scanned line. A fixture document can embed
// a whole YAML config or a base64 payload on one line, so the bound is generous;
// without one bufio's 64KiB default turns a long line into a silent truncation.
const maxFrontMatterLineBytes = 4 << 20
