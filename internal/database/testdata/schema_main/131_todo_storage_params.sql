-- phase: post

-- todo_issues and todo_issue_prompt_runs are update-in-place working sets, not
-- append-only logs. An issue row is rewritten on every status change, execution
-- state transition, label edit and body edit; a link row is rewritten by every
-- ownership heartbeat while a run is in flight (see todos/native/ownership.go).
-- Both tables stay small -- roughly a thousand issues and a few hundred links on
-- a busy developer database -- so the default 20% autovacuum scale factor is a
-- threshold they effectively never reach, and the default fillfactor of 100
-- leaves no in-page room, so each rewrite migrates the tuple to a new page
-- instead of staying HOT.
--
-- captain_sessions and captain_model_calls already carry this treatment for the
-- same reason; these two were simply never given it. This is preventative: at
-- the time of writing neither table shows meaningful bloat, and profiling the
-- todo read path found no vacuum-related cost. It keeps them that way as the
-- backlog and its run history grow.
--
-- Atlas OSS does not model PostgreSQL storage parameters, so the post-HCL SQL
-- phase owns them. This script is hash-gated: it re-runs only when its content
-- changes. If an HCL change recreates either table, bump this script so the
-- storage parameters are restored.
ALTER TABLE public.todo_issues SET (
  fillfactor = 70,
  autovacuum_vacuum_scale_factor = 0.02,
  autovacuum_analyze_scale_factor = 0.02,
  autovacuum_vacuum_cost_delay = 0
);

ALTER TABLE public.todo_issue_prompt_runs SET (
  fillfactor = 70,
  autovacuum_vacuum_scale_factor = 0.02,
  autovacuum_analyze_scale_factor = 0.02,
  autovacuum_vacuum_cost_delay = 0
);
