-- phase: pre

-- comment_agent_sessions was keyed by delivery group (tenant, scope, group);
-- it is now one row per managed session (tenant, scope, session). A session
-- recorded for two groups of one scope would break the new key, so keep only
-- its newest group before Atlas rebuilds the primary key.
DO $$
BEGIN
  IF to_regclass('public.comment_agent_sessions') IS NOT NULL THEN
    DELETE FROM public.comment_agent_sessions older
    USING public.comment_agent_sessions newer
    WHERE older.tenant_id = newer.tenant_id
      AND older.scope_uri = newer.scope_uri
      AND older.session_id = newer.session_id
      AND (older.recorded_at, older.group_key) < (newer.recorded_at, newer.group_key);
  END IF;
END $$;
