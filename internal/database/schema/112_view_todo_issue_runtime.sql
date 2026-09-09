-- phase: post
-- dependsOn: 110_todo_projection_functions.sql

CREATE OR REPLACE VIEW public.todo_issue_runtime AS
SELECT issue.id AS issue_id,
       public.gavel_todo_issue_execution_state(issue.id) AS execution_state
FROM public.todo_issues issue;

COMMENT ON VIEW public.todo_issue_runtime IS
  'Read-time Captain-derived execution state for Gavel issues';
