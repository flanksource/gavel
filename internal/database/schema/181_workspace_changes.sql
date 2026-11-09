-- phase: post

CREATE OR REPLACE FUNCTION public.gavel_notify_workspace_change()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  PERFORM pg_notify('gavel_workspace_change', json_build_object(
    'catalog', TG_TABLE_NAME IN ('todo_workspaces', 'repos'),
    'table', TG_TABLE_NAME
  )::text);
  RETURN NULL;
END;
$$;

DO $$
DECLARE target text;
BEGIN
  FOREACH target IN ARRAY ARRAY['todo_issues','todo_issue_events','todo_issue_prompt_runs','todo_issue_plans','todo_labels','todo_workspaces','repos'] LOOP
    IF to_regclass('public.' || target) IS NOT NULL THEN
      EXECUTE format('DROP TRIGGER IF EXISTS gavel_workspace_change ON public.%I', target);
      EXECUTE format('CREATE TRIGGER gavel_workspace_change AFTER INSERT OR UPDATE OR DELETE ON public.%I FOR EACH STATEMENT EXECUTE FUNCTION public.gavel_notify_workspace_change()', target);
    END IF;
  END LOOP;
END;
$$;
