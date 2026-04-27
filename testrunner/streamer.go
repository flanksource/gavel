package testrunner

import (
	"sync"

	"github.com/flanksource/gavel/testrunner/parsers"
)

type TestStreamer struct {
	mu             sync.Mutex
	completedTests []parsers.Test
	pendingPkgs    []parsers.Test
	fixtureTests   []parsers.Test
	updates        chan<- []parsers.Test
	closeUpdates   bool
	closed         bool
}

func NewTestStreamer(updates chan<- []parsers.Test) *TestStreamer {
	return newTestStreamer(updates, true)
}

func NewSharedTestStreamer(updates chan<- []parsers.Test) *TestStreamer {
	return newTestStreamer(updates, false)
}

func newTestStreamer(updates chan<- []parsers.Test, closeUpdates bool) *TestStreamer {
	return &TestStreamer{
		updates:      updates,
		closeUpdates: closeUpdates,
	}
}

func (s *TestStreamer) SetPackageOutline(pkgs []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.pendingPkgs = pkgs
	s.buildAndSendLocked()
}

// UpdateGinkgoProgress implements parsers.GinkgoProgressSink. Ginkgo
// streaming parsers call this while a suite is still running with a summary
// Test whose Children reflect observed spec completions so far.
func (s *TestStreamer) UpdateGinkgoProgress(progress parsers.Test) {
	s.UpdatePackageProgress(progress.PackagePath, progress.Framework, progress)
}

// UpdatePackageProgress replaces the pending outline entry for (pkgPath,
// framework) with a richer tree containing currently-running and already-
// finished children from the streaming parser. Called while the subprocess
// is still alive — final results come through CompletePackage once the
// authoritative parser has run.
func (s *TestStreamer) UpdatePackageProgress(pkgPath string, framework parsers.Framework, progress parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	for i, p := range s.pendingPkgs {
		if p.PackagePath == pkgPath && p.Framework == framework {
			progress.PackagePath = pkgPath
			progress.Framework = framework
			if progress.Name == "" {
				progress.Name = p.Name
			}
			s.pendingPkgs[i] = progress
			s.buildAndSendLocked()
			return
		}
	}
}

func (s *TestStreamer) CompletePackage(pkgPath string, framework parsers.Framework, results parsers.TestSuiteResults) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}

	// Remove matching pending package
	filtered := s.pendingPkgs[:0:0]
	for _, p := range s.pendingPkgs {
		if p.PackagePath != pkgPath || p.Framework != framework {
			filtered = append(filtered, p)
		}
	}
	s.pendingPkgs = filtered

	// Add all completed test results (UI always shows everything)
	for _, tr := range results {
		s.completedTests = append(s.completedTests, tr.Tests...)
	}

	s.buildAndSendLocked()
}

func (s *TestStreamer) SetFixtureOutline(tests []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.fixtureTests = tests
	s.buildAndSendLocked()
}

func (s *TestStreamer) UpdateFixtures(tests []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.fixtureTests = tests
	s.buildAndSendLocked()
}

func (s *TestStreamer) buildAndSendLocked() {
	tree := parsers.BuildTestTree(s.completedTests)
	tree = append(tree, s.pendingPkgs...)
	tree = append(tree, s.fixtureTests...)
	s.sendLocked(tree)
}

func (s *TestStreamer) sendLocked(tree []parsers.Test) {
	if s.updates == nil || s.closed {
		return
	}
	select {
	case s.updates <- tree:
	default:
		// Drop if channel is full; next update will carry latest state
	}
}

func (s *TestStreamer) Done() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.updates == nil || s.closed {
		return
	}
	s.buildAndSendLocked()
	s.closed = true
	if s.closeUpdates {
		close(s.updates)
	}
}
