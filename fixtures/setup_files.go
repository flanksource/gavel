package fixtures

import (
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/clicky/task"
	dbcontext "github.com/flanksource/commons-db/context"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
)

// fileSetups is the `setup:` state of one run, one entry per markdown file
// that declared a setup — the seam Runner and RunNodes share. A setup that
// reads `{{.params…}}` is deferred until the file's params step passed
// (afterParams); every other one is prepared up front (prepare).
type fileSetups struct {
	// host resolves a deferred setup's connection references ahead of gavel's
	// layered store (RunOptions.Connections); nil for Runner, which has none.
	host     dbcontext.ConnectionResolver
	mu       sync.Mutex
	prepared map[string]*PreparedSetup
	deferred map[string]deferredSetup
}

// declaredSetup is the `setup:` one file declared, with the markdown path to
// key it on and the directory to anchor its relative paths against.
type declaredSetup struct {
	spec      *SetupSpec
	mdPath    string
	sourceDir string
}

// declaredSetups lists, in tree order, the setup each file under nodes
// declared. Frontmatter is file-level, so it is reachable only through the
// tests that carry it; the first test of a file wins because every test in a
// file shares the same parsed frontmatter.
//
// A file whose tests were all filtered out contributes no node to walk and so
// gets no setup prepared: `--filter` should not clone a repo it has no tests for.
func declaredSetups(nodes []*FixtureNode) []declaredSetup {
	var declared []declaredSetup
	seen := map[string]bool{}
	for _, root := range nodes {
		root.Walk(func(node *FixtureNode) {
			if node.Test == nil || node.Test.Setup == nil {
				return
			}
			file := nodeFile(node)
			if seen[file] {
				return
			}
			seen[file] = true
			declared = append(declared, declaredSetup{spec: node.Test.Setup, mdPath: file, sourceDir: node.Test.SourceDir})
		})
	}
	return declared
}

// nodeFile is the markdown file a node came from, empty when it never touched
// disk (ParseMarkdownDocument).
func nodeFile(node *FixtureNode) string {
	if node == nil || node.Origin == nil {
		return ""
	}
	return node.Origin.File
}

// prepare prepares each file's eager setup under nodes and defers the ones
// that read params. It runs before `build:` deliberately: a setup can relocate
// the run into a worktree, and a build that ran in the original repo would
// build tree A while the tests exercise tree B — and pass.
func (s *fileSetups) prepare(ctx flanksourceContext.Context, nodes []*FixtureNode) error {
	for _, declared := range declaredSetups(nodes) {
		if setupReadsParams(declared.spec) {
			if err := s.deferSetup(nodes, declared); err != nil {
				return err
			}
			continue
		}
		prepared, err := prepareSetup(ctx, declared.mdPath, declared.sourceDir, declared.spec, nil)
		if err != nil {
			return err
		}
		s.store(declared.mdPath, prepared)
	}
	return nil
}

// deferSetup records a params-reading setup instead of preparing it. It fails
// when the file declares no params to read, or when the setup would relocate
// the run (a checkout or cwd): those must be prepared before `build:`, which
// runs before any step can answer the form.
func (s *fileSetups) deferSetup(nodes []*FixtureNode, declared declaredSetup) error {
	hasParams := false
	for _, root := range nodes {
		root.Walk(func(node *FixtureNode) {
			hasParams = hasParams || (nodeFile(node) == declared.mdPath && node.Test != nil && node.Test.Step != nil && node.Test.Step.ID == ParamsStepID)
		})
	}
	if !hasParams {
		return fmt.Errorf("%s: setup: reads {{.params…}} or a $cel value but the fixture declares no params:", declared.mdPath)
	}
	relocates, err := declared.spec.declares("checkout", "cwd")
	if err != nil {
		return fmt.Errorf("%s: setup: %w", declared.mdPath, err)
	}
	if relocates {
		return fmt.Errorf("%s: setup: a setup that reads {{.params…}} or a $cel value is prepared after the params form, so it cannot declare checkout: or cwd:", declared.mdPath)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.deferred == nil {
		s.deferred = map[string]deferredSetup{}
	}
	s.deferred[declared.mdPath] = deferredSetup{spec: declared.spec, sourceDir: declared.sourceDir}
	return nil
}

func (s *fileSetups) store(mdPath string, prepared *PreparedSetup) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.prepared == nil {
		s.prepared = map[string]*PreparedSetup{}
	}
	s.prepared[mdPath] = prepared
}

// get is the file's prepared setup, nil while none is (yet).
func (s *fileSetups) get(mdPath string) *PreparedSetup {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prepared[mdPath]
}

// count is how many setups are prepared.
func (s *fileSetups) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.prepared)
}

// afterParams prepares the file's deferred setup once its params step passed:
// the setup is rendered against the answered params, prepared, and its
// environment exported into the file's scope for the steps that follow. A
// failure errors the params step, which skips every later step of the file.
func (s *fileSetups) afterParams(ctx flanksourceContext.Context, test FixtureTest, file string, scope *StepScope, result FixtureResult) FixtureResult {
	if test.Step == nil || test.Step.ID != ParamsStepID || result.Status != task.StatusPASS {
		return result
	}
	s.mu.Lock()
	deferred, ok := s.deferred[file]
	delete(s.deferred, file)
	s.mu.Unlock()
	if !ok {
		return result
	}
	prepared, err := prepareParamsSetup(ctx, file, deferred, scope, s.host)
	if err == nil {
		s.store(file, prepared)
		err = exportSetupEnv(scope, prepared)
	}
	if err != nil {
		scope.MarkFailed(ParamsStepID)
		return result.Errorf(err, "prepare setup")
	}
	return result
}

// exportSetupEnv exports a setup prepared after the scope was seeded into it,
// so the steps that follow see its environment.
func exportSetupEnv(scope *StepScope, prepared *PreparedSetup) error {
	for name, value := range prepared.Env {
		if err := scope.Env.Export("setup", name, value, true); err != nil {
			return err
		}
	}
	return nil
}

// cleanup tears every prepared setup down, joining the failures.
func (s *fileSetups) cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var cleanupErr error
	for file, prepared := range s.prepared {
		if err := prepared.Cleanup(); err != nil {
			logger.Warnf("%s: setup cleanup failed: %v", file, err)
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("%s: setup cleanup: %w", file, err))
		}
	}
	return cleanupErr
}
