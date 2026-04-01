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
}

func NewTestStreamer(updates chan<- []parsers.Test) *TestStreamer {
	return &TestStreamer{
		updates: updates,
	}
}

func (s *TestStreamer) SetPackageOutline(pkgs []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingPkgs = pkgs
	s.buildAndSend()
}

func (s *TestStreamer) CompletePackage(pkgPath string, framework parsers.Framework, results parsers.TestSuiteResults) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove matching pending package
	filtered := s.pendingPkgs[:0:0]
	for _, p := range s.pendingPkgs {
		if !(p.PackagePath == pkgPath && p.Framework == framework) {
			filtered = append(filtered, p)
		}
	}
	s.pendingPkgs = filtered

	// Add all completed test results (UI always shows everything)
	for _, tr := range results {
		s.completedTests = append(s.completedTests, tr.Tests...)
	}

	s.buildAndSend()
}

func (s *TestStreamer) SetFixtureOutline(tests []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fixtureTests = tests
	s.buildAndSend()
}

func (s *TestStreamer) UpdateFixtures(tests []parsers.Test) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fixtureTests = tests
	s.buildAndSend()
}

func (s *TestStreamer) buildAndSend() {
	tree := parsers.BuildTestTree(s.completedTests)
	tree = append(tree, s.pendingPkgs...)
	tree = append(tree, s.fixtureTests...)
	s.send(tree)
}

func (s *TestStreamer) send(tree []parsers.Test) {
	if s.updates == nil {
		return
	}
	select {
	case s.updates <- tree:
	default:
		// Drop if channel is full; next update will carry latest state
	}
}

func (s *TestStreamer) Done() {
	if s.updates == nil {
		return
	}
	s.mu.Lock()
	s.buildAndSend()
	s.mu.Unlock()
	close(s.updates)
}
