package types

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestResolveAgentChecks(t *testing.T) {
	tests := []struct {
		name          string
		project       AgentChecksConfig
		frontmatter   *AgentChecksConfig
		wantEnabled   bool
		wantRetry     string
		wantHasChecks bool
	}{
		{
			name:        "disabled by default when nothing set",
			wantEnabled: false,
		},
		{
			name:          "project enables with explicit checks",
			project:       AgentChecksConfig{Enabled: boolPtr(true), Test: &AgentTestConfig{Changed: true}},
			wantEnabled:   true,
			wantHasChecks: true,
		},
		{
			name:          "frontmatter enable overrides disabled project",
			project:       AgentChecksConfig{Enabled: boolPtr(false)},
			frontmatter:   &AgentChecksConfig{Enabled: boolPtr(true)},
			wantEnabled:   true,
			wantHasChecks: true, // enabled with no checks → default test+lint
		},
		{
			name:          "frontmatter disable overrides enabled project",
			project:       AgentChecksConfig{Enabled: boolPtr(true), Test: &AgentTestConfig{Changed: true}},
			frontmatter:   &AgentChecksConfig{Enabled: boolPtr(false)},
			wantEnabled:   false,
			wantHasChecks: true, // project test config survives the overlay
		},
		{
			name:          "project enable with no checks named defaults both",
			project:       AgentChecksConfig{Enabled: boolPtr(true)},
			wantEnabled:   true,
			wantHasChecks: true, // enabled with no checks → default test+lint
		},
		{
			name:          "frontmatter retry overrides project",
			project:       AgentChecksConfig{Enabled: boolPtr(true), Retry: "verify.summary.failed > 0", Test: &AgentTestConfig{Changed: true}},
			frontmatter:   &AgentChecksConfig{Retry: "verify.summary.warned > 0"},
			wantEnabled:   true,
			wantRetry:     "verify.summary.warned > 0",
			wantHasChecks: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveAgentChecks(tc.project, tc.frontmatter)
			if got.IsEnabled() != tc.wantEnabled {
				t.Errorf("IsEnabled() = %v, want %v", got.IsEnabled(), tc.wantEnabled)
			}
			if tc.wantRetry != "" && got.Retry != tc.wantRetry {
				t.Errorf("Retry = %q, want %q", got.Retry, tc.wantRetry)
			}
			if got.HasChecks() != tc.wantHasChecks {
				t.Errorf("HasChecks() = %v, want %v", got.HasChecks(), tc.wantHasChecks)
			}
		})
	}
}

func TestResolveAgentChecksEnabledDefaultsBothChecksToChangedFiles(t *testing.T) {
	got := ResolveAgentChecks(AgentChecksConfig{Enabled: boolPtr(true)}, nil)
	if got.Test == nil || !got.Test.Changed {
		t.Errorf("enabled default should run changed tests, got Test=%+v", got.Test)
	}
	if got.Lint == nil || !got.Lint.Changed {
		t.Errorf("enabled default should run changed lint, got Lint=%+v", got.Lint)
	}
}
