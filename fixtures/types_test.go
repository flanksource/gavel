package fixtures

import (
	"runtime"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/stretchr/testify/assert"
)

func TestStatsAddCountsStatusFAIL(t *testing.T) {
	tests := []struct {
		name           string
		status         task.Status
		expectedFailed int
		expectedPassed int
	}{
		{name: "StatusFAIL counts as failed", status: task.StatusFAIL, expectedFailed: 1},
		{name: "StatusFailed counts as failed", status: task.StatusFailed, expectedFailed: 1},
		{name: "StatusPASS counts as passed", status: task.StatusPASS, expectedPassed: 1},
		{name: "StatusERR counts as error", status: task.StatusERR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Stats{}
			result := &FixtureResult{Status: tt.status}
			s = s.Add(result)
			assert.Equal(t, 1, s.Total)
			assert.Equal(t, tt.expectedFailed, s.Failed)
			assert.Equal(t, tt.expectedPassed, s.Passed)
		})
	}
}

func TestFixtureResultStatsCountsStatusFAIL(t *testing.T) {
	result := FixtureResult{Status: task.StatusFAIL}
	stats := result.Stats()
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Total)
	assert.Equal(t, 0, stats.Passed)
}

func TestStatsVisitCountsStatusFAIL(t *testing.T) {
	s := &Stats{}
	node := &FixtureNode{
		Results: &FixtureResult{Status: task.StatusFAIL},
	}
	s.Visit(node)
	assert.Equal(t, 1, s.Total)
	assert.Equal(t, 1, s.Failed)
}

func TestStatsHasFailuresWithStatusFAIL(t *testing.T) {
	result := FixtureResult{Status: task.StatusFAIL}
	stats := result.Stats()
	assert.True(t, stats.HasFailures())
	assert.False(t, stats.IsOK())
}

func TestFrontMatterShouldSkip(t *testing.T) {
	tests := []struct {
		name     string
		fm       FrontMatter
		wantSkip bool
	}{
		{
			name:     "no constraints",
			fm:       FrontMatter{},
			wantSkip: false,
		},
		{
			name:     "matching os",
			fm:       FrontMatter{OS: runtime.GOOS},
			wantSkip: false,
		},
		{
			name:     "non-matching os",
			fm:       FrontMatter{OS: "plan9"},
			wantSkip: true,
		},
		{
			name:     "negated os excludes current",
			fm:       FrontMatter{OS: "!" + runtime.GOOS},
			wantSkip: true,
		},
		{
			name:     "negated os allows other",
			fm:       FrontMatter{OS: "!plan9"},
			wantSkip: false,
		},
		{
			name:     "matching arch",
			fm:       FrontMatter{Arch: runtime.GOARCH},
			wantSkip: false,
		},
		{
			name:     "non-matching arch",
			fm:       FrontMatter{Arch: "mips"},
			wantSkip: true,
		},
		{
			name:     "skip command returns true",
			fm:       FrontMatter{Skip: "true"},
			wantSkip: true,
		},
		{
			name:     "skip command returns false",
			fm:       FrontMatter{Skip: "false"},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := tt.fm.ShouldSkip()
			if tt.wantSkip {
				assert.NotEmpty(t, reason)
			} else {
				assert.Empty(t, reason)
			}
		})
	}
}
