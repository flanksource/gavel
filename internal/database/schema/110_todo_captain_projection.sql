-- phase: post
-- dependsOn: 100_todo_captain_constraints.sql
-- runs: always

-- Derive transient execution state from Captain at read time. The source='gavel'
-- session remains only the prompt-run admission root; monitored Claude/Codex
-- sessions with the same provider identity own agent lifecycle and health.
CREATE OR REPLACE FUNCTION public.gavel_todo_issue_execution_state(p_issue_id uuid)
RETURNS text
LANGUAGE sql
STABLE
SET search_path = pg_catalog, public
AS $$
WITH active AS (
  SELECT
    issue.active_prompt_run_id,
    link.step_kind,
    run.state::text AS prompt_state,
    run.phase::text AS prompt_phase,
    run.root_session_id,
    root.provider_session_id
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
    CASE WHEN active.step_kind = 'verify'
           OR (active.step_kind = 'run' AND (active.prompt_phase = 'verify' OR latest_iteration.verification_failed))
         THEN 'verification_failed' ELSE 'failed' END
  WHEN signals.failed THEN 'failed'
  WHEN signals.stalled THEN 'stalled'
  WHEN active.prompt_state = 'waiting' OR signals.waiting OR pending.waiting THEN 'waiting'
  WHEN signals.terminal THEN 'idle'
  WHEN active.prompt_state = 'succeeded' AND active.prompt_phase = 'finished' THEN 'idle'
  WHEN active.prompt_phase = 'verify' OR active.step_kind = 'verify' THEN 'verifying'
  WHEN active.step_kind = 'plan' THEN 'planning'
  ELSE 'running'
END
FROM active
CROSS JOIN signals
CROSS JOIN latest_iteration
CROSS JOIN pending
$$;

CREATE OR REPLACE VIEW public.todo_issue_runtime AS
SELECT issue.id AS issue_id,
       public.gavel_todo_issue_execution_state(issue.id) AS execution_state
FROM public.todo_issues issue;

COMMENT ON VIEW public.todo_issue_runtime IS
  'Read-time Captain-derived execution state for Gavel issues';

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

-- Reconcile only durable workflow status. Transient execution state is read
-- directly from Captain and is never copied into the issue row or event log.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_issue(p_issue_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  current_issue public.todo_issues%ROWTYPE;
  prompt_step text;
  prompt_state text;
  prompt_phase text;
  prompt_verification text;
  prompt_spec jsonb;
  prompt_run_version bigint;
  latest_verification_failed boolean := false;
  desired_status text;
  verification_policy text := NULL;
  next_version bigint;
  next_sequence bigint;
  event_kind text;
BEGIN
  SELECT * INTO current_issue FROM public.todo_issues WHERE id = p_issue_id FOR UPDATE;
  IF NOT FOUND OR current_issue.active_prompt_run_id IS NULL THEN
    RETURN false;
  END IF;

  SELECT link.step_kind, run.state::text, run.phase::text,
         run.verification_markdown, run.rendered_spec, run.version
    INTO prompt_step, prompt_state, prompt_phase,
         prompt_verification, prompt_spec, prompt_run_version
  FROM public.todo_issue_prompt_runs link
  JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
  WHERE link.issue_id = current_issue.id
    AND link.prompt_run_id = current_issue.active_prompt_run_id;
  IF NOT FOUND THEN
    RETURN false;
  END IF;

  SELECT COALESCE((
    SELECT iteration.state::text = 'failed' AND iteration.verification_result IS NOT NULL
    FROM public.captain_prompt_run_iterations iteration
    WHERE iteration.prompt_run_id = current_issue.active_prompt_run_id
    ORDER BY iteration.iteration DESC, iteration.created_at DESC, iteration.id DESC
    LIMIT 1
  ), false) INTO latest_verification_failed;

  desired_status := current_issue.status;
  IF prompt_state = 'failed'
     AND (prompt_step = 'verify'
       OR (prompt_step = 'run' AND (prompt_phase = 'verify' OR latest_verification_failed)))
     AND current_issue.status IN ('draft', 'open', 'verified') THEN
    desired_status := 'open';
  ELSIF prompt_state = 'succeeded' AND prompt_phase = 'finished' THEN
    IF prompt_step = 'verify' THEN
      verification_policy := 'verify_step';
    ELSIF prompt_step = 'run' AND btrim(COALESCE(prompt_verification, '')) <> '' THEN
      verification_policy := 'fixture';
    ELSIF prompt_step = 'run'
      AND COALESCE(prompt_spec #> '{workflow,autoVerifyWithoutFixture}' = 'true'::jsonb, false) THEN
      verification_policy := 'auto_verify_without_fixture';
    END IF;
    IF verification_policy IS NOT NULL AND current_issue.status IN ('draft', 'open') THEN
      desired_status := 'verified';
    ELSIF verification_policy IS NULL AND prompt_step = 'run'
      AND current_issue.status IN ('draft', 'open', 'verified') THEN
      desired_status := 'open';
    END IF;
  END IF;

  IF desired_status = current_issue.status THEN
    RETURN false;
  END IF;

  next_version := current_issue.version + 1;
  SELECT COALESCE(MAX(sequence), 0) + 1 INTO next_sequence
  FROM public.todo_issue_events WHERE issue_id = current_issue.id;

  UPDATE public.todo_issues
     SET status = desired_status, version = next_version,
         updated_at = GREATEST(updated_at, now())
   WHERE id = current_issue.id;

  event_kind := CASE
    WHEN prompt_state = 'failed' THEN 'verification_failed'
    WHEN desired_status = 'verified' THEN 'verification_succeeded'
    WHEN desired_status = 'open' AND current_issue.status = 'verified' THEN 'verification_required'
    ELSE 'verification_failed'
  END;

  INSERT INTO public.todo_issue_events (
    id, issue_id, sequence, kind, actor, payload, source, source_id, created_at
  ) VALUES (
    gen_random_uuid(), current_issue.id, next_sequence, event_kind, 'captain',
    jsonb_strip_nulls(jsonb_build_object(
      'promptRunId', current_issue.active_prompt_run_id,
      'promptRunVersion', prompt_run_version,
      'latestVerificationFailed', latest_verification_failed,
      'stepKind', prompt_step,
      'oldStatus', current_issue.status,
      'newStatus', desired_status,
      'verificationPolicy', verification_policy
    )),
    'captain-projection',
    format('prompt-run:%s:version:%s:issue:%s:version:%s',
      current_issue.active_prompt_run_id, prompt_run_version,
      current_issue.id, next_version),
    now()
  );
  RETURN true;
END
$$;

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
    PERFORM public.gavel_touch_todo_issue(linked.issue_id, linked.activity_at);
    IF public.gavel_project_todo_issue(linked.issue_id) THEN
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
