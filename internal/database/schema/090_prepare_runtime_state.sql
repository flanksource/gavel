-- phase: pre
-- dependsOn: 089_drop_captain_table_hooks.sql

-- The prior projection function has a hard dependency on
-- todo_issues.execution_state. Remove the SQL-owned projection objects before
-- Atlas drops that cached column; the post-schema bundle recreates them as
-- read-time state and activity functions. 089 has already removed the triggers
-- on Captain's tables that used the trigger functions dropped here.
DROP VIEW IF EXISTS public.todo_issue_runtime;
DROP FUNCTION IF EXISTS public.gavel_todo_iteration_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_request_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_session_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_todo_prompt_run_projection_trigger();
DROP FUNCTION IF EXISTS public.gavel_project_todo_session(uuid);
DROP FUNCTION IF EXISTS public.gavel_project_todo_prompt_run(uuid);
DROP FUNCTION IF EXISTS public.gavel_touch_todo_prompt_run(uuid, timestamptz);
DROP FUNCTION IF EXISTS public.gavel_project_todo_issue(uuid);
DROP FUNCTION IF EXISTS public.gavel_touch_todo_issue(uuid, timestamptz);
DROP FUNCTION IF EXISTS public.gavel_todo_issue_execution_state(uuid);
