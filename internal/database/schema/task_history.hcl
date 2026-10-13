table "task_run_history" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "task_run_history_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "task_run_history_owner_idx" { columns = [column.org_id, column.user_id] }

  column "id" {
    null = false
    type = text
  }
  column "started_at" {
    null = false
    type = timestamptz
  }
  column "run" {
    null = false
    type = jsonb
  }
  column "snapshots" {
    null = false
    type = jsonb
  }
  column "archived_at" {
    null = false
    type = timestamptz
  }
  column "expires_at" {
    null = false
    type = timestamptz
  }

  primary_key { columns = [column.id] }
  # There is no started_at index. List sorts by started_at, but it does so after
  # an expires_at filter over a table that holds a few dozen live rows, so the
  # planner sorts the whole set instead -- 0 scans measured against 147,915 on
  # the expires_at index it does use.
  index "idx_task_run_history_expires_at" { columns = [column.expires_at] }
}
