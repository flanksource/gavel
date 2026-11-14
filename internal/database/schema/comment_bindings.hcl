table "comment_bindings" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "comment_id" { type = text }
 column "provider" { type = text }
 column "remote_key" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id, column.provider, column.remote_key] }
 index "comment_bindings_local" {
  unique = true
  columns = [column.tenant_id, column.comment_id, column.provider]
 }
 foreign_key "comment_binding_root" {
  columns = [column.tenant_id, column.comment_id]
  ref_columns = [table.comments.column.tenant_id, table.comments.column.id]
 }
}
