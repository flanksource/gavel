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
	const fixture = "```bash\necho ok\n```"
	spec := api.Spec{
		Model:       api.Model{Name: "claude-code-sonnet", Mode: api.ModeAgent, Effort: api.EffortHigh},
		Budget:      api.Budget{Timeout: "45m", MaxTurns: 12},
		Permissions: api.Permissions{Mode: api.PermissionPlan},
	}

	rendered, err := renderedSpec(spec, fixture)
	require.NoError(t, err)

	assert.Equal(t, "claude-code-sonnet", rendered["model"])
	assert.Equal(t, string(api.ModeAgent), rendered["backend"])
	assert.Equal(t, string(api.EffortHigh), rendered["effort"])
	assert.Equal(t, "45m", rendered["budget"].(map[string]any)["timeout"])
	assert.Equal(t, float64(12), rendered["budget"].(map[string]any)["maxTurns"])
	assert.Equal(t, string(api.PermissionPlan), rendered["permissions"].(map[string]any)["mode"])
}

// 110_todo_projection_functions.sql reads {workflow,verify,fixture}'s sibling
// {workflow,autoVerifyWithoutFixture}; the fixture itself is what a later
// verification replays, so the durable record must name the definition of done
// the run was dispatched against.
func TestRenderedSpecStampsTheIssueFixtureOntoTheWorkflow(t *testing.T) {
	const fixture = "```bash\necho ok\n```"

	rendered, err := renderedSpec(api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: 3}}}, fixture)
	require.NoError(t, err)

	verify := rendered["workflow"].(map[string]any)["verify"].(map[string]any)
	assert.Equal(t, fixture, verify["fixture"])
	assert.Equal(t, float64(3), verify["maxIterations"], "stamping the fixture must not erase the run's verify settings")

	bare, err := renderedSpec(api.Spec{}, "   ")
	require.NoError(t, err)
	assert.NotContains(t, bare, "workflow", "an issue with no fixture declares no workflow verification")
}

// Stamping happens on a copy: the dispatched spec is live on the runner while
// this projection is built, and a shared Workflow pointer would mutate it.
func TestRenderedSpecDoesNotMutateTheDispatchedSpec(t *testing.T) {
	workflow := &api.Workflow{Verify: &api.Verify{MaxIterations: 2}}
	spec := api.Spec{Workflow: workflow}

	_, err := renderedSpec(spec, "```bash\ntrue\n```")
	require.NoError(t, err)

	assert.Empty(t, workflow.Verify.Fixture)
}

// H9: only a spec whose checkout has already been consumed is safe to replay.
// A persisted spec that still said "check this out" would clone a second tree
// on every continuation.
func TestRenderedSpecCarriesThePreparedTreeNotTheCheckoutRequest(t *testing.T) {
	prepared := api.Spec{Setup: &shell.Setup{Cwd: "/work/.worktrees/todo-1"}}

	rendered, err := renderedSpec(prepared, "")
	require.NoError(t, err)

	setup := rendered["setup"].(map[string]any)
	assert.Equal(t, "/work/.worktrees/todo-1", setup["cwd"])
	assert.NotContains(t, setup, "checkout")
	// Setup.Env is json:"-" — a shell.Prepare output, re-derived on the next run.
	assert.NotContains(t, setup, "env")
}
