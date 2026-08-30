-- phase: post

-- Reconcile todo_issue_prompt_runs.step_kind's CHECK with its declaration in
-- todos.hcl.
--
-- Atlas never applies a CHECK whose expression changed while its name stayed
-- the same: sqlx.checksSimilarDiff pairs the live constraint with the declared
-- one by name and then compares only PostgreSQL's NO INHERIT flag, so no
-- ModifyCheck is ever planned (ariga.io/atlas sql/internal/sqlx/diff.go). A new
-- database gets the declared expression for free because the table is CREATEd
-- from the HCL; a database that already has the table silently keeps the old
-- one. Widening the domain — 'triage' earning its own step kind — is therefore
-- invisible on every existing database until it is applied here, and the
-- backfill in 116 fails with 23514 when it is not.
--
-- Any later change to that check block in todos.hcl has to be mirrored below;
-- schema_bundle_test.go fails when the two drift.
ALTER TABLE public.todo_issue_prompt_runs
  DROP CONSTRAINT IF EXISTS todo_issue_prompt_runs_step_kind_check;

ALTER TABLE public.todo_issue_prompt_runs
  ADD CONSTRAINT todo_issue_prompt_runs_step_kind_check
  CHECK (step_kind = ANY (ARRAY['plan'::text, 'run'::text, 'verify'::text, 'triage'::text]));
