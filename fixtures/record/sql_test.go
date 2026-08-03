package record

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyNamesTheOpAndTheRelations(t *testing.T) {
	cases := []struct {
		name   string
		sql    string
		op     string
		tables []string
	}{
		{"select", "SELECT id FROM users WHERE id = $1", "SELECT", []string{"users"}},
		{"join", "SELECT * FROM orders o JOIN users u ON u.id = o.user_id", "SELECT", []string{"orders", "users"}},
		{"insert", `INSERT INTO "todo_issues" (id) VALUES ($1)`, "INSERT", []string{"todo_issues"}},
		{"update", "UPDATE runs SET status = $1", "UPDATE", []string{"runs"}},
		{"ddl", "CREATE TABLE widgets (id int)", "CREATE", []string{"widgets"}},
		{"lowercase is normalised", "select 1 from dual", "SELECT", []string{"dual"}},
		{"schema qualified", "SELECT 1 FROM public.users", "SELECT", []string{"public.users"}},
		// A sub-select's `FROM (` has no identifier to name, and must not
		// contribute a junk table rather than simply none.
		{"subselect names nothing", "SELECT * FROM (SELECT 1) t", "SELECT", nil},
		{"no relations", "BEGIN", "BEGIN", nil},
		{"empty", "   ", "", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op, tables := classify(tc.sql)
			assert.Equal(t, tc.op, op)
			assert.Equal(t, tc.tables, tables)
		})
	}
}

func TestRowsFromTagDistinguishesNoneFromZero(t *testing.T) {
	cases := []struct {
		tag  string
		rows int
	}{
		{"INSERT 0 1", 1},
		{"SELECT 3", 3},
		{"UPDATE 0", 0},
		{"DELETE 12", 12},
		// No count at all is -1, so a fixture can tell "matched nothing" from
		// "not a row-returning statement".
		{"BEGIN", -1},
		{"SET", -1},
		{"", -1},
	}

	for _, tc := range cases {
		t.Run(tc.tag, func(t *testing.T) {
			assert.Equal(t, tc.rows, rowsFromTag(tc.tag))
		})
	}
}

func TestSQLCELVarsAggregatesExactlyAndCapsDetail(t *testing.T) {
	statements := []Statement{
		{SQL: "BEGIN", Op: "BEGIN", Rows: -1, DurationMs: 1},
		{SQL: "INSERT INTO users (id) VALUES ($1)", Op: "INSERT", Tables: []string{"users"}, Rows: 1, DurationMs: 4},
		{SQL: "SELECT * FROM users", Op: "SELECT", Tables: []string{"users"}, Rows: 1, DurationMs: 2},
		{SQL: "SELECT * FROM missing", Op: "SELECT", Tables: []string{"missing"}, Rows: -1, DurationMs: 3,
			Error: "ERROR: relation \"missing\" does not exist"},
	}

	vars := SQLCELVars(statements, "/tmp/rec.sql.jsonl")

	assert.Equal(t, 4, vars["statements"])
	assert.Equal(t, 1, vars["errors"])
	assert.Equal(t, int64(10), vars["duration_ms"])
	assert.Equal(t, map[string]int{"BEGIN": 1, "INSERT": 1, "SELECT": 2}, vars["by_op"])
	assert.Equal(t, []string{"missing", "users"}, vars["tables"])
	assert.Equal(t, "/tmp/rec.sql.jsonl", vars["path"])
	assert.Len(t, vars["queries"], 4)
}

func TestSQLCELVarsKeepsEveryKeyOnAnEmptyCapture(t *testing.T) {
	vars := SQLCELVars(nil, "/tmp/rec.sql.jsonl")

	// `sql.errors == 0` has to be a legal assertion for a fixture that issued
	// no statements, so no key may be conditional on there being data.
	for _, key := range []string{"statements", "errors", "duration_ms", "by_op", "tables", "queries", "path"} {
		assert.Contains(t, vars, key)
	}
	assert.Equal(t, 0, vars["statements"])
	assert.Empty(t, vars["queries"])
}

func TestSQLCELVarsCapsPerQueryDetail(t *testing.T) {
	statements := make([]Statement, celQueryCap+50)
	for i := range statements {
		statements[i] = Statement{SQL: "SELECT 1", Op: "SELECT", Rows: 1, DurationMs: 1}
	}

	vars := SQLCELVars(statements, "")

	assert.Equal(t, celQueryCap+50, vars["statements"], "counts stay exact past the cap")
	assert.Len(t, vars["queries"], celQueryCap)
}

func TestWriteStatementsEmitsOneObjectPerLine(t *testing.T) {
	started := time.Date(2026, 8, 2, 10, 30, 0, 0, time.UTC)
	statements := []Statement{
		{started: started, SQL: "BEGIN", Op: "BEGIN", Rows: -1},
		{started: started.Add(time.Second), SQL: "COMMIT", Op: "COMMIT", Rows: -1},
	}

	var out strings.Builder
	require.NoError(t, WriteStatements(&out, statements))

	lines := strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], `"sql":"BEGIN"`)
	assert.Contains(t, lines[0], `"started_at":`, "the unexported start time is serialised for the artifact")
	assert.Contains(t, lines[1], `"sql":"COMMIT"`)
}

func TestStatementsBetweenNarrowsToTheWindow(t *testing.T) {
	base := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	statements := []Statement{
		{started: base, SQL: "before"},
		{started: base.Add(2 * time.Second), SQL: "inside"},
		{started: base.Add(10 * time.Second), SQL: "after"},
	}

	window := StatementsBetween(statements, base.Add(time.Second), base.Add(5*time.Second))

	require.Len(t, window, 1)
	assert.Equal(t, "inside", window[0].SQL)
}
