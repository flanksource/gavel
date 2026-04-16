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
	closed         bool
}

func NewTestStreamer(updates chan<- []parsers.Test) *TestStreamer {
	return &TestStreamer{
		updates: updates,
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
	close(s.updates)
}
