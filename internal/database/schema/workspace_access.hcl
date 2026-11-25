table "workspace_access_grants" {
  schema = schema.public
  column "id" { type = uuid }
  column "machine_id" { type = uuid }
  column "user_id" { type = text }
  column "path" { type = text }
  column "created_at" {
    type = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  index "workspace_access_grants_scope_path_key" {
    unique = true
    columns = [column.machine_id, column.user_id, column.path]
  }
}
