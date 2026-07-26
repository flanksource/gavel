table "task_run_history" {
  schema = schema.public

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
  index "idx_task_run_history_started_at" { columns = [column.started_at] }
  index "idx_task_run_history_expires_at" { columns = [column.expires_at] }
}
