-- phase: post

-- The local organization owns every row written without authentication.
-- Tenant tables default org_id to it, so it must exist before their FKs.
INSERT INTO public.orgs (id, issuer, clerk_org_id, slug, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'local', 'local', 'local', 'Local')
ON CONFLICT (id) DO NOTHING;

-- Installed here rather than in HCL: Atlas adds a column and its FK in one
-- ALTER TABLE, which fails on existing rows before the local org exists.
DO $$
DECLARE
  tbl text;
BEGIN
  FOREACH tbl IN ARRAY ARRAY[
    'todo_workspaces', 'todo_labels',
    'session_owners', 'api_token_owners', 'runtime_profile_owners', 'runtime_preset_owners', 'git_agent_task_owners',
    'http_cache_entries', 'workflow_run_caches', 'job_log_caches', 'workflow_def_caches', 'seen_prs', 'pull_requests',
    'test_run_caches', 'test_run_cursors',
    'file_scans', 'violations', 'linter_executions', 'debounce_metadata', 'task_run_history',
    'git_repos'
  ]
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM pg_constraint
      WHERE conrelid = format('public.%I', tbl)::regclass
        AND conname = tbl || '_org_id_fkey'
    ) THEN
      EXECUTE format(
        'ALTER TABLE public.%I ADD CONSTRAINT %I FOREIGN KEY (org_id) REFERENCES public.orgs (id) ON UPDATE NO ACTION ON DELETE RESTRICT',
        tbl, tbl || '_org_id_fkey'
      );
    END IF;
  END LOOP;
END
$$;
