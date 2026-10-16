-- phase: post

-- The todo list's run diff badges used to be cached per workspace in
-- commit_stat_caches / commit_stat_cursors with a 60 s TTL. They are now read
-- from git_range_stats, keyed by the run's commits, which never go stale, so
-- the HCL no longer declares the old tables. Atlas leaves undeclared tables in
-- place (drops are suppressed for the shared database), so this one-time
-- script removes them from an existing database.
DROP TABLE IF EXISTS public.commit_stat_caches;
DROP TABLE IF EXISTS public.commit_stat_cursors;
