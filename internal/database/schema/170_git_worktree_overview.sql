-- phase: post

-- A status scan of an unchanged worktree used to read git_worktrees'
-- fingerprint and then rewrite its status_scanned_at: measured on a developer
-- database, 2.2M of each over nine days against 186 rows, every rewrite a new
-- WAL-logged tuple. That volatile pair lives here instead, in an UNLOGGED
-- table: confirming an unchanged scan is one UPDATE that writes no WAL, and the
-- logged git_worktrees row is rewritten only when the status changed.
--
-- Everything here is derivable from git_worktrees, which still records the
-- fingerprint and the scan time of the last scan that changed something. Crash
-- recovery truncates an UNLOGGED table; gavel_git_overview_ensure() (171)
-- rebuilds it when the tracker starts, and a worktree whose row is missing
-- simply runs one full status scan, whose write refills the row by trigger.
--
-- Atlas does not model UNLOGGED tables, so this post-phase script owns the
-- table and the migrator excludes it from inspection (internal/database).
CREATE UNLOGGED TABLE IF NOT EXISTS public.git_worktree_overview (
  repo_id            uuid        NOT NULL,
  path               text        NOT NULL,
  -- NULL while the worktree's last scan failed, so the next scan is a full one.
  status_fingerprint text,
  status_scanned_at  timestamptz,
  PRIMARY KEY (repo_id, path),
  FOREIGN KEY (repo_id, path) REFERENCES public.git_worktrees (repo_id, path) ON DELETE CASCADE
);

-- Mirrors a written status (or a recorded scan failure) of one worktree.
CREATE OR REPLACE FUNCTION public.gavel_git_worktree_overview_sync()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  INSERT INTO public.git_worktree_overview AS overview (repo_id, path, status_fingerprint, status_scanned_at)
  VALUES (
    NEW.repo_id, NEW.path,
    CASE WHEN NEW.status_error IS NULL THEN NEW.status_fingerprint END,
    NEW.status_scanned_at
  )
  ON CONFLICT (repo_id, path) DO UPDATE
    SET status_fingerprint = EXCLUDED.status_fingerprint,
        status_scanned_at = GREATEST(overview.status_scanned_at, EXCLUDED.status_scanned_at);
  RETURN NULL;
END
$$;

DROP TRIGGER IF EXISTS git_worktree_overview_sync ON public.git_worktrees;
CREATE TRIGGER git_worktree_overview_sync
  AFTER INSERT OR UPDATE OF status_fingerprint, status_error, status_scanned_at ON public.git_worktrees
  FOR EACH ROW EXECUTE FUNCTION public.gavel_git_worktree_overview_sync();

-- Recomputes every overview row from git_worktrees. A rebuilt row keeps the
-- later of its own scan time and the logged one.
CREATE OR REPLACE FUNCTION public.gavel_git_worktree_overview_rebuild()
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  affected integer := 0;
BEGIN
  INSERT INTO public.git_worktree_overview AS overview (repo_id, path, status_fingerprint, status_scanned_at)
  SELECT worktree.repo_id, worktree.path,
         CASE WHEN worktree.status_error IS NULL THEN worktree.status_fingerprint END,
         worktree.status_scanned_at
  FROM public.git_worktrees worktree
  ON CONFLICT (repo_id, path) DO UPDATE
    SET status_fingerprint = EXCLUDED.status_fingerprint,
        status_scanned_at = GREATEST(overview.status_scanned_at, EXCLUDED.status_scanned_at);
  GET DIAGNOSTICS affected = ROW_COUNT;
  RETURN affected;
END
$$;

SELECT public.gavel_git_worktree_overview_rebuild();
