package runtime

import (
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The column used to hold a two-key map built from the issue's fixture string,
// so nothing could answer "what actually ran". It now holds the dispatched spec.
func TestRenderedSpecPersistsTheExecutedRuntime(t *testing.T) {
	spec := api.Spec{
		Model:       api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Effort: api.EffortHigh},
		Budget:      api.Budget{Timeout: "45m", MaxTurns: 12},
		Permissions: api.Permissions{Mode: api.PermissionPlan},
	}

	rendered, err := renderedSpec(spec)
	require.NoError(t, err)

	assert.Equal(t, "claude-sonnet-5", rendered["model"])
	assert.Equal(t, string(api.ModeAgent), rendered["mode"])
	assert.Equal(t, string(api.EffortHigh), rendered["effort"])
	assert.Equal(t, "45m", rendered["budget"].(map[string]any)["timeout"])
	assert.Equal(t, float64(12), rendered["budget"].(map[string]any)["maxTurns"])
	assert.Equal(t, string(api.PermissionPlan), rendered["permissions"].(map[string]any)["mode"])
}

func TestRenderedSpecKeepsOnlyDeclaredVerification(t *testing.T) {
	rendered, err := renderedSpec(api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 3}}})
	require.NoError(t, err)

	verify := rendered["workflow"].(map[string]any)["verify"].(map[string]any)
	assert.NotContains(t, verify, "fixture")
	assert.Equal(t, float64(3), verify["maxIterations"])

	bare, err := renderedSpec(api.Spec{})
	require.NoError(t, err)
	assert.NotContains(t, bare, "workflow")
}

func TestRenderedSpecKeepsAFixtureTheDispatchedSpecDeclares(t *testing.T) {
	const dispatched = "```bash\necho dispatched\n```"
	spec := api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{Fixture: dispatched, MaxIterations: 2}}}

	rendered, err := renderedSpec(spec)
	require.NoError(t, err)

	verify := rendered["workflow"].(map[string]any)["verify"].(map[string]any)
	assert.Equal(t, dispatched, verify["fixture"])
	assert.Equal(t, float64(2), verify["maxIterations"])
}

// H9: only a spec whose checkout has already been consumed is safe to replay.
// A persisted spec that still said "check this out" would clone a second tree
// on every continuation.
func TestRenderedSpecCarriesThePreparedTreeNotTheCheckoutRequest(t *testing.T) {
	prepared := api.Spec{Setup: &shell.Setup{Cwd: "/work/.worktrees/todo-1"}}

	rendered, err := renderedSpec(prepared)
	require.NoError(t, err)

	setup := rendered["setup"].(map[string]any)
	assert.Equal(t, "/work/.worktrees/todo-1", setup["cwd"])
	assert.NotContains(t, setup, "checkout")
	// Setup.Env is json:"-" — a shell.Prepare output, re-derived on the next run.
	assert.NotContains(t, setup, "env")
}
