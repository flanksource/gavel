package types

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestResolveAgentChecks(t *testing.T) {
	tests := []struct {
		name          string
		project       AgentChecksConfig
		frontmatter   *AgentChecksConfig
		forceEnabled  bool
		wantEnabled   bool
		wantMaxIters  int
		wantHasChecks bool
	}{
		{
			name:         "disabled by default when nothing set",
			wantEnabled:  false,
			wantMaxIters: DefaultMaxCheckIterations,
		},
		{
			name:          "project enables with explicit checks",
			project:       AgentChecksConfig{Enabled: boolPtr(true), Test: &AgentTestConfig{Changed: true}},
			wantEnabled:   true,
			wantMaxIters:  DefaultMaxCheckIterations,
			wantHasChecks: true,
		},
		{
			name:          "frontmatter enable overrides disabled project",
			project:       AgentChecksConfig{Enabled: boolPtr(false)},
			frontmatter:   &AgentChecksConfig{Enabled: boolPtr(true)},
			wantEnabled:   true,
			wantMaxIters:  DefaultMaxCheckIterations,
			wantHasChecks: true, // enabled with no checks → default test+lint
		},
		{
			name:          "frontmatter disable overrides enabled project",
			project:       AgentChecksConfig{Enabled: boolPtr(true), Test: &AgentTestConfig{Changed: true}},
			frontmatter:   &AgentChecksConfig{Enabled: boolPtr(false)},
			wantEnabled:   false,
			wantMaxIters:  DefaultMaxCheckIterations,
			wantHasChecks: true, // project test config survives the overlay
		},
		{
			name:          "force enable from flag turns on a silent project",
			forceEnabled:  true,
			wantEnabled:   true,
			wantMaxIters:  DefaultMaxCheckIterations,
			wantHasChecks: true, // forced-on with no checks → default test+lint
		},
		{
			name:          "frontmatter maxIterations overrides project",
			project:       AgentChecksConfig{Enabled: boolPtr(true), MaxIterations: 2, Test: &AgentTestConfig{Changed: true}},
			frontmatter:   &AgentChecksConfig{MaxIterations: 7},
			wantEnabled:   true,
			wantMaxIters:  7,
			wantHasChecks: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveAgentChecks(tc.project, tc.frontmatter, tc.forceEnabled)
			if got.IsEnabled() != tc.wantEnabled {
				t.Errorf("IsEnabled() = %v, want %v", got.IsEnabled(), tc.wantEnabled)
			}
			if got.MaxIterations != tc.wantMaxIters {
				t.Errorf("MaxIterations = %d, want %d", got.MaxIterations, tc.wantMaxIters)
			}
			if got.HasChecks() != tc.wantHasChecks {
				t.Errorf("HasChecks() = %v, want %v", got.HasChecks(), tc.wantHasChecks)
			}
		})
	}
}

func TestResolveAgentChecksForcedDefaultsBothChecks(t *testing.T) {
	got := ResolveAgentChecks(AgentChecksConfig{}, nil, true)
	if got.Test == nil || !got.Test.Changed {
		t.Errorf("forced default should run changed tests, got Test=%+v", got.Test)
	}
	if got.Lint == nil || !got.Lint.Changed {
		t.Errorf("forced default should run changed lint, got Lint=%+v", got.Lint)
	}
}
