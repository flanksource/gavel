table "file_scans" {
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
  foreign_key "file_scans_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "file_scans_owner_idx" { columns = [column.org_id, column.user_id] }

  column "file_path" {
    null = false
    type = text
  }
  column "last_scan_time" {
    null = false
    type = bigint
  }
  column "file_mod_time" {
    null = false
    type = bigint
  }
  column "file_hash" {
    null = false
    type = text
  }

  primary_key { columns = [column.org_id, column.file_path] }
}

table "violations" {
  schema = schema.public

  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  column "id" {
    null = false
    type = bigserial
  }
  column "file_path" {
    null = false
    type = text
  }
  column "line" {
    null = false
    type = bigint
  }
  column "column" {
    null = false
    type = bigint
  }
  column "message" {
    null = true
    type = text
  }
  column "source" {
    null = false
    type = text
  }
  column "rule" {
    null = true
    type = jsonb
  }
  column "severity" {
    null = false
    type = text
  }
  column "fixable" {
    null = false
    type = boolean
  }
  column "fix_applicability" {
    null = false
    type = text
  }
  column "code" {
    null = true
    type = text
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key { columns = [column.id] }
  index "idx_violations_file_path" { columns = [column.file_path] }
  index "idx_violations_source" { columns = [column.source] }
  index "idx_violations_created_at" { columns = [column.created_at] }

  foreign_key "violations_file_scan_fkey" {
    columns     = [column.org_id, column.file_path]
    ref_columns = [table.file_scans.column.org_id, table.file_scans.column.file_path]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "linter_executions" {
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
  foreign_key "linter_executions_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "linter_executions_owner_idx" { columns = [column.org_id, column.user_id] }

  column "id" {
    null = false
    type = bigserial
  }
  column "linter_name" {
    null = false
    type = text
  }
  column "work_dir" {
    null = false
    type = text
  }
  column "executed_at" {
    null = false
    type = timestamptz
  }
  column "duration_ms" {
    null = false
    type = bigint
  }
  column "violation_count" {
    null = false
    type = bigint
  }
  column "success" {
    null = false
    type = boolean
  }

  primary_key { columns = [column.id] }
  index "linter_executions_unique" {
    unique  = true
    columns = [column.org_id, column.linter_name, column.work_dir, column.executed_at]
  }
  index "idx_linter_workdir" { columns = [column.linter_name, column.work_dir] }
  index "idx_executed_at" { columns = [column.executed_at] }
}

table "debounce_metadata" {
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
  foreign_key "debounce_metadata_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "debounce_metadata_owner_idx" { columns = [column.org_id, column.user_id] }

  column "linter_name" {
    null = false
    type = text
  }
  column "work_dir" {
    null = false
    type = text
  }
  column "last_debounce_used_ms" {
    null = true
    type = bigint
  }
  column "consecutive_no_violations" {
    null    = false
    type    = bigint
    default = 0
  }
  column "consecutive_violations" {
    null    = false
    type    = bigint
    default = 0
  }
  column "adaptation_factor" {
    null    = false
    type    = double_precision
    default = 1.0
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key { columns = [column.org_id, column.linter_name, column.work_dir] }
}
