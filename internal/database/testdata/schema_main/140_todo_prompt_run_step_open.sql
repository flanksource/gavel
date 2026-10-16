-- phase: post
-- dependsOn: 116_backfill_triage_step_kind.sql

-- Open todo_issue_prompt_runs.step_kind from a closed enum to a shape check.
--
-- Lifecycle steps are data, not code: todos/lifecycle/todos.yaml declares the
-- built-in steps by name and a project may declare its own. A closed enum makes
-- the database the authority on that vocabulary, so the first project to add a
-- step of its own gets a 23514 from a constraint that knows nothing about its
-- lifecycle. Only the shape of the name is a database concern — non-empty, and
-- already lower-cased and trimmed so a link row cannot be recorded under a
-- spelling readers will not match. Which names are legal is decided by the
-- lifecycle definition and enforced above the database.
--
-- 102 and the 116 backfill still run first: 102 admits 'triage' so 116 can
-- write it, and this script then drops the enum entirely. Ordering is explicit
-- through dependsOn because the backfill's UPDATE is rejected by neither
-- expression but its *source* rows must exist before the domain widens.
--
-- Same Atlas caveat 102 documents: sqlx.checksSimilarDiff pairs the live
-- constraint with the declared one by name and then compares only PostgreSQL's
-- NO INHERIT flag, so a CHECK whose expression changed under an unchanged name
-- never produces a ModifyCheck (ariga.io/atlas sql/internal/sqlx/diff.go). A
-- fresh database gets the declared expression because the table is CREATEd from
-- todos.hcl; an existing database silently keeps the old one until this script
-- applies it. Any later change to that check block in todos.hcl has to be
-- mirrored below; schema_bundle_test.go fails when the two drift.
ALTER TABLE public.todo_issue_prompt_runs
  DROP CONSTRAINT IF EXISTS todo_issue_prompt_runs_step_kind_check;

ALTER TABLE public.todo_issue_prompt_runs
  ADD CONSTRAINT todo_issue_prompt_runs_step_kind_check
  CHECK (step_kind <> '' AND step_kind = lower(btrim(step_kind)));
