-- phase: post
-- dependsOn: 090_prepare_runtime_state.sql

-- 090 drops the execution-state and touch functions below whenever it re-runs,
-- so this file must re-run with it.
--
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
  -- Two indexed lookups joined by UNION, not one OR: the OR form cannot use
  -- captain_sessions_pkey / captain_sessions_root_session_id_idx and seq-scanned
  -- every session once per issue with an active run.
  SELECT session.id, session.lifecycle_status::text AS lifecycle_status,
         session.activity_state::text AS activity_state,
         session.health_state::text AS health_state
  FROM public.captain_sessions session
  WHERE session.id IN (SELECT id FROM agent_roots)
  UNION
  SELECT session.id, session.lifecycle_status::text,
         session.activity_state::text,
         session.health_state::text
  FROM public.captain_sessions session
  WHERE session.root_session_id IN (SELECT id FROM agent_roots)
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
), answered AS (
  -- A parked ask is answered once a Captain ask_answered event for the active
  -- run is newer than the latest ask outcome. Captain can resume a parked run's
  -- session without gavel; the read that notices records the answer, and the run
  -- stays `waiting` until that turn settles — but it is no longer asking.
  -- Gavel's own answers are not read here: a resume gavel drives moves the run
  -- itself, and one whose dispatcher died after admission is still asking.
  SELECT EXISTS (
    SELECT 1
    FROM active
    JOIN public.todo_issue_events answer
      ON answer.issue_id = p_issue_id
     AND answer.kind = 'ask_answered'
     AND answer.source = 'captain'
     AND answer.payload ->> 'promptRunId' = active.active_prompt_run_id::text
    WHERE active.prompt_state = 'waiting'
      AND answer.sequence > COALESCE((
      SELECT max(outcome.sequence)
      FROM public.todo_issue_events outcome
      WHERE outcome.issue_id = p_issue_id
        AND outcome.kind = 'lifecycle_outcome'
        AND outcome.payload ->> 'status' = 'ask'
    ), 0)
  ) AS after_ask
)
SELECT CASE
  WHEN active.active_prompt_run_id IS NULL THEN 'idle'
  WHEN active.prompt_state = 'cancelled' THEN 'idle'
  WHEN active.prompt_state = 'failed' THEN
    CASE WHEN active.verify_only OR active.prompt_phase = 'verify' OR latest_iteration.verification_failed
         THEN 'verification_failed' ELSE 'failed' END
  WHEN signals.failed THEN 'failed'
  WHEN signals.stalled THEN 'stalled'
  WHEN active.prompt_state = 'waiting' AND answered.after_ask AND NOT pending.waiting THEN 'running'
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
CROSS JOIN answered
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

-- The functions above and below are called by Gavel's row-change projection
-- (internal/database/todoprojection), which LISTENs on Captain's
-- captain_row_change channel. Gavel installs no trigger on a Captain table: it
-- is told which row changed after Captain commits and re-reads it here.

-- A turn request dates its run's activity by when it was raised or resolved.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_turn_request(p_request_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  request record;
BEGIN
  SELECT turn.prompt_run_id,
         GREATEST(turn.created_at, COALESCE(turn.resolved_at, turn.created_at)) AS activity_at
    INTO request
    FROM public.captain_turn_requests turn
   WHERE turn.id = p_request_id;
  IF NOT FOUND OR request.prompt_run_id IS NULL THEN
    RETURN 0;
  END IF;
  RETURN public.gavel_touch_todo_prompt_run(request.prompt_run_id, request.activity_at);
END
$$;

-- A TODO's root session shares the TODO's id, so any session in that tree —
-- the root, its operation and plan-source sessions — is activity on the TODO.
CREATE OR REPLACE FUNCTION public.gavel_project_todo_root_session(p_session_id uuid)
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  target record;
BEGIN
  SELECT COALESCE(session.root_session_id, session.id) AS root_id,
         GREATEST(
           COALESCE(session.last_activity_at, '-infinity'::timestamptz),
           COALESCE(session.state_observed_at, '-infinity'::timestamptz),
           COALESCE(session.started_at, '-infinity'::timestamptz),
           COALESCE(session.ended_at, '-infinity'::timestamptz),
           session.updated_at
         ) AS activity_at
    INTO target
    FROM public.captain_sessions session
   WHERE session.id = p_session_id;
  IF NOT FOUND THEN
    RETURN 0;
  END IF;
  IF public.gavel_touch_todo_issue(target.root_id, target.activity_at) THEN
    RETURN 1;
  END IF;
  RETURN 0;
END
$$;

-- Nothing is delivered while the projection is not LISTENing, so on every
-- (re)established LISTEN it re-derives every watermark the live mapping could
-- have advanced: the active run, its turn requests, its agent session family
-- (the sessions sharing the admission root's provider identity, and their
-- descendants), and the TODO's own root session tree. Monotonic like every
-- touch: a watermark never moves backwards. Returns the issues advanced.
CREATE OR REPLACE FUNCTION public.gavel_resync_todo_activity()
RETURNS integer
LANGUAGE sql
SET search_path = pg_catalog, public
AS $$
WITH active AS (
  SELECT issue.id AS issue_id, run.id AS run_id, run.updated_at AS run_at,
         admission.provider_session_id
  FROM public.todo_issues issue
  JOIN public.todo_issue_prompt_runs link
    ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
  JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
  LEFT JOIN public.captain_sessions admission
    ON admission.id = run.root_session_id AND admission.source = 'gavel'
), agent_roots AS (
  SELECT active.issue_id, agent.id AS root_id
  FROM active
  JOIN public.captain_sessions agent
    ON active.provider_session_id IS NOT NULL
   AND agent.provider_session_id = active.provider_session_id
), tree_roots AS (
  SELECT issue.id AS issue_id, issue.id AS root_id FROM public.todo_issues issue
), roots AS (
  SELECT issue_id, root_id FROM agent_roots
  UNION ALL
  SELECT issue_id, root_id FROM tree_roots
), sessions AS (
  -- Two indexed lookups per root, joined by UNION ALL rather than one OR, for
  -- the same reason gavel_todo_issue_execution_state avoids the OR form.
  SELECT roots.issue_id, session.last_activity_at, session.state_observed_at,
         session.started_at, session.ended_at, session.updated_at
  FROM roots
  JOIN public.captain_sessions session ON session.id = roots.root_id
  UNION ALL
  SELECT roots.issue_id, session.last_activity_at, session.state_observed_at,
         session.started_at, session.ended_at, session.updated_at
  FROM roots
  JOIN public.captain_sessions session ON session.root_session_id = roots.root_id
), activity AS (
  SELECT active.issue_id, active.run_at AS activity_at FROM active
  UNION ALL
  SELECT active.issue_id,
         GREATEST(request.created_at, COALESCE(request.resolved_at, request.created_at))
  FROM active
  JOIN public.captain_turn_requests request ON request.prompt_run_id = active.run_id
  UNION ALL
  SELECT sessions.issue_id,
         GREATEST(
           COALESCE(sessions.last_activity_at, '-infinity'::timestamptz),
           COALESCE(sessions.state_observed_at, '-infinity'::timestamptz),
           COALESCE(sessions.started_at, '-infinity'::timestamptz),
           COALESCE(sessions.ended_at, '-infinity'::timestamptz),
           sessions.updated_at
         )
  FROM sessions
), latest AS (
  SELECT issue_id, max(activity_at) AS activity_at FROM activity GROUP BY issue_id
), advanced AS (
  UPDATE public.todo_issues issue
     SET updated_at = latest.activity_at
    FROM latest
   WHERE issue.id = latest.issue_id
     AND issue.updated_at < latest.activity_at
  RETURNING issue.id
)
SELECT count(*)::integer FROM advanced
$$;
