-- phase: post
-- dependsOn: 110_todo_projection_functions.sql

CREATE OR REPLACE VIEW public.todo_issue_runtime AS
WITH active AS NOT MATERIALIZED (
  SELECT issue.id AS issue_id, issue.active_prompt_run_id,
         run.state::text AS prompt_state, run.phase::text AS prompt_phase,
         run.rendered_spec, root.provider_session_id,
         run.rendered_spec #> '{workflow,verify}' IS NOT NULL
           AND COALESCE(run.rendered_spec #>> '{prompt,user}', '') = '' AS verify_only
  FROM public.todo_issues issue
  LEFT JOIN public.todo_issue_prompt_runs link
    ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
  LEFT JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
  LEFT JOIN public.captain_sessions root ON root.id = run.root_session_id
), agent_roots AS NOT MATERIALIZED (
  SELECT active.issue_id, session.id
  FROM active
  JOIN public.captain_sessions session ON session.provider_session_id = active.provider_session_id
    AND session.source IN ('claude', 'codex')
), session_tree AS NOT MATERIALIZED (
  SELECT roots.issue_id, session.id, session.lifecycle_status, session.activity_state, session.health_state
  FROM agent_roots roots JOIN public.captain_sessions session ON session.id = roots.id
  UNION
  SELECT roots.issue_id, session.id, session.lifecycle_status, session.activity_state, session.health_state
  FROM agent_roots roots JOIN public.captain_sessions session ON session.root_session_id = roots.id
), signals AS (
  SELECT issue_id,
    bool_or(health_state = 'zombie' OR lifecycle_status = 'failed') AS failed,
    bool_or(health_state = 'stalled') AS stalled,
    bool_or(activity_state IN ('ask', 'approval')) AS waiting,
    bool_or(lifecycle_status IN ('succeeded', 'cancelled')) AS terminal
  FROM session_tree GROUP BY issue_id
), pending AS (
  SELECT active.issue_id FROM active
  JOIN public.captain_turn_requests request ON request.prompt_run_id = active.active_prompt_run_id
  WHERE request.state = 'pending'
  UNION
  SELECT tree.issue_id FROM session_tree tree
  JOIN public.captain_turn_requests request ON request.session_id = tree.id
  WHERE request.state = 'pending'
), latest_iteration AS (
  SELECT DISTINCT ON (iteration.prompt_run_id) iteration.prompt_run_id,
    iteration.state::text = 'failed' AND iteration.verification_result IS NOT NULL AS verification_failed
  FROM public.captain_prompt_run_iterations iteration
  ORDER BY iteration.prompt_run_id, iteration.iteration DESC, iteration.created_at DESC, iteration.id DESC
), outcomes AS (
  SELECT event.issue_id,
    max(sequence) FILTER (WHERE kind = 'lifecycle_outcome' AND payload->>'status' = 'ask') AS asked,
    max(sequence) FILTER (WHERE kind = 'ask_answered' AND source = 'captain'
      AND payload->>'promptRunId' = active.active_prompt_run_id::text) AS answered
  FROM public.todo_issue_events event JOIN active ON active.issue_id = event.issue_id
  GROUP BY event.issue_id
), unpersisted AS (
  SELECT DISTINCT ON (event.issue_id) event.issue_id,
    event.payload->>'status' = 'failed' AND event.payload->>'error' IS NOT NULL AS failed
  FROM public.todo_issue_events event JOIN active ON active.issue_id = event.issue_id
  WHERE event.kind = 'lifecycle_outcome' AND event.payload->>'promptRunId' = active.active_prompt_run_id::text
  ORDER BY event.issue_id, event.sequence DESC
)
SELECT active.issue_id, CASE
  WHEN active.active_prompt_run_id IS NULL THEN 'idle'
  WHEN active.prompt_state = 'cancelled' THEN 'idle'
  WHEN active.prompt_state = 'failed' THEN
    CASE WHEN active.verify_only OR active.prompt_phase = 'verify' OR latest_iteration.verification_failed
      THEN 'verification_failed' ELSE 'failed' END
  WHEN unpersisted.failed OR signals.failed THEN 'failed'
  WHEN signals.stalled THEN 'stalled'
  WHEN active.prompt_state = 'waiting' AND outcomes.answered > COALESCE(outcomes.asked, 0)
    AND pending.issue_id IS NULL THEN 'running'
  WHEN active.prompt_state = 'waiting' OR signals.waiting OR pending.issue_id IS NOT NULL THEN 'waiting'
  WHEN signals.terminal THEN 'idle'
  WHEN active.prompt_state = 'succeeded' AND active.prompt_phase = 'finished' THEN 'idle'
  WHEN active.verify_only OR active.prompt_phase = 'verify' THEN 'verifying'
  WHEN active.rendered_spec #>> '{permissions,mode}' = 'plan' THEN 'planning'
  ELSE 'running'
END AS execution_state
FROM active
LEFT JOIN signals USING (issue_id)
LEFT JOIN pending USING (issue_id)
LEFT JOIN outcomes USING (issue_id)
LEFT JOIN unpersisted USING (issue_id)
LEFT JOIN latest_iteration ON latest_iteration.prompt_run_id = active.active_prompt_run_id;

COMMENT ON VIEW public.todo_issue_runtime IS
  'Set-based, read-time Captain-derived execution state for Gavel issues';
