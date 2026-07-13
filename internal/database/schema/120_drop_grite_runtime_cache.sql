-- phase: post

-- Grite is no longer a runtime TODO provider. These two tables contained only
-- reconstructable provider cache state; native TODO import evidence lives in
-- the todo_* tables and in the explicit filesystem migration artifacts.
DROP TABLE IF EXISTS public.grite_sync_cursors, public.grite_issue_caches RESTRICT;
