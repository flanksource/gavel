table "comment_reads" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "actor_kind" { type = text }
 column "actor_id" { type = text }
 column "comment_id" { type = text }
 column "version" { type = bigint }
 column "read_at" { type = timestamptz }
 primary_key { columns = [column.tenant_id,column.actor_kind,column.actor_id,column.comment_id] }
 foreign_key "comment_reads_root" {
  columns = [column.tenant_id,column.comment_id]
  ref_columns = [table.comments.column.tenant_id,table.comments.column.id]
 }
}
table "comment_attachments" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "comment_id" { type = text }
 column "id" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id,column.id] }
 foreign_key "comment_attachments_root" {
  columns = [column.tenant_id,column.comment_id]
  ref_columns = [table.comments.column.tenant_id,table.comments.column.id]
 }
}
table "comment_blobs" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "id" { type = text }
 column "data" { type = bytea }
 primary_key { columns = [column.tenant_id,column.id] }
}
