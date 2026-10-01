-- phase: pre

-- Gavel installs no DDL on Captain's tables. It used to own five projection
-- triggers on them and two ON DELETE RESTRICT foreign keys onto them; Captain
-- now provides both hooks itself: the captain_row_change NOTIFY channel that
-- Gavel's projection LISTENs on, and delete guards that Gavel registers through
-- captaindb.RegisterDeleteGuard after migrating. This one-time script removes
-- what earlier Gavel versions installed on an existing database. It is a pre
-- script ahead of 090, which drops the trigger functions these triggers use.
DROP TRIGGER IF EXISTS gavel_todo_prompt_run_projection ON public.captain_prompt_runs;
DROP TRIGGER IF EXISTS gavel_todo_session_projection ON public.captain_sessions;
DROP TRIGGER IF EXISTS gavel_todo_session_delete_projection ON public.captain_sessions;
DROP TRIGGER IF EXISTS gavel_todo_turn_request_projection ON public.captain_turn_requests;
DROP TRIGGER IF EXISTS gavel_todo_prompt_run_iteration_projection ON public.captain_prompt_run_iterations;

DROP FUNCTION IF EXISTS public.gavel_todo_iteration_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_request_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_session_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_prompt_run_projection_trigger();

-- The link tables do not exist yet on a fresh database: this runs before Atlas.
ALTER TABLE IF EXISTS public.todo_issue_prompt_runs
  DROP CONSTRAINT IF EXISTS todo_issue_prompt_runs_captain_prompt_run_fkey;
ALTER TABLE IF EXISTS public.todo_issue_plans
  DROP CONSTRAINT IF EXISTS todo_issue_plans_captain_plan_fkey;
