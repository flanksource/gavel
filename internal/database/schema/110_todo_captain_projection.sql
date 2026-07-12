-- phase: post
-- dependsOn: 100_todo_captain_constraints.sql
-- runs: always

-- Project one issue from the current authoritative Captain rows. The issue row
-- lock serializes projection with native mutations and makes event sequence and
-- issue version allocation atomic. No event is emitted when neither durable
-- status nor externally visible execution state changed.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_issue(p_issue_id uuid)
RETURNS boolean
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  current_issue public.todo_issues%ROWTYPE;
  desired_status text;
  desired_execution_state text;
  prompt_step text;
  prompt_state text;
  prompt_phase text;
  prompt_verification text;
  prompt_spec jsonb;
  prompt_run_version bigint;
  prompt_root_session_id uuid;
  has_zombie boolean := false;
  has_stalled boolean := false;
  has_waiting_activity boolean := false;
  has_pending_request boolean := false;
  root_failed boolean := false;
  root_cancelled boolean := false;
  has_fixture boolean := false;
  auto_verify_without_fixture boolean := false;
  session_state_versions jsonb := '{}'::jsonb;
  request_versions jsonb := '{}'::jsonb;
  latest_verification_failed boolean := false;
  verification_policy text := NULL;
  next_version bigint;
  next_sequence bigint;
  event_kind text;
BEGIN
  SELECT *
  INTO current_issue
  FROM public.todo_issues
  WHERE id = p_issue_id
  FOR UPDATE;

  IF NOT FOUND THEN
    RETURN false;
  END IF;

  desired_status := current_issue.status;
  desired_execution_state := 'idle';

  IF current_issue.active_prompt_run_id IS NOT NULL THEN
    SELECT
      link.step_kind,
      run.state::text,
      run.phase::text,
      run.verification_markdown,
      run.rendered_spec,
      run.version,
      run.root_session_id
    INTO
      prompt_step,
      prompt_state,
      prompt_phase,
      prompt_verification,
      prompt_spec,
      prompt_run_version,
      prompt_root_session_id
    FROM public.todo_issue_prompt_runs AS link
    JOIN public.captain_prompt_runs AS run
      ON run.id = link.prompt_run_id
    WHERE link.issue_id = current_issue.id
      AND link.prompt_run_id = current_issue.active_prompt_run_id;

    IF FOUND THEN
      WITH RECURSIVE session_tree AS (
        SELECT
          session.id,
          session.parent_session_id,
          session.lifecycle_status::text AS lifecycle_status,
          session.activity_state::text AS activity_state,
          session.health_state::text AS health_state,
          session.state_version
        FROM public.captain_sessions AS session
        WHERE session.id = prompt_root_session_id
          OR session.root_session_id = prompt_root_session_id

        UNION

        SELECT
          child.id,
          child.parent_session_id,
          child.lifecycle_status::text,
          child.activity_state::text,
          child.health_state::text,
          child.state_version
        FROM public.captain_sessions AS child
        JOIN session_tree AS parent
          ON child.parent_session_id = parent.id
      )
      SELECT
        EXISTS (
          SELECT 1 FROM session_tree WHERE health_state = 'zombie'
        ),
        EXISTS (
          SELECT 1 FROM session_tree WHERE health_state = 'stalled'
        ),
        EXISTS (
          SELECT 1 FROM session_tree WHERE activity_state IN ('ask', 'approval')
        ),
        EXISTS (
          SELECT 1
          FROM public.captain_turn_requests AS request
          WHERE request.state::text = 'pending'
            AND (
              request.prompt_run_id = current_issue.active_prompt_run_id
              OR request.session_id IN (SELECT id FROM session_tree)
            )
        ),
        EXISTS (
          SELECT 1
          FROM session_tree
          WHERE id = prompt_root_session_id AND lifecycle_status = 'failed'
        ),
        EXISTS (
          SELECT 1
          FROM session_tree
          WHERE id = prompt_root_session_id
            AND lifecycle_status IN ('cancelled', 'interrupted')
        ),
        COALESCE((
          SELECT jsonb_object_agg(id::text, state_version ORDER BY id::text)
          FROM session_tree
        ), '{}'::jsonb),
        COALESCE((
          SELECT jsonb_object_agg(request.id::text, request.version ORDER BY request.id::text)
          FROM public.captain_turn_requests AS request
          WHERE request.prompt_run_id = current_issue.active_prompt_run_id
            OR request.session_id IN (SELECT id FROM session_tree)
        ), '{}'::jsonb)
      INTO
        has_zombie,
        has_stalled,
        has_waiting_activity,
        has_pending_request,
        root_failed,
        root_cancelled,
        session_state_versions,
        request_versions;

      SELECT COALESCE((
        SELECT
          iteration.state::text = 'failed'
            AND iteration.verification_result IS NOT NULL
        FROM public.captain_prompt_run_iterations AS iteration
        WHERE iteration.prompt_run_id = current_issue.active_prompt_run_id
        ORDER BY iteration.iteration DESC, iteration.created_at DESC, iteration.id DESC
        LIMIT 1
      ), false)
      INTO latest_verification_failed;

      has_fixture := btrim(COALESCE(prompt_verification, '')) <> '';
      -- Captain's rendered Workflow is the authoritative policy source. Exact
      -- jsonb equality accepts only the JSON boolean true; a string such as
      -- "true", an absent key, or malformed policy data is always false.
      auto_verify_without_fixture := COALESCE(
        prompt_spec #> '{workflow,autoVerifyWithoutFixture}' = 'true'::jsonb,
        false
      );

      IF prompt_state = 'cancelled' OR root_cancelled THEN
        desired_execution_state := 'idle';
      ELSIF prompt_state = 'failed' THEN
        IF prompt_step = 'verify'
          OR (prompt_step = 'run' AND (
            prompt_phase = 'verify' OR latest_verification_failed
          )) THEN
          desired_execution_state := 'verification_failed';
          IF current_issue.status IN ('draft', 'open', 'verified') THEN
            desired_status := 'open';
          END IF;
        ELSE
          desired_execution_state := 'failed';
        END IF;
      ELSIF has_zombie OR root_failed THEN
        desired_execution_state := 'failed';
      ELSIF has_stalled THEN
        desired_execution_state := 'stalled';
      ELSIF prompt_state = 'waiting' OR has_pending_request OR has_waiting_activity THEN
        desired_execution_state := 'waiting';
      ELSIF prompt_state = 'succeeded' AND prompt_phase = 'finished' THEN
        desired_execution_state := 'idle';
        IF prompt_step = 'verify' THEN
          verification_policy := 'verify_step';
        ELSIF prompt_step = 'run' AND has_fixture THEN
          verification_policy := 'fixture';
        ELSIF prompt_step = 'run' AND auto_verify_without_fixture THEN
          verification_policy := 'auto_verify_without_fixture';
        END IF;

        IF verification_policy IS NOT NULL
          AND current_issue.status IN ('draft', 'open') THEN
          desired_status := 'verified';
        ELSIF verification_policy IS NULL
          AND prompt_step = 'run'
          AND current_issue.status IN ('draft', 'open', 'verified') THEN
          -- A prior verification result cannot silently carry over to a new
          -- successful run that supplied neither a fixture nor explicit
          -- auto-verification policy.
          desired_status := 'open';
        END IF;
      ELSIF prompt_phase = 'verify' OR prompt_step = 'verify' THEN
        desired_execution_state := 'verifying';
      ELSIF prompt_step = 'plan' THEN
        desired_execution_state := 'planning';
      ELSE
        desired_execution_state := 'running';
      END IF;
    END IF;
  END IF;

  IF desired_status = current_issue.status
    AND desired_execution_state = current_issue.execution_state THEN
    RETURN false;
  END IF;

  next_version := current_issue.version + 1;
  SELECT COALESCE(MAX(sequence), 0) + 1
  INTO next_sequence
  FROM public.todo_issue_events
  WHERE issue_id = current_issue.id;

  UPDATE public.todo_issues
  SET
    status = desired_status,
    execution_state = desired_execution_state,
    version = next_version,
    updated_at = now()
  WHERE id = current_issue.id;

  event_kind := CASE
    WHEN desired_execution_state = 'verification_failed'
      THEN 'verification_failed'
    WHEN desired_status = 'verified'
      AND desired_status IS DISTINCT FROM current_issue.status
      THEN 'verification_succeeded'
    WHEN desired_status IS DISTINCT FROM current_issue.status
      THEN 'verification_required'
    ELSE 'execution_state_changed'
  END;

  INSERT INTO public.todo_issue_events (
    id,
    issue_id,
    sequence,
    kind,
    actor,
    payload,
    source,
    source_id,
    created_at
  ) VALUES (
    gen_random_uuid(),
    current_issue.id,
    next_sequence,
    event_kind,
    'captain',
    jsonb_strip_nulls(jsonb_build_object(
      'promptRunId', current_issue.active_prompt_run_id,
      'promptRunVersion', prompt_run_version,
      'sessionStateVersions', session_state_versions,
      'requestVersions', request_versions,
      'latestVerificationFailed', latest_verification_failed,
      'stepKind', prompt_step,
      'oldStatus', current_issue.status,
      'newStatus', desired_status,
      'oldExecutionState', current_issue.execution_state,
      'newExecutionState', desired_execution_state,
      'verificationPolicy', verification_policy
    )),
    'captain-projection',
    format(
      'prompt-run:%s:version:%s:issue:%s:version:%s',
      COALESCE(current_issue.active_prompt_run_id::text, 'none'),
      COALESCE(prompt_run_version::text, 'none'),
      current_issue.id,
      next_version
    ),
    now()
  );

  RETURN true;
END
$$;

COMMENT ON FUNCTION public.gavel_project_todo_issue(uuid) IS
  'Projects current Captain state into one Gavel issue and emits an event only on a visible change';

-- Stable application entrypoint. Gavel calls this after atomically linking and
-- activating a Captain run; Captain triggers use the same path.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_prompt_run(p_prompt_run_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  linked_issue_id uuid;
  changed_count integer := 0;
BEGIN
  FOR linked_issue_id IN
    SELECT issue.id
    FROM public.todo_issues AS issue
    JOIN public.todo_issue_prompt_runs AS link
      ON link.issue_id = issue.id
      AND link.prompt_run_id = issue.active_prompt_run_id
    WHERE link.prompt_run_id = p_prompt_run_id
  LOOP
    IF public.gavel_project_todo_issue(linked_issue_id) THEN
      changed_count := changed_count + 1;
    END IF;
  END LOOP;
  RETURN changed_count;
END
$$;

COMMENT ON FUNCTION public.gavel_project_todo_prompt_run(uuid) IS
  'Projects every active Gavel issue linked to one Captain prompt run; returns the number changed';

-- Session activity and health can come from any descendant of a prompt root.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_session(p_session_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  linked_prompt_run_id uuid;
  changed_count integer := 0;
BEGIN
  FOR linked_prompt_run_id IN
    WITH RECURSIVE session_ancestors AS (
      SELECT session.id, session.parent_session_id, session.root_session_id
      FROM public.captain_sessions AS session
      WHERE session.id = p_session_id

      UNION

      SELECT parent.id, parent.parent_session_id, parent.root_session_id
      FROM public.captain_sessions AS parent
      JOIN session_ancestors AS child
        ON child.parent_session_id = parent.id
    ), session_roots AS (
      SELECT COALESCE(root_session_id, id) AS id
      FROM session_ancestors

      UNION

      SELECT id FROM session_ancestors
    )
    SELECT DISTINCT run.id
    FROM public.captain_prompt_runs AS run
    JOIN public.todo_issue_prompt_runs AS link
      ON link.prompt_run_id = run.id
    JOIN public.todo_issues AS issue
      ON issue.id = link.issue_id
      AND issue.active_prompt_run_id = run.id
    WHERE run.root_session_id IN (SELECT id FROM session_roots)
  LOOP
    changed_count := changed_count
      + public.gavel_project_todo_prompt_run(linked_prompt_run_id);
  END LOOP;
  RETURN changed_count;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_todo_prompt_run_projection_trigger()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  -- Captain's state version deliberately covers lifecycle/result changes, not
  -- immutable association and rendered-policy fields. If one of those fields
  -- is corrected, project the authoritative row even though the version is
  -- unchanged. Revisit the old aggregate as well when a run moves between
  -- session trees so no old-root projection is left stale.
  IF TG_OP = 'UPDATE'
    AND ROW(
      NEW.verification_markdown,
      NEW.rendered_spec,
      NEW.session_id,
      NEW.root_session_id
    ) IS DISTINCT FROM ROW(
      OLD.verification_markdown,
      OLD.rendered_spec,
      OLD.session_id,
      OLD.root_session_id
    )
  THEN
    IF ROW(NEW.session_id, NEW.root_session_id)
      IS DISTINCT FROM ROW(OLD.session_id, OLD.root_session_id)
    THEN
      PERFORM public.gavel_project_todo_session(
        COALESCE(OLD.root_session_id, OLD.session_id)
      );
    END IF;
    PERFORM public.gavel_project_todo_prompt_run(NEW.id);
    RETURN NEW;
  END IF;

  -- Captain versions are monotonic. A replay at the same version or an
  -- out-of-order lower version must not regress a visible Gavel projection.
  IF TG_OP = 'UPDATE' AND NEW.version <= OLD.version THEN
    RETURN NEW;
  END IF;
  PERFORM public.gavel_project_todo_prompt_run(NEW.id);
  RETURN NEW;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_todo_session_projection_trigger()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF COALESCE(OLD.root_session_id, OLD.parent_session_id) IS NOT NULL THEN
      PERFORM public.gavel_project_todo_session(
        COALESCE(OLD.root_session_id, OLD.parent_session_id)
      );
    END IF;
    RETURN OLD;
  END IF;

  IF TG_OP = 'UPDATE'
    AND ROW(NEW.parent_session_id, NEW.root_session_id)
      IS DISTINCT FROM ROW(OLD.parent_session_id, OLD.root_session_id)
  THEN
    IF COALESCE(OLD.root_session_id, OLD.parent_session_id) IS NOT NULL THEN
      PERFORM public.gavel_project_todo_session(
        COALESCE(OLD.root_session_id, OLD.parent_session_id)
      );
    END IF;
    PERFORM public.gavel_project_todo_session(NEW.id);
    RETURN NEW;
  END IF;
  IF TG_OP = 'UPDATE' AND NEW.state_version <= OLD.state_version THEN
    RETURN NEW;
  END IF;
  PERFORM public.gavel_project_todo_session(NEW.id);
  RETURN NEW;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_todo_request_projection_trigger()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    IF OLD.prompt_run_id IS NOT NULL THEN
      PERFORM public.gavel_project_todo_prompt_run(OLD.prompt_run_id);
    END IF;
    PERFORM public.gavel_project_todo_session(OLD.session_id);
    RETURN OLD;
  END IF;

  IF TG_OP = 'UPDATE'
    AND ROW(OLD.prompt_run_id, OLD.session_id)
      IS DISTINCT FROM ROW(NEW.prompt_run_id, NEW.session_id)
  THEN
    IF OLD.prompt_run_id IS NOT NULL THEN
      PERFORM public.gavel_project_todo_prompt_run(OLD.prompt_run_id);
    END IF;
    PERFORM public.gavel_project_todo_session(OLD.session_id);
    IF NEW.prompt_run_id IS NOT NULL THEN
      PERFORM public.gavel_project_todo_prompt_run(NEW.prompt_run_id);
    END IF;
    PERFORM public.gavel_project_todo_session(NEW.session_id);
    RETURN NEW;
  END IF;
  IF TG_OP = 'UPDATE' AND NEW.version <= OLD.version THEN
    RETURN NEW;
  END IF;

  IF NEW.prompt_run_id IS NOT NULL THEN
    PERFORM public.gavel_project_todo_prompt_run(NEW.prompt_run_id);
  END IF;
  PERFORM public.gavel_project_todo_session(NEW.session_id);
  RETURN NEW;
END
$$;

CREATE OR REPLACE FUNCTION public.gavel_todo_iteration_projection_trigger()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
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
END
$$;

CREATE OR REPLACE TRIGGER gavel_todo_prompt_run_projection
AFTER INSERT OR UPDATE OF
  phase,
  state,
  verification_markdown,
  rendered_spec,
  session_id,
  root_session_id,
  version
ON public.captain_prompt_runs
FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_prompt_run_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_session_projection
AFTER INSERT OR DELETE OR UPDATE OF
  parent_session_id,
  root_session_id,
  lifecycle_status,
  activity_state,
  health_state,
  state_version
ON public.captain_sessions
FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_session_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_turn_request_projection
AFTER INSERT OR DELETE OR UPDATE OF
  prompt_run_id,
  session_id,
  kind,
  state,
  version
ON public.captain_turn_requests
FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_request_projection_trigger();

CREATE OR REPLACE TRIGGER gavel_todo_prompt_run_iteration_projection
AFTER INSERT OR DELETE OR UPDATE OF
  prompt_run_id,
  iteration,
  state,
  verification_result
ON public.captain_prompt_run_iterations
FOR EACH ROW
EXECUTE FUNCTION public.gavel_todo_iteration_projection_trigger();
