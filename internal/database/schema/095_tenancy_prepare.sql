-- phase: pre

-- violations references file_scans by file_path alone. Atlas rebuilds the
-- file_scans primary key as (org_id, file_path), which PostgreSQL refuses while
-- this FK depends on it; the schema recreates it as violations_file_scan_fkey.
DO $$
BEGIN
  -- to_regclass returns NULL for a missing table; a ::regclass cast would raise
  -- on a fresh database even behind an IS NOT NULL guard, since it's folded at plan time.
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conrelid = to_regclass('public.violations')
      AND conname = 'violations_file_path_fkey'
  ) THEN
    ALTER TABLE public.violations DROP CONSTRAINT violations_file_path_fkey;
  END IF;
END
$$;
