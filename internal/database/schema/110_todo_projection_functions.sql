-- phase: post

-- The projection used to expose gavel_project_todo_issue, a second writer of
-- durable status that was later stubbed to `RETURN false`. Its callers all
-- discarded the result, so it is dropped rather than kept as a contract nobody
-- depends on. Activity propagation and read-time state are what remain.
DROP FUNCTION IF EXISTS public.gavel_project_todo_issue(uuid);

-- Derive transient execution state from Captain at read time. The source='gavel'
-- session remains only the prompt-run admission root; monitored Claude/Codex
-- sessions with the same provider identity own agent lifecycle and health.
--
-- Nothing here reads a step's NAME. Steps are project-defined lifecycle data,
-- so any literal the projection compared against would be one a project may
-- not use; a run is classified by what it was asked to do instead. A planning
-- pass is a plan-mode spec. A verification pass is a spec that declares a
-- definition of done and no prompt — the shape the lifecycle's verify-only
-- step dispatches — or a run whose phase captain moved to verify, or one whose
-- latest iteration recorded a verifier verdict.
CREATE OR REPLACE FUNCTION public.gavel_todo_issue_execution_state(p_issue_id uuid)
RETURNS text
LANGUAGE sql
STABLE
SET search_path = pg_catalog, public
AS $$
WITH active AS (
  SELECT
    issue.active_prompt_run_id,
    run.state::text AS prompt_state,
    run.phase::text AS prompt_phase,
    run.rendered_spec,
    run.root_session_id,
    root.provider_session_id,
    run.rendered_spec #> '{workflow,verify}' IS NOT NULL
      AND COALESCE(run.rendered_spec #>> '{prompt,user}', '') = '' AS verify_only
  FROM public.todo_issues issue
  LEFT JOIN public.todo_issue_prompt_runs link
    ON link.issue_id = issue.id
   AND link.prompt_run_id = issue.active_prompt_run_id
  LEFT JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
  LEFT JOIN public.captain_sessions root ON root.id = run.root_session_id
  WHERE issue.id = p_issue_id
), agent_roots AS (
  SELECT session.id
  FROM active
  JOIN public.captain_sessions session
    ON active.provider_session_id IS NOT NULL
   AND session.provider_session_id = active.provider_session_id
   AND session.source IN ('claude', 'codex')
), session_tree AS (
  SELECT session.id, session.lifecycle_status::text AS lifecycle_status,
         session.activity_state::text AS activity_state,
         session.health_state::text AS health_state
  FROM public.captain_sessions session
  WHERE session.id IN (SELECT id FROM agent_roots)
     OR session.root_session_id IN (SELECT id FROM agent_roots)
), signals AS (
  SELECT
    EXISTS (SELECT 1 FROM session_tree WHERE health_state = 'zombie' OR lifecycle_status = 'failed') AS failed,
    EXISTS (SELECT 1 FROM session_tree WHERE health_state = 'stalled') AS stalled,
    EXISTS (SELECT 1 FROM session_tree WHERE activity_state IN ('ask', 'approval')) AS waiting,
    EXISTS (SELECT 1 FROM session_tree WHERE lifecycle_status IN ('succeeded', 'cancelled')) AS terminal
), latest_iteration AS (
  SELECT COALESCE((
    SELECT iteration.state::text = 'failed' AND iteration.verification_result IS NOT NULL
    FROM active
    JOIN public.captain_prompt_run_iterations iteration
      ON iteration.prompt_run_id = active.active_prompt_run_id
    ORDER BY iteration.iteration DESC, iteration.created_at DESC, iteration.id DESC
    LIMIT 1
  ), false) AS verification_failed
), pending AS (
  SELECT EXISTS (
    SELECT 1
    FROM active
    JOIN public.captain_turn_requests request
      ON request.state::text = 'pending'
     AND (
       request.prompt_run_id = active.active_prompt_run_id
       OR request.session_id IN (SELECT id FROM session_tree)
     )
  ) AS waiting
)
SELECT CASE
  WHEN active.active_prompt_run_id IS NULL THEN 'idle'
  WHEN active.prompt_state = 'cancelled' THEN 'idle'
  WHEN active.prompt_state = 'failed' THEN
    CASE WHEN active.verify_only OR active.prompt_phase = 'verify' OR latest_iteration.verification_failed
         THEN 'verification_failed' ELSE 'failed' END
  WHEN signals.failed THEN 'failed'
  WHEN signals.stalled THEN 'stalled'
  WHEN active.prompt_state = 'waiting' OR signals.waiting OR pending.waiting THEN 'waiting'
  WHEN signals.terminal THEN 'idle'
  WHEN active.prompt_state = 'succeeded' AND active.prompt_phase = 'finished' THEN 'idle'
  WHEN active.verify_only OR active.prompt_phase = 'verify' THEN 'verifying'
  WHEN active.rendered_spec #>> '{permissions,mode}' = 'plan' THEN 'planning'
  ELSE 'running'
END
FROM active
CROSS JOIN signals
CROSS JOIN latest_iteration
CROSS JOIN pending
$$;

-- Timestamp propagation is intentionally not an issue mutation: it neither
-- advances optimistic-lock version nor emits an audit event.
CREATE OR REPLACE FUNCTION public.gavel_touch_todo_issue(
  p_issue_id uuid,
  p_activity_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  affected integer := 0;
BEGIN
  IF p_issue_id IS NULL OR p_activity_at IS NULL THEN
    RETURN false;
  END IF;
  UPDATE public.todo_issues
     SET updated_at = GREATEST(updated_at, p_activity_at)
   WHERE id = p_issue_id
     AND updated_at < p_activity_at;
  GET DIAGNOSTICS affected = ROW_COUNT;
  RETURN affected > 0;
END
$$;

-- Durable workflow status has exactly one writer: the Go lifecycle host's
-- OnOutcome, which records a lifecycle_outcome event alongside the status it
-- decided. The projection used to be a second writer, re-deriving status from
-- the active prompt run's terminal shape. Two writers over one column cannot
-- agree once the lifecycle is data — the projection only ever saw the hard-coded
-- 'run'/'verify' vocabulary, so a project's own step read as an unverifiable
-- run and reopened an issue the host had just verified.
--
-- What a prompt run's change projects onto its issue is therefore its activity
-- alone: the watermark advances, nothing durable moves, and transient
-- execution state is derived at read time by gavel_todo_issue_execution_state.
-- The result counts the issues whose watermark advanced.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_prompt_run(p_prompt_run_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  linked record;
  changed_count integer := 0;
BEGIN
  FOR linked IN
    SELECT issue.id AS issue_id, run.updated_at AS activity_at
    FROM public.todo_issues issue
    JOIN public.todo_issue_prompt_runs link
      ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
    JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
    WHERE link.prompt_run_id = p_prompt_run_id
  LOOP
    IF public.gavel_touch_todo_issue(linked.issue_id, linked.activity_at) THEN
      changed_count := changed_count + 1;
    END IF;
  END LOOP;
  RETURN changed_count;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_touch_todo_prompt_run(
  p_prompt_run_id uuid,
  p_activity_at timestamptz
)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  linked record;
  changed_count integer := 0;
BEGIN
  IF p_prompt_run_id IS NULL OR p_activity_at IS NULL THEN
    RETURN 0;
  END IF;
  FOR linked IN
    SELECT issue.id AS issue_id
    FROM public.todo_issues issue
    JOIN public.todo_issue_prompt_runs link
      ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
    WHERE link.prompt_run_id = p_prompt_run_id
  LOOP
    IF public.gavel_touch_todo_issue(linked.issue_id, p_activity_at) THEN
      changed_count := changed_count + 1;
    END IF;
  END LOOP;
  RETURN changed_count;
END
$$;

-- Map any monitored session or descendant back through its root provider
-- identity to the admission root and active Gavel issue.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_session(p_session_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  linked record;
  changed_count integer := 0;
BEGIN
  FOR linked IN
    WITH target AS (
      SELECT session.*,
             COALESCE(session.root_session_id, session.id) AS family_root_id
      FROM public.captain_sessions session WHERE session.id = p_session_id
    ), identity AS (
      SELECT COALESCE(root.provider_session_id, target.provider_session_id) AS provider_session_id,
             GREATEST(
               COALESCE(target.last_activity_at, '-infinity'::timestamptz),
               COALESCE(target.state_observed_at, '-infinity'::timestamptz),
               COALESCE(target.started_at, '-infinity'::timestamptz),
               COALESCE(target.ended_at, '-infinity'::timestamptz),
               target.updated_at
             ) AS activity_at
      FROM target
      LEFT JOIN public.captain_sessions root ON root.id = target.family_root_id
    )
    SELECT DISTINCT issue.id AS issue_id, identity.activity_at
    FROM identity
    JOIN public.captain_sessions admission
      ON identity.provider_session_id IS NOT NULL
     AND admission.provider_session_id = identity.provider_session_id
     AND admission.source = 'gavel'
    JOIN public.captain_prompt_runs run ON run.root_session_id = admission.id
    JOIN public.todo_issue_prompt_runs link ON link.prompt_run_id = run.id
    JOIN public.todo_issues issue
      ON issue.id = link.issue_id AND issue.active_prompt_run_id = run.id
  LOOP
    IF public.gavel_touch_todo_issue(linked.issue_id, linked.activity_at) THEN
      changed_count := changed_count + 1;
    END IF;
  END LOOP;
  RETURN changed_count;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_todo_prompt_run_projection_trigger()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
  IF TG_OP = 'UPDATE' AND ROW(NEW.session_id, NEW.root_session_id)
    IS DISTINCT FROM ROW(OLD.session_id, OLD.root_session_id) THEN
    PERFORM public.gavel_project_todo_prompt_run(OLD.id);
  END IF;
  PERFORM public.gavel_project_todo_prompt_run(NEW.id);
  RETURN NEW;
END $$;

CREATE OR REPLACE FUNCTION public.gavel_todo_session_projection_trigger()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    PERFORM public.gavel_project_todo_session(OLD.id);
    RETURN OLD;
  END IF;
  IF TG_OP = 'UPDATE' AND ROW(NEW.provider_session_id, NEW.parent_session_id, NEW.root_session_id)
    IS DISTINCT FROM ROW(OLD.provider_session_id, OLD.parent_session_id, OLD.root_session_id) THEN
    PERFORM public.gavel_project_todo_session(OLD.id);
  END IF;
  PERFORM public.gavel_project_todo_session(NEW.id);
  RETURN NEW;
END $$;

CREATE OR REPLACE FUNCTION public.gavel_todo_request_projection_trigger()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF OLD.prompt_run_id IS NOT NULL THEN
      PERFORM public.gavel_touch_todo_prompt_run(
        OLD.prompt_run_id, GREATEST(OLD.created_at, COALESCE(OLD.resolved_at, OLD.created_at)));
      PERFORM public.gavel_project_todo_prompt_run(OLD.prompt_run_id);
    END IF;
    PERFORM public.gavel_project_todo_session(OLD.session_id);
    RETURN OLD;
  END IF;
  IF NEW.prompt_run_id IS NOT NULL THEN
    PERFORM public.gavel_touch_todo_prompt_run(
      NEW.prompt_run_id, GREATEST(NEW.created_at, COALESCE(NEW.resolved_at, NEW.created_at)));
    PERFORM public.gavel_project_todo_prompt_run(NEW.prompt_run_id);
  END IF;
  PERFORM public.gavel_project_todo_session(NEW.session_id);
  RETURN NEW;
END $$;

CREATE OR REPLACE FUNCTION public.gavel_todo_iteration_projection_trigger()
RETURNS trigger LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    PERFORM public.gavel_project_todo_prompt_run(OLD.prompt_run_id);
    RETURN OLD;
  END IF;
  IF TG_OP = 'UPDATE' AND OLD.prompt_run_id IS DISTINCT FROM NEW.prompt_run_id THEN
    PERFORM public.gavel_project_todo_prompt_run(OLD.prompt_run_id);
  END IF;
  PERFORM public.gavel_project_todo_prompt_run(NEW.prompt_run_id);
  RETURN NEW;
END $$;
