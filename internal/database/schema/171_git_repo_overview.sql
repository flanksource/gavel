-- phase: post
-- dependsOn: 170_git_worktree_overview.sql

-- The repo-level counterpart of git_worktree_overview (170), and the one place
-- a repository's change is announced.
--
-- An unchanged ref scan used to read git_repos' fingerprint and rewrite its
-- refs_scanned_at, and every write the UI could see advanced
-- git_repos.generation and sent gavel_git_change from Go (Store.bump), so a
-- write path that forgot to bump left the UI stale. Here the volatile pair and
-- the generation live in an UNLOGGED row, and triggers on the logged git_*
-- tables advance the generation and notify whenever a column a reader renders
-- actually changes, whoever wrote it.
--
-- generation comes from a logged sequence, not a counter on the row: crash
-- recovery truncates this table, and the rebuild must never hand a UI client a
-- generation lower than one it has already seen.
--
-- Atlas does not model UNLOGGED tables, so this post-phase script owns the
-- table and the migrator excludes it from inspection (internal/database).
CREATE SEQUENCE IF NOT EXISTS public.git_repo_generation_seq;

CREATE UNLOGGED TABLE IF NOT EXISTS public.git_repo_overview (
  repo_id          uuid        PRIMARY KEY REFERENCES public.git_repos (id) ON DELETE CASCADE,
  -- NULL while the repo's last ref scan failed, so the next scan is a full one.
  refs_fingerprint text,
  refs_scanned_at  timestamptz,
  generation       bigint      NOT NULL
);

-- Advances the repo's generation and announces it on gavel_git_change, the
-- channel pr/ui's GitChangeListener (gitstate.ChangeChannel) LISTENs on. A repo
-- with no overview row yet, as after crash recovery, gets one.
CREATE OR REPLACE FUNCTION public.gavel_git_repo_changed(p_repo_id uuid)
RETURNS void
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  INSERT INTO public.git_repo_overview AS overview (repo_id, refs_fingerprint, refs_scanned_at, generation)
  SELECT repo.id, CASE WHEN repo.error IS NULL THEN repo.refs_fingerprint END, repo.refs_scanned_at,
         nextval('public.git_repo_generation_seq')
  FROM public.git_repos repo WHERE repo.id = p_repo_id
  ON CONFLICT (repo_id) DO UPDATE SET generation = EXCLUDED.generation;
  PERFORM pg_notify('gavel_git_change', p_repo_id::text);
END
$$;

-- git_repos: a registered repo gets its row, and a ref scan's write or a
-- recorded failure mirrors the fingerprint. The mirror is unconditional, so a
-- row crash recovery removed comes back on the repo's next written scan (an
-- unchanged scan finds no row, so it writes); only a change to what a reader
-- renders, or a recreated row, is announced.
CREATE OR REPLACE FUNCTION public.gavel_git_repo_overview_sync()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  created boolean;
BEGIN
  INSERT INTO public.git_repo_overview AS overview (repo_id, refs_fingerprint, refs_scanned_at, generation)
  VALUES (NEW.id, CASE WHEN NEW.error IS NULL THEN NEW.refs_fingerprint END, NEW.refs_scanned_at,
          nextval('public.git_repo_generation_seq'))
  ON CONFLICT (repo_id) DO UPDATE
    SET refs_fingerprint = EXCLUDED.refs_fingerprint,
        refs_scanned_at = GREATEST(overview.refs_scanned_at, EXCLUDED.refs_scanned_at)
  RETURNING (xmax = 0) INTO created;
  IF TG_OP = 'UPDATE' AND (created OR
     (OLD.base_branch, OLD.base_sha, OLD.current_branch, OLD.base_checked_out, OLD.refs_fingerprint, OLD.error)
     IS DISTINCT FROM
     (NEW.base_branch, NEW.base_sha, NEW.current_branch, NEW.base_checked_out, NEW.refs_fingerprint, NEW.error)) THEN
    PERFORM public.gavel_git_repo_changed(NEW.id);
  END IF;
  RETURN NULL;
END
$$;

DROP TRIGGER IF EXISTS git_repo_overview_insert ON public.git_repos;
CREATE TRIGGER git_repo_overview_insert
  AFTER INSERT ON public.git_repos
  FOR EACH ROW EXECUTE FUNCTION public.gavel_git_repo_overview_sync();

DROP TRIGGER IF EXISTS git_repo_overview_update ON public.git_repos;
CREATE TRIGGER git_repo_overview_update
  AFTER UPDATE OF base_branch, base_sha, current_branch, base_checked_out, refs_fingerprint, refs_scanned_at, error
  ON public.git_repos
  FOR EACH ROW EXECUTE FUNCTION public.gavel_git_repo_overview_sync();

-- git_worktrees: a worktree appearing, disappearing, or changing anything
-- Store.Load renders. status_scanned_at and status_fingerprint are excluded: a
-- scan time is not a change, and the fingerprint only moves with the status.
CREATE OR REPLACE FUNCTION public.gavel_git_worktree_changed()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
  PERFORM public.gavel_git_repo_changed(CASE WHEN TG_OP = 'DELETE' THEN OLD.repo_id ELSE NEW.repo_id END);
  RETURN NULL;
END
$$;

DROP TRIGGER IF EXISTS git_worktree_changed_insert_delete ON public.git_worktrees;
CREATE TRIGGER git_worktree_changed_insert_delete
  AFTER INSERT OR DELETE ON public.git_worktrees
  FOR EACH ROW EXECUTE FUNCTION public.gavel_git_worktree_changed();

DROP TRIGGER IF EXISTS git_worktree_changed_update ON public.git_worktrees;
CREATE TRIGGER git_worktree_changed_update
  AFTER UPDATE ON public.git_worktrees
  FOR EACH ROW
  WHEN ((OLD.branch, OLD.head_sha, OLD.is_primary, OLD.detached, OLD.prunable, OLD.ordinal, OLD.last_commit_at,
         OLD.staged, OLD.unstaged, OLD."both", OLD.untracked, OLD.conflict, OLD.adds, OLD.dels, OLD.touched_at,
         OLD.files, OLD.status_error, OLD.removed_at)
        IS DISTINCT FROM
        (NEW.branch, NEW.head_sha, NEW.is_primary, NEW.detached, NEW.prunable, NEW.ordinal, NEW.last_commit_at,
         NEW.staged, NEW.unstaged, NEW."both", NEW.untracked, NEW.conflict, NEW.adds, NEW.dels, NEW.touched_at,
         NEW.files, NEW.status_error, NEW.removed_at))
  EXECUTE FUNCTION public.gavel_git_worktree_changed();

-- git_branches and git_range_stats: ApplyRefs replaces every branch row of a
-- repo in two statements, so these are statement-level and announce each repo
-- a statement touched once. A range is written once and never updated in a way
-- Store.Load renders (file_list is read on demand).
CREATE OR REPLACE FUNCTION public.gavel_git_rows_changed()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  changed uuid;
BEGIN
  FOR changed IN SELECT DISTINCT repo_id FROM changed_rows ORDER BY repo_id LOOP
    PERFORM public.gavel_git_repo_changed(changed);
  END LOOP;
  RETURN NULL;
END
$$;

DROP TRIGGER IF EXISTS git_branches_changed_insert ON public.git_branches;
CREATE TRIGGER git_branches_changed_insert
  AFTER INSERT ON public.git_branches
  REFERENCING NEW TABLE AS changed_rows
  FOR EACH STATEMENT EXECUTE FUNCTION public.gavel_git_rows_changed();

DROP TRIGGER IF EXISTS git_branches_changed_update ON public.git_branches;
CREATE TRIGGER git_branches_changed_update
  AFTER UPDATE ON public.git_branches
  REFERENCING NEW TABLE AS changed_rows
  FOR EACH STATEMENT EXECUTE FUNCTION public.gavel_git_rows_changed();

DROP TRIGGER IF EXISTS git_branches_changed_delete ON public.git_branches;
CREATE TRIGGER git_branches_changed_delete
  AFTER DELETE ON public.git_branches
  REFERENCING OLD TABLE AS changed_rows
  FOR EACH STATEMENT EXECUTE FUNCTION public.gavel_git_rows_changed();

DROP TRIGGER IF EXISTS git_range_stats_changed_insert ON public.git_range_stats;
CREATE TRIGGER git_range_stats_changed_insert
  AFTER INSERT ON public.git_range_stats
  REFERENCING NEW TABLE AS changed_rows
  FOR EACH STATEMENT EXECUTE FUNCTION public.gavel_git_rows_changed();

-- Recomputes every repo overview row from git_repos, assigning each a fresh
-- generation so a client holding a pre-crash one refreshes.
CREATE OR REPLACE FUNCTION public.gavel_git_repo_overview_rebuild()
RETURNS integer
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  affected integer := 0;
BEGIN
  INSERT INTO public.git_repo_overview AS overview (repo_id, refs_fingerprint, refs_scanned_at, generation)
  SELECT repo.id, CASE WHEN repo.error IS NULL THEN repo.refs_fingerprint END, repo.refs_scanned_at,
         nextval('public.git_repo_generation_seq')
  FROM public.git_repos repo
  ORDER BY repo.id
  ON CONFLICT (repo_id) DO UPDATE
    SET refs_fingerprint = EXCLUDED.refs_fingerprint,
        refs_scanned_at = GREATEST(overview.refs_scanned_at, EXCLUDED.refs_scanned_at),
        generation = EXCLUDED.generation;
  GET DIAGNOSTICS affected = ROW_COUNT;
  RETURN affected;
END
$$;

-- Rebuilds both git overviews when crash recovery left either empty while its
-- source is not, worktrees first, and announces every repo so listeners reload.
-- Reports whether it rebuilt.
CREATE OR REPLACE FUNCTION public.gavel_git_overview_ensure()
RETURNS boolean
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
  repo uuid;
BEGIN
  IF (EXISTS (SELECT 1 FROM public.git_worktree_overview) OR NOT EXISTS (SELECT 1 FROM public.git_worktrees))
     AND (EXISTS (SELECT 1 FROM public.git_repo_overview) OR NOT EXISTS (SELECT 1 FROM public.git_repos)) THEN
    RETURN false;
  END IF;
  PERFORM public.gavel_git_worktree_overview_rebuild();
  PERFORM public.gavel_git_repo_overview_rebuild();
  FOR repo IN SELECT id FROM public.git_repos ORDER BY id LOOP
    PERFORM pg_notify('gavel_git_change', repo::text);
  END LOOP;
  RETURN true;
END
$$;

SELECT public.gavel_git_repo_overview_rebuild();
