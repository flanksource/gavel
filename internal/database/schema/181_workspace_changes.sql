-- phase: post

CREATE OR REPLACE FUNCTION public.gavel_notify_workspace_change()
RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE row_data jsonb; workspace uuid;
BEGIN
  row_data := CASE WHEN TG_OP = 'DELETE' THEN to_jsonb(OLD) ELSE to_jsonb(NEW) END;
  IF TG_TABLE_NAME = 'todo_workspaces' THEN
    workspace := (row_data->>'id')::uuid;
  ELSIF row_data ? 'workspace_id' THEN
    workspace := (row_data->>'workspace_id')::uuid;
  ELSIF row_data ? 'issue_id' THEN
    SELECT workspace_id INTO workspace FROM todo_issues WHERE id = (row_data->>'issue_id')::uuid;
  ELSIF TG_TABLE_NAME = 'todo_run_landings' THEN
    SELECT i.workspace_id INTO workspace FROM todo_issue_prompt_runs l JOIN todo_issues i ON i.id=l.issue_id WHERE l.prompt_run_id=(row_data->>'prompt_run_id')::uuid LIMIT 1;
  END IF;
  PERFORM pg_notify('gavel_workspace_change', json_build_object(
    'catalog', TG_TABLE_NAME IN ('todo_workspaces', 'repos'),
    'workspaceId', workspace
  )::text);
  RETURN NULL;
END;
$$;

DO $$
DECLARE target text;
BEGIN
  FOREACH target IN ARRAY ARRAY['todo_issues','todo_issue_events','todo_issue_aliases','todo_issue_prompt_runs','todo_issue_plans','todo_run_landings','todo_labels','todo_workspaces','repos'] LOOP
    IF to_regclass('public.' || target) IS NOT NULL THEN
      EXECUTE format('DROP TRIGGER IF EXISTS gavel_workspace_change ON public.%I', target);
      EXECUTE format('CREATE TRIGGER gavel_workspace_change AFTER INSERT OR UPDATE OR DELETE ON public.%I FOR EACH ROW EXECUTE FUNCTION public.gavel_notify_workspace_change()', target);
    END IF;
  END LOOP;
END;
$$;
