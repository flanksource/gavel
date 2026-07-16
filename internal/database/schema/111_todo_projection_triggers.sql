-- phase: post
-- dependsOn: 110_todo_projection_functions.sql

CREATE OR REPLACE TRIGGER gavel_todo_prompt_run_projection
AFTER INSERT OR UPDATE OF phase, state, verification_markdown, rendered_spec,
  session_id, root_session_id, version, updated_at
ON public.captain_prompt_runs FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_prompt_run_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_session_projection
AFTER INSERT OR UPDATE OF provider_session_id, parent_session_id,
  root_session_id, lifecycle_status, activity_state, health_state,
  state_version, state_observed_at, started_at, ended_at, last_activity_at, updated_at
ON public.captain_sessions FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_session_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_session_delete_projection
BEFORE DELETE ON public.captain_sessions FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_session_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_turn_request_projection
AFTER INSERT OR DELETE OR UPDATE OF prompt_run_id, session_id, kind, state, version, resolved_at
ON public.captain_turn_requests FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_request_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_prompt_run_iteration_projection
AFTER INSERT OR DELETE OR UPDATE OF prompt_run_id, iteration, state,
  verification_result, updated_at
ON public.captain_prompt_run_iterations FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_iteration_projection_trigger();
