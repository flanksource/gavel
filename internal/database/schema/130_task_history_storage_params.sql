-- phase: post

-- task_run_history is an update-in-place ring: the live set stays at a few
-- dozen rows while the `run` and `snapshots` JSONB documents are rewritten
-- continuously as a run progresses. Measured on a developer database before
-- this change: 475,081 updates against a single insert and 40 live rows,
-- leaving 1,734 dead tuples spread over 387 heap pages -- roughly 63 KB of
-- heap per surviving row -- and driving 109 autovacuum cycles.
--
-- The default fillfactor of 100 is the reason autovacuum could not keep the
-- table compact: with no free space reserved in a page, each rewrite migrates
-- the tuple to a new page instead of staying HOT, and plain VACUUM cannot
-- return those pages to the file. Reserving in-page room keeps the rewrite
-- HOT, and the lowered scale factors matter on a table far too small for the
-- default 20% threshold to ever be a meaningful trigger.
--
-- Atlas OSS does not model PostgreSQL storage parameters, so the post-HCL SQL
-- phase owns them. This script is hash-gated: it re-runs only when its content
-- changes. If an HCL change recreates task_run_history, bump this script so the
-- storage parameters are restored.
ALTER TABLE public.task_run_history SET (
  fillfactor = 70,
  autovacuum_vacuum_scale_factor = 0.01,
  autovacuum_analyze_scale_factor = 0.01,
  autovacuum_vacuum_cost_delay = 0
);
