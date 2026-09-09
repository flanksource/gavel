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
  # There is no started_at index. List sorts by started_at, but it does so after
  # an expires_at filter over a table that holds a few dozen live rows, so the
  # planner sorts the whole set instead -- 0 scans measured against 147,915 on
  # the expires_at index it does use.
  index "idx_task_run_history_expires_at" { columns = [column.expires_at] }
}
