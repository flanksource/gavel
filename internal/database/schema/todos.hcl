table "todo_workspaces" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "repo_key" {
    null = true
    type = text
  }
  column "display_name" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  index "todo_workspaces_repo_key_key" {
    unique  = true
    columns = [column.repo_key]
  }

  check "todo_workspaces_repo_key_normalized" {
    expr = "repo_key IS NULL OR (repo_key <> '' AND repo_key = lower(btrim(repo_key)))"
  }
}

table "todo_workspace_paths" {
  schema = schema.public

  column "workspace_id" {
    null = false
    type = uuid
  }
  column "path" {
    null = false
    type = text
  }
  column "is_primary" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.workspace_id, column.path]
  }

  foreign_key "todo_workspace_paths_workspace_id_fkey" {
    columns     = [column.workspace_id]
    ref_columns = [table.todo_workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "todo_workspace_paths_path_key" {
    unique  = true
    columns = [column.path]
  }
  index "todo_workspace_paths_primary_key" {
    unique  = true
    columns = [column.workspace_id]
    where   = "is_primary"
  }

  check "todo_workspace_paths_path_normalized" {
    expr = "path <> '' AND path = btrim(path)"
  }
}

table "todo_issues" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "title" {
    null = false
    type = text
  }
  column "body" {
    null    = false
    type    = text
    default = ""
  }
  column "verification" {
    null    = false
    type    = text
    default = ""
  }
  column "labels" {
    null    = false
    type    = sql("text[]")
    default = sql("'{}'::text[]")
  }
  column "priority" {
    null    = false
    type    = text
    default = "medium"
  }
  column "status" {
    null    = false
    type    = text
    default = "open"
  }
  column "execution_state" {
    null    = false
    type    = text
    default = "idle"
  }
  column "active_prompt_run_id" {
    null = true
    type = uuid
  }
  column "selected_plan_id" {
    null = true
    type = uuid
  }
  column "version" {
    null    = false
    type    = bigint
    default = 0
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "todo_issues_workspace_id_fkey" {
    columns     = [column.workspace_id]
    ref_columns = [table.todo_workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "todo_issues_active_prompt_run_link_fkey" {
    columns     = [column.id, column.active_prompt_run_id]
    ref_columns = [table.todo_issue_prompt_runs.column.issue_id, table.todo_issue_prompt_runs.column.prompt_run_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "todo_issues_selected_plan_link_fkey" {
    columns     = [column.id, column.selected_plan_id]
    ref_columns = [table.todo_issue_plans.column.issue_id, table.todo_issue_plans.column.plan_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }

  index "todo_issues_id_workspace_id_key" {
    unique  = true
    columns = [column.id, column.workspace_id]
  }
  index "todo_issues_workspace_id_idx" {
    columns = [column.workspace_id]
  }
  index "todo_issues_status_idx" {
    columns = [column.workspace_id, column.status]
  }
  index "todo_issues_execution_state_idx" {
    columns = [column.workspace_id, column.execution_state]
  }
  index "todo_issues_labels_idx" {
    columns = [column.labels]
    type    = GIN
  }

  check "todo_issues_title_not_empty" {
    expr = "btrim(title) <> ''"
  }
  check "todo_issues_priority_check" {
    expr = "priority = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text])"
  }
  check "todo_issues_status_check" {
    expr = "status = ANY (ARRAY['draft'::text, 'open'::text, 'verified'::text, 'closed'::text, 'cancelled'::text])"
  }
  check "todo_issues_execution_state_check" {
    expr = "execution_state = ANY (ARRAY['idle'::text, 'planning'::text, 'running'::text, 'waiting'::text, 'stalled'::text, 'failed'::text, 'verifying'::text, 'verification_failed'::text])"
  }
  check "todo_issues_version_nonnegative" {
    expr = "version >= 0"
  }
}

table "todo_issue_aliases" {
  schema = schema.public

  column "workspace_id" {
    null = false
    type = uuid
  }
  column "alias" {
    null = false
    type = text
  }
  column "issue_id" {
    null = false
    type = uuid
  }
  column "kind" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.workspace_id, column.alias]
  }

  foreign_key "todo_issue_aliases_issue_workspace_fkey" {
    columns     = [column.issue_id, column.workspace_id]
    ref_columns = [table.todo_issues.column.id, table.todo_issues.column.workspace_id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "todo_issue_aliases_issue_id_idx" {
    columns = [column.issue_id]
  }

  check "todo_issue_aliases_alias_normalized" {
    expr = "alias <> '' AND alias = lower(btrim(alias))"
  }
  check "todo_issue_aliases_kind_not_empty" {
    expr = "kind IS NULL OR (kind <> '' AND kind = lower(btrim(kind)))"
  }
}

table "todo_issue_relationships" {
  schema = schema.public

  column "workspace_id" {
    null = false
    type = uuid
  }
  column "issue_id" {
    null = false
    type = uuid
  }
  column "target_issue_id" {
    null = false
    type = uuid
  }
  column "relation" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.workspace_id, column.issue_id, column.target_issue_id, column.relation]
  }

  foreign_key "todo_issue_relationships_workspace_id_fkey" {
    columns     = [column.workspace_id]
    ref_columns = [table.todo_workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "todo_issue_relationships_issue_workspace_fkey" {
    columns     = [column.issue_id, column.workspace_id]
    ref_columns = [table.todo_issues.column.id, table.todo_issues.column.workspace_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "todo_issue_relationships_target_workspace_fkey" {
    columns     = [column.target_issue_id, column.workspace_id]
    ref_columns = [table.todo_issues.column.id, table.todo_issues.column.workspace_id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }

  index "todo_issue_relationships_target_idx" {
    columns = [column.workspace_id, column.target_issue_id, column.relation]
  }

  check "todo_issue_relationships_relation_check" {
    expr = "relation = ANY (ARRAY['depends_on'::text, 'related_to'::text])"
  }
  check "todo_issue_relationships_not_self" {
    expr = "issue_id <> target_issue_id"
  }
  check "todo_issue_relationships_related_order" {
    expr = "relation <> 'related_to' OR issue_id < target_issue_id"
  }
}

table "todo_issue_events" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "issue_id" {
    null = false
    type = uuid
  }
  column "sequence" {
    null = false
    type = bigint
  }
  column "kind" {
    null = false
    type = text
  }
  column "actor" {
    null = true
    type = text
  }
  column "body" {
    null = true
    type = text
  }
  column "payload" {
    null    = false
    type    = jsonb
    default = sql("'{}'::jsonb")
  }
  column "source" {
    null    = false
    type    = text
    default = "gavel"
  }
  column "source_id" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "todo_issue_events_issue_id_fkey" {
    columns     = [column.issue_id]
    ref_columns = [table.todo_issues.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "todo_issue_events_issue_sequence_key" {
    unique  = true
    columns = [column.issue_id, column.sequence]
  }
  index "todo_issue_events_source_id_key" {
    unique  = true
    columns = [column.source, column.source_id]
    where   = "source_id IS NOT NULL"
  }
  index "todo_issue_events_created_at_idx" {
    columns = [column.issue_id, column.created_at]
  }

  check "todo_issue_events_sequence_positive" {
    expr = "sequence > 0"
  }
  check "todo_issue_events_kind_not_empty" {
    expr = "kind <> '' AND kind = lower(btrim(kind))"
  }
  check "todo_issue_events_source_not_empty" {
    expr = "source <> '' AND source = lower(btrim(source))"
  }
  check "todo_issue_events_source_id_not_empty" {
    expr = "source_id IS NULL OR btrim(source_id) <> ''"
  }
}

table "todo_issue_prompt_runs" {
  schema = schema.public

  column "issue_id" {
    null = false
    type = uuid
  }
  column "prompt_run_id" {
    null = false
    type = uuid
  }
  column "step_kind" {
    null = false
    type = text
  }
  column "ordinal" {
    null = false
    type = integer
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.issue_id, column.prompt_run_id]
  }

  foreign_key "todo_issue_prompt_runs_issue_id_fkey" {
    columns     = [column.issue_id]
    ref_columns = [table.todo_issues.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "todo_issue_prompt_runs_prompt_run_id_key" {
    unique  = true
    columns = [column.prompt_run_id]
  }
  index "todo_issue_prompt_runs_step_ordinal_key" {
    unique  = true
    columns = [column.issue_id, column.step_kind, column.ordinal]
  }

  check "todo_issue_prompt_runs_step_kind_check" {
    expr = "step_kind = ANY (ARRAY['plan'::text, 'run'::text, 'verify'::text])"
  }
  check "todo_issue_prompt_runs_ordinal_nonnegative" {
    expr = "ordinal >= 0"
  }
}

table "todo_issue_plans" {
  schema = schema.public

  column "issue_id" {
    null = false
    type = uuid
  }
  column "plan_id" {
    null = false
    type = uuid
  }
  column "ordinal" {
    null = false
    type = integer
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key {
    columns = [column.issue_id, column.plan_id]
  }

  foreign_key "todo_issue_plans_issue_id_fkey" {
    columns     = [column.issue_id]
    ref_columns = [table.todo_issues.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "todo_issue_plans_ordinal_key" {
    unique  = true
    columns = [column.issue_id, column.ordinal]
  }

  check "todo_issue_plans_ordinal_nonnegative" {
    expr = "ordinal >= 0"
  }
}
