package types

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
)

func ResolveAIStepSpecForTest(fixture fixtures.FixtureTest, opts fixtures.RunOptions) (api.Spec, ai.AgentConfig) {
	return resolveAIStepSpec(fixture, opts, &checklistResponse{})
}
