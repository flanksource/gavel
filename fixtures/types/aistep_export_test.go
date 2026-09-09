package types

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
)

func ResolveAIStepSpecForTest(fixture fixtures.FixtureTest, opts fixtures.RunOptions) (api.Spec, ai.AgentConfig, error) {
	resolved, config, err := resolveAIStepSpec(fixture, opts, &checklistResponse{})
	return resolved.Spec, config, err
}
