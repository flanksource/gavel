// Package verifier registers gavel's fixture engine as captain's `fixture`
// verifier kind: a workflow that declares `verify.fixture` gets that markdown
// document executed in-process by the fixture runner, and answered with a typed
// api.VerifyReport.
//
// It deliberately imports only fixtures and its step runners, never todos: the
// same verifier serves an agent loop, `gavel todos check`, and the
// `gavel fixtures run --format captain-verify-report` external-runner contract,
// none of which should need a TODO to exist.
package verifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"

	"github.com/flanksource/gavel/fixtures"

	// Registers the exec/query fixture types and the ai / `yaml test` /
	// `yaml lint` step hooks the documents dispatch to. Without it a declared
	// step errors instead of running, which is a false red rather than a false
	// green — but a false red on every run all the same.
	_ "github.com/flanksource/gavel/fixtures/types"
)

func init() { capverify.Register(capverify.KindFixture, New) }

// New is the capverify.Factory for KindFixture. It returns exactly one plugin —
// the document is one definition of done, however many steps it declares — and
// nothing at all when no fixture is declared.
//
// A `scope: changed` workflow needs no vetting here: every node kind the
// fixture engine dispatches honours the changed set (see
// fixtures.RunOptions.Changed), so there is no document this verifier would
// have to refuse for ignoring it.
func New(_ context.Context, spec api.Verify, opts capverify.Options) ([]*capverify.Plugin, error) {
	document := strings.TrimSpace(spec.Fixture)
	if document == "" {
		return nil, nil
	}
	verifier, err := NewVerifier(document)
	if err != nil {
		return nil, err
	}
	verifier.Timeout = opts.Timeout
	return []*capverify.Plugin{capverify.New(Name, verifier)}, nil
}

// Verifier executes one fixture markdown document as a definition of done.
type Verifier struct {
	// Document is the fixture markdown, verbatim — front matter included.
	Document string
	// Spec is the runtime an `ai` step grades on. It is nil on the registry
	// path: captain's Factory signature carries the declaration (api.Verify) and
	// the run's confinement (verify.Options) and has no room for a grading
	// runtime, so a document dispatched through the registry says how it wants to
	// be graded in its own `ai:` front matter — which is also the only channel
	// that survives the external-runner contract. Callers that drive the verifier
	// directly (`gavel todos check`, `gavel fixtures run`) set it from the
	// resolved verification chain.
	Spec *api.Spec
	// Timeout is the confinement bound the caller applies to the whole document.
	// Zero means the caller's context is the only bound.
	Timeout time.Duration

	retry    *retryPredicate
	progress func(api.VerifyReport)
}

// NewVerifier compiles a fixture document's declaration: its retry predicate is
// checked now so a broken definition of done fails before an agent turn is spent
// on it, and the tree is parsed per run because the run's cwd is what relative
// paths and file globs in the document resolve against.
func NewVerifier(document string) (*Verifier, error) {
	frontMatter, _, err := fixtures.SplitFrontMatter(document)
	if err != nil {
		return nil, fmt.Errorf("verify fixture: %w", err)
	}
	var declared string
	if frontMatter != nil && frontMatter.Verify != nil {
		declared = frontMatter.Verify.Retry
	}
	retry, err := compileRetry(declared)
	if err != nil {
		return nil, err
	}
	return &Verifier{Document: document, retry: retry}, nil
}

// SetProgress implements capverify.ProgressVerifier: each execution snapshot the
// fixture engine publishes is forwarded as a `running` report, so a reader
// watches the same tree fill in that it will read the verdict from.
func (v *Verifier) SetProgress(fn func(api.VerifyReport)) { v.progress = fn }

// Verify executes the document in cwd. A non-empty changed set — the files the
// change under verification touched, when the workflow scoped verification to
// them — reaches every step through fixtures.RunOptions.Changed: test steps
// narrow to the packages those files affect, lint steps to the files, the
// acceptance-criteria grader is told which files to judge, and exec fixtures
// see them as the CEL variable `changed_files`. Nil verifies the whole tree.
func (v *Verifier) Verify(ctx context.Context, cwd string, changed []string) (capverify.Verdict, error) {
	if v.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v.Timeout)
		defer cancel()
	}
	tree, err := fixtures.ParseMarkdownDocument("definition-of-done", v.Document, cwd)
	if err != nil {
		return capverify.Verdict{}, fmt.Errorf("verify fixture: %w", err)
	}

	results, snapshot, runErr := fixtures.RunNodes(v.progressContext(ctx), tree.Children,
		fixtures.RunOptions{WorkDir: cwd, Spec: v.Spec, Changed: changed})
	if runErr != nil {
		// The walk itself could not finish, so some declared step never ran. That
		// is a scheduling failure, and Report says so from the snapshot — the
		// error rides along as the reason rather than aborting the hook, because
		// the loop still has to record what did run.
		report := Report(results, snapshot)
		report.Reason = strings.TrimSpace(strings.TrimSuffix(report.Reason, ";") + "; " + runErr.Error())
		report.Feedback = report.Reason
		report.State, report.Ran, report.Passed = api.VerifyStateErrored, false, false
		if err := report.Validate(); err != nil {
			return capverify.Verdict{}, err
		}
		return capverify.Verdict{Reason: report.Reason, Feedback: report.Feedback, Report: &report}, nil
	}

	report := Report(results, snapshot)
	if err := v.retry.applyRetry(&report); err != nil {
		return capverify.Verdict{}, err
	}
	if err := report.Validate(); err != nil {
		return capverify.Verdict{}, fmt.Errorf("verify fixture: %w", err)
	}
	return capverify.Verdict{
		OK: report.Passed, Reason: report.Reason, Feedback: report.Feedback, Report: &report,
	}, nil
}

// progressContext installs the sink that turns fixture execution snapshots into
// running reports. A caller that registered no progress sink gets the plain
// context, so the engine does no snapshot work nobody reads.
//
// Only a snapshot that is still in flight is published: the run's last one
// already holds a verdict, and the verdict report that follows it is the only
// thing allowed to stamp that. A running report that fails captain's own
// validation stops the walk rather than reaching a reader, because a report the
// host would reject mid-stream is a bug here, not a state the checks got into.
func (v *Verifier) progressContext(ctx context.Context) context.Context {
	if v.progress == nil {
		return ctx
	}
	return fixtures.WithProgressSink(ctx, func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
		report, live := RunningReport(snapshot)
		if !live {
			return nil
		}
		if err := report.Validate(); err != nil {
			return fmt.Errorf("verify fixture: running report: %w", err)
		}
		v.progress(report)
		return nil
	})
}

// neverRan names the steps a run left queued. It reads the snapshot rather than
// the results because a step that never started produced no result to inspect.
//
// Only executable leaves count. A markdown document can hold a heading with no
// steps under it — a `## Notes` section, a heading that only introduces the next
// one — and the tracker gives those a node too; reading them as steps that never
// ran would turn every prose section into a scheduling failure.
func neverRan(snapshot *fixtures.ExecutionSnapshot) []string {
	if snapshot == nil {
		return nil
	}
	var stalled []string
	var walk func(node *fixtures.ExecutionNode)
	walk = func(node *fixtures.ExecutionNode) {
		if node == nil {
			return
		}
		if len(node.Children) > 0 {
			for _, child := range node.Children {
				walk(child)
			}
			return
		}
		if node.State == fixtures.ExecutionQueued && executable(node.Kind) {
			stalled = append(stalled, node.Name)
		}
	}
	walk(snapshot.Root)
	return stalled
}

// executable reports whether a progress node is a step that was meant to run,
// as opposed to a structural node the tracker mirrors from the markdown.
func executable(kind fixtures.ExecutionKind) bool {
	switch kind {
	case fixtures.ExecutionKindRoot, fixtures.ExecutionKindFile,
		fixtures.ExecutionKindSection, fixtures.ExecutionKindTable:
		return false
	default:
		return true
	}
}
