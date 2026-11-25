-- phase: post
-- dependsOn: 112_view_todo_issue_runtime.sql

CREATE OR REPLACE VIEW public.todo_issue_overview AS
SELECT issue.id AS issue_id,
       runtime.execution_state,
       run.id AS overview_run_id,
       admission.provider_session_id,
       COALESCE(run.result_text, '') AS result_text,
       COALESCE(run.finished_at, run.started_at, run.queued_at) AS run_at,
       COALESCE(link.step_kind::text, '') AS step_kind,
       (SELECT count(*) FROM public.todo_issue_prompt_runs history WHERE history.issue_id = issue.id) AS attempts,
       plan.id AS overview_plan_id,
       COALESCE(plan.approval_state::text, '') AS approval_state,
       plan.latest_revision AS plan_revision,
       overview.cost_usd,
       overview.iteration_count,
       overview.latest_verification_result,
       session.summary AS session_summary
FROM public.todo_issues issue
LEFT JOIN public.todo_issue_runtime runtime ON runtime.issue_id = issue.id
LEFT JOIN public.captain_prompt_runs run ON run.id = issue.active_prompt_run_id
LEFT JOIN LATERAL (
  SELECT cost_usd, iteration_count, latest_verification_result
  FROM public.captain_prompt_run_overview
  WHERE id = issue.active_prompt_run_id
  OFFSET 0
) overview ON true
LEFT JOIN public.captain_sessions admission ON admission.id = run.session_id
LEFT JOIN public.todo_issue_prompt_runs link
  ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
LEFT JOIN LATERAL (
  SELECT id, approval_state, latest_revision
  FROM public.captain_session_plans
  WHERE id = issue.selected_plan_id
  OFFSET 0
) plan ON true
LEFT JOIN LATERAL (
  WITH RECURSIVE candidates AS (
    SELECT identity.id,
      (identity.path IS NOT NULL OR metadata.history_file IS NOT NULL OR metadata.message_count > 0 OR metadata.turn_count > 0) AS bearing
    FROM public.captain_sessions identity
    LEFT JOIN LATERAL (
      SELECT history_file, message_count, turn_count FROM public.captain_session_overview
      WHERE id = identity.id OFFSET 0
    ) metadata ON true
    WHERE identity.provider_session_id = admission.provider_session_id AND identity.source IN ('claude', 'codex', 'api')
  ), root_id AS (
    SELECT (
      SELECT id FROM candidates WHERE bearing OR NOT EXISTS (SELECT 1 FROM candidates WHERE bearing)
    ) AS id
  ), thread AS (
    SELECT id FROM root_id WHERE id IS NOT NULL
    UNION ALL
    SELECT child.id FROM public.captain_sessions child JOIN thread parent ON
      (child.parent_session_id = parent.id OR (child.root_session_id = parent.id AND child.parent_session_id IS NULL))
      AND child.parent_relation IS DISTINCT FROM 'transcript'
  ), rows AS (
    SELECT summary.* FROM thread JOIN LATERAL (
      SELECT id, lifecycle_status, health_state, activity_state, started_at, created_at,
        last_activity_at, process_active, cost_usd, pending_request_count
      FROM public.captain_session_overview WHERE id = thread.id OFFSET 0
    ) summary ON true
  ), totals AS (
    SELECT sum(cost_usd) AS cost, max(last_activity_at) AS last_activity,
      bool_or(lifecycle_status = 'running' OR process_active) AS live
    FROM rows
  )
  SELECT jsonb_build_object(
    'found', true, 'inProgress', totals.live,
    'state', CASE
      WHEN root.pending_request_count > 0 THEN 'approval'
      WHEN run.state = 'waiting' AND run.result_json ->> 'endStatus' = 'ask' THEN 'ask'
      WHEN root.lifecycle_status = 'failed' OR root.health_state = 'zombie' THEN 'error'
      WHEN root.health_state = 'stalled' THEN 'stalled'
      WHEN root.activity_state <> 'idle' THEN root.activity_state::text
      WHEN totals.live THEN 'working'
      WHEN root.lifecycle_status = 'succeeded' THEN 'completed'
      ELSE root.lifecycle_status::text END,
    'durationMs', GREATEST(0, EXTRACT(EPOCH FROM (
      CASE WHEN totals.live THEN statement_timestamp() ELSE COALESCE(totals.last_activity, root.started_at, root.created_at) END
      - COALESCE(root.started_at, root.created_at))) * 1000)::bigint,
    'costUsd', COALESCE(totals.cost, 0),
    'updatedAt', statement_timestamp()
  ) AS summary
  FROM totals CROSS JOIN rows root
  WHERE root.id = (SELECT id FROM root_id)
) session ON true;
