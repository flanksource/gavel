table "connections" {
  schema = schema.public
  column "id" { type = uuid }
  column "org_id" { type = uuid }
  column "namespace" { type = text }
  column "name" { type = text }
  column "type" { type = text }
  column "key_id" { type = text }
  column "payload" { type = bytea }
  column "created_by" {
    type = uuid
    null = true
  }
  column "created_at" {
    type = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  index "connections_identity" {
    unique = true
    columns = [column.org_id, column.namespace, column.name]
  }
  foreign_key "connections_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = RESTRICT
  }
}
