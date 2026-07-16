-- phase: post
-- dependsOn: 111_todo_projection_triggers.sql

-- One-time migration for issues whose monitored agent activity already
-- predates the activity propagation trigger.
WITH agent_roots AS (
  SELECT link.issue_id, agent.id AS agent_root_id
  FROM public.todo_issue_prompt_runs link
  JOIN public.captain_prompt_runs run ON run.id = link.prompt_run_id
  JOIN public.captain_sessions admission ON admission.id = run.root_session_id
  JOIN public.captain_sessions agent
    ON admission.provider_session_id IS NOT NULL
   AND agent.provider_session_id = admission.provider_session_id
   AND agent.source IN ('claude', 'codex')
), activity AS (
  SELECT agent_roots.issue_id,
         MAX(GREATEST(
           COALESCE(session.last_activity_at, '-infinity'::timestamptz),
           COALESCE(session.state_observed_at, '-infinity'::timestamptz),
           COALESCE(session.started_at, '-infinity'::timestamptz),
           COALESCE(session.ended_at, '-infinity'::timestamptz),
           session.updated_at
         )) AS activity_at
  FROM agent_roots
  JOIN public.captain_sessions session
    ON session.id = agent_roots.agent_root_id
    OR session.root_session_id = agent_roots.agent_root_id
  GROUP BY agent_roots.issue_id
)
UPDATE public.todo_issues issue
   SET updated_at = GREATEST(issue.updated_at, activity.activity_at)
  FROM activity
 WHERE issue.id = activity.issue_id
   AND issue.updated_at < activity.activity_at;
