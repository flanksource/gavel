package database

import (
	"io/fs"
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
	assert.Contains(t, names, "100_todo_captain_constraints.sql")
	assert.Contains(t, names, "110_todo_captain_projection.sql")

	constraintSQL := readEmbeddedSchemaFile(t, "schema/100_todo_captain_constraints.sql")
	assert.Contains(t, constraintSQL, "-- phase: post")
	assert.Contains(t, constraintSQL, "-- runs: always")
	assert.Contains(t, constraintSQL, "todo_issue_prompt_runs_captain_prompt_run_fkey")
	assert.Contains(t, constraintSQL, "todo_issue_plans_captain_plan_fkey")
	assert.Contains(t, constraintSQL, "todo_issue_plan_revision_details")

	projectionSQL := readEmbeddedSchemaFile(t, "schema/110_todo_captain_projection.sql")
	assert.Contains(t, projectionSQL, "-- dependsOn: 100_todo_captain_constraints.sql")
	assert.Contains(t, projectionSQL, "-- runs: always")
	assert.Contains(t, projectionSQL, "gavel_project_todo_prompt_run")
	assert.Contains(t, projectionSQL, "gavel_todo_turn_request_projection")
	assert.Contains(t, projectionSQL, "gavel_todo_prompt_run_iteration_projection")
	assert.Contains(t, projectionSQL, "{workflow,autoVerifyWithoutFixture}")
	assert.GreaterOrEqual(t, strings.Count(projectionSQL, "SET search_path = pg_catalog, public"), 7)
}

func readEmbeddedSchemaFile(t *testing.T, name string) string {
	t.Helper()
	data, err := fs.ReadFile(schemaFS, name)
	require.NoError(t, err)
	return string(data)
}
