package database_test

import (
	"errors"
	"os"
	"path/filepath"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// retiredCaptainTriggers are the projection triggers earlier Gavel versions
// installed on Captain's tables, keyed by the table they sat on.
var retiredCaptainTriggers = map[string]string{
	"gavel_todo_prompt_run_projection":           "captain_prompt_runs",
	"gavel_todo_session_projection":              "captain_sessions",
	"gavel_todo_session_delete_projection":       "captain_sessions",
	"gavel_todo_turn_request_projection":         "captain_turn_requests",
	"gavel_todo_prompt_run_iteration_projection": "captain_prompt_run_iterations",
}

type deleteGuardRow struct {
	TargetTable string
	HostTable   string
	HostColumn  string
	Owner       string
}

var wantDeleteGuards = []deleteGuardRow{
	{TargetTable: "captain_plans", HostTable: "public.todo_issue_plans", HostColumn: "plan_id", Owner: "gavel"},
	{TargetTable: "captain_prompt_runs", HostTable: "public.todo_issue_prompt_runs", HostColumn: "prompt_run_id", Owner: "gavel"},
}

var _ = Describe("Gavel's hold on Captain's tables", Ordered, func() {
	var db *gorm.DB

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
		}
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_captain_hooks",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })
		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		db = migrateAndOpen()
	})

	It("installs no trigger on and no foreign key into a Captain table", func() {
		Expect(gavelTriggersOnCaptainTables(db)).To(BeEmpty())
		Expect(foreignKeysIntoCaptainTables(db)).To(BeEmpty())
	})

	It("registers Captain delete guards for both link columns, idempotently", func() {
		Expect(deleteGuards(db)).To(Equal(wantDeleteGuards))
		migrateAndOpen()
		Expect(deleteGuards(db)).To(Equal(wantDeleteGuards), "re-migrating must not duplicate a guard")
	})

	It("indexes both guarded link columns so a Captain delete probes an index", func() {
		Expect(leadingColumnIndexes(db, "todo_issue_prompt_runs", "prompt_run_id")).To(ContainElement("todo_issue_prompt_runs_prompt_run_id_key"))
		Expect(leadingColumnIndexes(db, "todo_issue_plans", "plan_id")).To(ContainElement("todo_issue_plans_plan_id_idx"))
	})

	It("refuses to delete a linked prompt run or plan until Gavel unlinks it", func() {
		link := newLinkedRunAndPlan(db)
		expectGuardViolation(db.Exec(`DELETE FROM captain_prompt_runs WHERE id = ?`, link.runID).Error)
		expectGuardViolation(db.Exec(`DELETE FROM captain_plans WHERE id = ?`, link.planID).Error)
		expectGuardViolation(db.Exec(`DELETE FROM captain_sessions WHERE id = ?`, link.sessionID).Error)

		Expect(db.Exec(`DELETE FROM todo_issue_plans WHERE plan_id = ?`, link.planID).Error).To(Succeed())
		Expect(db.Exec(`DELETE FROM captain_plans WHERE id = ?`, link.planID).Error).To(Succeed())
		Expect(db.Exec(`UPDATE todo_issues SET active_prompt_run_id = NULL WHERE id = ?`, link.issueID).Error).To(Succeed())
		Expect(db.Exec(`DELETE FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, link.runID).Error).To(Succeed())
		Expect(db.Exec(`DELETE FROM captain_prompt_runs WHERE id = ?`, link.runID).Error).To(Succeed())
	})

	It("removes the retired triggers and foreign keys from a database an older Gavel migrated", func() {
		installRetiredCaptainHooks(db)
		Expect(gavelTriggersOnCaptainTables(db)).To(HaveLen(len(retiredCaptainTriggers)))
		Expect(foreignKeysIntoCaptainTables(db)).To(HaveLen(2))
		Expect(db.Exec(`
			DELETE FROM schema_migration_scripts
			WHERE scope = 'gavel' AND path = '089_drop_captain_table_hooks.sql'`).Error).To(Succeed())

		migrateAndOpen()

		Expect(gavelTriggersOnCaptainTables(db)).To(BeEmpty())
		Expect(foreignKeysIntoCaptainTables(db)).To(BeEmpty())
		var triggerFunctions int64
		Expect(db.Raw(`
			SELECT count(*) FROM pg_proc
			WHERE proname LIKE 'gavel\_todo\_%\_projection\_trigger'`).Scan(&triggerFunctions).Error).To(Succeed())
		Expect(triggerFunctions).To(BeZero())
		var runtimeRows int64
		Expect(db.Raw(`SELECT count(*) FROM todo_issue_runtime`).Scan(&runtimeRows).Error).
			To(Succeed(), "090 re-ran and dropped the runtime view; 110 and 112 must have recreated it")
		Expect(deleteGuards(db)).To(Equal(wantDeleteGuards))
	})
})

func migrateAndOpen() *gorm.DB {
	GinkgoHelper()
	opened, err := database.Open(GinkgoT().Context(), database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
	return opened.Gorm()
}

func gavelTriggersOnCaptainTables(db *gorm.DB) []string {
	GinkgoHelper()
	var names []string
	Expect(db.Raw(`
		SELECT trg.tgname || ' ON ' || relation.relname
		FROM pg_trigger AS trg
		JOIN pg_class AS relation ON relation.oid = trg.tgrelid
		WHERE NOT trg.tgisinternal
		  AND relation.relname LIKE 'captain\_%'
		  AND trg.tgname LIKE 'gavel\_%'
		ORDER BY 1`).Scan(&names).Error).To(Succeed())
	return names
}

func foreignKeysIntoCaptainTables(db *gorm.DB) []string {
	GinkgoHelper()
	var names []string
	Expect(db.Raw(`
		SELECT con.conname
		FROM pg_constraint AS con
		JOIN pg_class AS host ON host.oid = con.conrelid
		JOIN pg_class AS target ON target.oid = con.confrelid
		WHERE con.contype = 'f'
		  AND target.relname LIKE 'captain\_%'
		  AND host.relname NOT LIKE 'captain\_%'
		ORDER BY 1`).Scan(&names).Error).To(Succeed())
	return names
}

func deleteGuards(db *gorm.DB) []deleteGuardRow {
	GinkgoHelper()
	var rows []deleteGuardRow
	Expect(db.Raw(`
		SELECT target_table, host_table, host_column, owner
		FROM captain_delete_guards
		ORDER BY target_table, host_table, host_column`).Scan(&rows).Error).To(Succeed())
	return rows
}

func leadingColumnIndexes(db *gorm.DB, table, column string) []string {
	GinkgoHelper()
	var names []string
	Expect(db.Raw(`
		SELECT index_class.relname
		FROM pg_index AS idx
		JOIN pg_class AS index_class ON index_class.oid = idx.indexrelid
		JOIN pg_attribute AS attribute
		  ON attribute.attrelid = idx.indrelid AND attribute.attnum = idx.indkey[0]
		WHERE idx.indrelid = ?::regclass AND attribute.attname = ?
		ORDER BY 1`, "public."+table, column).Scan(&names).Error).To(Succeed())
	return names
}

type linkedRunAndPlan struct {
	issueID, sessionID, runID, planID uuid.UUID
}

func newLinkedRunAndPlan(db *gorm.DB) linkedRunAndPlan {
	GinkgoHelper()
	link := linkedRunAndPlan{issueID: uuid.New(), sessionID: uuid.New(), runID: uuid.New(), planID: uuid.New()}
	workspaceID := uuid.New()
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO todo_workspaces (id, repo_key) VALUES (?, ?)`,
			[]any{workspaceID, "github.com/example/guard-" + workspaceID.String()}},
		{`INSERT INTO todo_issues (id, workspace_id, title) VALUES (?, ?, 'Guarded')`,
			[]any{link.issueID, workspaceID}},
		{`INSERT INTO captain_sessions (id, source, provider, host_id) VALUES (?, 'gavel', 'test', 'local')`,
			[]any{link.sessionID}},
		{`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin) VALUES (?, ?, ?, 'gavel-test')`,
			[]any{link.runID, link.sessionID, link.sessionID}},
		{`INSERT INTO captain_plans (id, source_session_id, source_prompt_run_id, title) VALUES (?, ?, ?, 'Guarded plan')`,
			[]any{link.planID, link.sessionID, link.runID}},
		{`INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal) VALUES (?, ?, 'run', 0)`,
			[]any{link.issueID, link.runID}},
		{`UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`, []any{link.runID, link.issueID}},
		{`INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal) VALUES (?, ?, 0)`,
			[]any{link.issueID, link.planID}},
	} {
		Expect(db.Exec(statement.sql, statement.args...).Error).To(Succeed(), statement.sql)
	}
	return link
}

func expectGuardViolation(err error) {
	GinkgoHelper()
	var pgErr *pgconn.PgError
	Expect(errors.As(err, &pgErr)).To(BeTrue(), "want a PostgreSQL error, got %v", err)
	Expect(pgErr.Code).To(Equal("23503"))
	Expect(pgErr.ConstraintName).To(Equal("captain_delete_guard"))
}

// installRetiredCaptainHooks recreates, by name, the triggers and foreign keys
// an older Gavel left on Captain's tables. The trigger function is a stand-in:
// only the objects' existence matters to the upgrade under test.
func installRetiredCaptainHooks(db *gorm.DB) {
	GinkgoHelper()
	Expect(db.Exec(`
		CREATE OR REPLACE FUNCTION public.gavel_todo_session_projection_trigger()
		RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RETURN NULL; END $$`).Error).To(Succeed())
	for trigger, table := range retiredCaptainTriggers {
		Expect(db.Exec(`CREATE TRIGGER ` + trigger + ` AFTER UPDATE ON public.` + table +
			` FOR EACH ROW EXECUTE FUNCTION public.gavel_todo_session_projection_trigger()`).Error).To(Succeed())
	}
	Expect(db.Exec(`
		ALTER TABLE public.todo_issue_prompt_runs
		  ADD CONSTRAINT todo_issue_prompt_runs_captain_prompt_run_fkey
		  FOREIGN KEY (prompt_run_id) REFERENCES public.captain_prompt_runs (id) ON DELETE RESTRICT`).Error).To(Succeed())
	Expect(db.Exec(`
		ALTER TABLE public.todo_issue_plans
		  ADD CONSTRAINT todo_issue_plans_captain_plan_fkey
		  FOREIGN KEY (plan_id) REFERENCES public.captain_plans (id) ON DELETE RESTRICT`).Error).To(Succeed())
}
