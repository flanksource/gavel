package fixtures

import (
	"context"
	"fmt"
	"strings"

	flanksourceContext "github.com/flanksource/commons/context"
)

// nodeSetup is the `setup:` RunNodes prepared for its nodes' document,
// through the same fileSetups seam Runner uses. A nil *nodeSetup — the caller
// passed a prepared opts.Setup, or the document declares none — does nothing.
type nodeSetup struct {
	setups *fileSetups
	file   string
	// err is the eager setup's prepare failure; the first node reports it.
	err error
}

// prepareNodeSetup prepares the eager setup of the document nodes belong to
// and defers a params-reading one. RunNodes runs every node in one scope, so
// nodes spanning several files that each declare a setup are refused.
func prepareNodeSetup(ctx context.Context, nodes []*FixtureNode, opts RunOptions) (*nodeSetup, error) {
	if opts.Setup != nil {
		return nil, nil
	}
	declared := declaredSetups(nodes)
	switch len(declared) {
	case 0:
		return nil, nil
	case 1:
	default:
		files := make([]string, len(declared))
		for i, d := range declared {
			files[i] = d.mdPath
		}
		return nil, fmt.Errorf("fixtures: RunNodes runs its nodes in one scope, but %d files declare setup: (%s); run each file separately", len(declared), strings.Join(files, ", "))
	}
	setup := &nodeSetup{setups: &fileSetups{host: opts.Connections}, file: declared[0].mdPath}
	setup.err = setup.setups.prepare(flanksourceContext.NewContext(ctx), nodes)
	return setup, nil
}

// prepared is the document's prepared setup, nil while none is (yet).
func (s *nodeSetup) prepared() *PreparedSetup {
	if s == nil {
		return nil
	}
	return s.setups.get(s.file)
}

// attach exports an eager setup's environment into a scope the caller seeded
// and registers the teardown of every setup the run prepares. Registered
// first, it runs after every step's cleanup, which may still need the
// setup's environment (a kubeconfig, a credentials file).
func (s *nodeSetup) attach(scope *StepScope, seeded bool) error {
	if s == nil {
		return nil
	}
	if prepared := s.prepared(); prepared != nil && !seeded {
		if err := exportSetupEnv(scope, prepared); err != nil {
			return fmt.Errorf("fixtures: export setup environment: %w", err)
		}
	}
	return scope.AddCleanup("setup", func(context.Context) error { return s.setups.cleanup() })
}

// run runs one test node under the document's setup: a failed eager setup
// errors the first node and skips the rest, and a deferred setup is prepared
// once the params step passed.
func (s *nodeSetup) run(ctx context.Context, test FixtureTest, opts RunOptions) FixtureResult {
	if s == nil {
		return RunNode(ctx, test, opts)
	}
	if err := s.err; err != nil {
		s.err = nil
		failed := test.Name
		if test.Step != nil {
			failed = stepID(test)
		}
		opts.Scope.MarkFailed(failed)
		return nodeError(test, "prepare setup: "+err.Error())
	}
	opts.Setup = s.prepared()
	result := RunNode(ctx, test, opts)
	return s.setups.afterParams(flanksourceContext.NewContext(ctx), test, s.file, opts.Scope, result)
}
