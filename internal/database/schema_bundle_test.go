package database

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaBundleIncludesOrderedCaptainProjectionSQL(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(schemaFS, "schema")
	require.NoError(t, err)
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	assert.Contains(t, names, "090_prepare_runtime_state.sql")
	assert.Contains(t, names, "100_todo_captain_constraints.sql")
	assert.Contains(t, names, "102_todo_prompt_run_step_kind.sql")
	assert.Contains(t, names, "105_view_todo_plan_revisions.sql")
	assert.Contains(t, names, "110_todo_projection_functions.sql")
	assert.Contains(t, names, "111_todo_projection_triggers.sql")
	assert.Contains(t, names, "112_view_todo_issue_runtime.sql")
	assert.Contains(t, names, "115_backfill_todo_activity.sql")
	assert.Contains(t, names, "116_backfill_triage_step_kind.sql")
	assert.Contains(t, names, "120_backfill_todo_labels.sql")
	assert.Contains(t, names, "130_task_history_storage_params.sql")

	prepareSQL := readEmbeddedSchemaFile(t, "schema/090_prepare_runtime_state.sql")
	assert.Contains(t, prepareSQL, "-- phase: pre")
	assert.Contains(t, prepareSQL, "DROP VIEW IF EXISTS public.todo_issue_runtime")

	// Hash-gated run-once scripts keep steady-state applies free of DDL; views
	// are restored via commons-db view-dependency invalidation, so no script may
	// opt back into re-running on every apply.
	for _, name := range names {
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		assert.NotContains(t, readEmbeddedSchemaFile(t, "schema/"+name), "-- runs: always",
			"%s must be a hash-gated run-once script", name)
	}

	constraintSQL := readEmbeddedSchemaFile(t, "schema/100_todo_captain_constraints.sql")
	assert.Contains(t, constraintSQL, "-- phase: post")
	assert.Contains(t, constraintSQL, "todo_issue_prompt_runs_captain_prompt_run_fkey")
	assert.Contains(t, constraintSQL, "todo_issue_plans_captain_plan_fkey")
	assert.NotContains(t, constraintSQL, "todo_issue_plan_revision_details",
		"the plan-revisions view moved to its own dependency-scoped file")

	revisionsViewSQL := readEmbeddedSchemaFile(t, "schema/105_view_todo_plan_revisions.sql")
	assert.Contains(t, revisionsViewSQL, "-- phase: post")
	assert.Contains(t, revisionsViewSQL, "CREATE OR REPLACE VIEW public.todo_issue_plan_revision_details")

	functionsSQL := readEmbeddedSchemaFile(t, "schema/110_todo_projection_functions.sql")
	assert.Contains(t, functionsSQL, "-- phase: post")
	assert.Contains(t, functionsSQL, "gavel_project_todo_prompt_run")
	assert.Contains(t, functionsSQL, "gavel_todo_issue_execution_state")
	assert.Contains(t, functionsSQL, "GREATEST(updated_at, p_activity_at)")
	assert.Contains(t, functionsSQL, "{workflow,autoVerifyWithoutFixture}")
	assert.GreaterOrEqual(t, strings.Count(functionsSQL, "SET search_path = pg_catalog, public"), 7)
	assert.NotContains(t, functionsSQL, "CREATE OR REPLACE TRIGGER",
		"projection triggers moved to their own file")
	assert.NotContains(t, functionsSQL, "CREATE OR REPLACE VIEW",
		"the runtime view moved to its own dependency-scoped file")

	triggersSQL := readEmbeddedSchemaFile(t, "schema/111_todo_projection_triggers.sql")
	assert.Contains(t, triggersSQL, "-- dependsOn: 110_todo_projection_functions.sql")
	assert.Contains(t, triggersSQL, "gavel_todo_turn_request_projection")
	assert.Contains(t, triggersSQL, "gavel_todo_prompt_run_iteration_projection")

	runtimeViewSQL := readEmbeddedSchemaFile(t, "schema/112_view_todo_issue_runtime.sql")
	assert.Contains(t, runtimeViewSQL, "-- dependsOn: 110_todo_projection_functions.sql")
	assert.Contains(t, runtimeViewSQL, "CREATE OR REPLACE VIEW public.todo_issue_runtime")

	backfillSQL := readEmbeddedSchemaFile(t, "schema/115_backfill_todo_activity.sql")
	assert.Contains(t, backfillSQL, "-- dependsOn: 111_todo_projection_triggers.sql")
	assert.Contains(t, backfillSQL, "GREATEST(issue.updated_at, activity.activity_at)")

	// Atlas leaves a renamed-nothing CHECK alone even when its expression
	// changed, so the declared step_kind domain only reaches an existing
	// database through 102. The two spellings must stay identical, and the
	// triage backfill must not run before the widened CHECK lands.
	stepKindSQL := readEmbeddedSchemaFile(t, "schema/102_todo_prompt_run_step_kind.sql")
	assert.Contains(t, stepKindSQL, "-- phase: post")
	assert.Contains(t, stepKindSQL, declaredCheckExpr(t, "todo_issue_prompt_runs_step_kind_check"),
		"102 must reconcile the CHECK exactly as todos.hcl declares it")

	triageSQL := readEmbeddedSchemaFile(t, "schema/116_backfill_triage_step_kind.sql")
	assert.Contains(t, triageSQL, "102_todo_prompt_run_step_kind.sql")

	taskHistorySchema := readEmbeddedSchemaFile(t, "schema/task_history.hcl")
	assert.Contains(t, taskHistorySchema, `table "task_run_history"`)
	assert.Contains(t, taskHistorySchema, `type = jsonb`)
	assert.Contains(t, taskHistorySchema, `index "idx_task_run_history_expires_at"`)

	// Atlas OSS cannot express storage parameters, so the update-in-place churn
	// settings for task_run_history live in post-phase SQL rather than the HCL.
	storageParamsSQL := readEmbeddedSchemaFile(t, "schema/130_task_history_storage_params.sql")
	assert.Contains(t, storageParamsSQL, "-- phase: post")
	assert.Contains(t, storageParamsSQL, "ALTER TABLE public.task_run_history SET (")
	assert.Contains(t, storageParamsSQL, "fillfactor = 70")
	assert.NotContains(t, taskHistorySchema, "fillfactor")
}

// declaredCheckExpr returns the expr of the named check block in todos.hcl, so
// a reconciliation script can be compared against the single declaration rather
// than against a second hand-copied literal.
func declaredCheckExpr(t *testing.T, constraint string) string {
	t.Helper()
	lines := strings.Split(readEmbeddedSchemaFile(t, "schema/todos.hcl"), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != `check "`+constraint+`" {` {
			continue
		}
		require.Less(t, i+1, len(lines), "check %q declares no expr", constraint)
		expr := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i+1]), "expr ="))
		expr, err := strconv.Unquote(expr)
		require.NoError(t, err, "check %q has an unreadable expr", constraint)
		return expr
	}
	t.Fatalf("todos.hcl declares no check %q", constraint)
	return ""
}

func readEmbeddedSchemaFile(t *testing.T, name string) string {
	t.Helper()
	data, err := fs.ReadFile(schemaFS, name)
	require.NoError(t, err)
	return string(data)
}
