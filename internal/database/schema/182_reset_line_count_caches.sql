-- phase: post

-- Range totals and per-file summaries previously included non-SLOC assets.
-- Recompute them and worktree counts even when commits and files are unchanged.
DELETE FROM public.git_range_stats;
UPDATE public.git_repos SET refs_fingerprint = '';
UPDATE public.git_worktrees SET status_fingerprint = '';
