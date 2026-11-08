schema "public" {}
table "comment_operations" {
 schema = schema.public
 column "id" { type = text }
 column "tenant_id" { type = text }
 column "comment_id" { type = text }
 column "request_id" { type = text }
 column "state" { type = text }
 column "payload" { type = jsonb }
 column "lease" {
  type = text
  null = true
 }
 column "lease_until" {
  type = timestamptz
  null = true
 }
 primary_key { columns = [column.tenant_id, column.id] }
 index "comment_operations_request" {
  unique = true
  columns = [column.tenant_id, column.request_id]
 }
 index "comment_operations_active" {
  unique = true
  columns = [column.tenant_id, column.comment_id]
  where = "state IN ('pending','inflight')"
 }
 foreign_key "comment_operations_root" {
  columns = [column.tenant_id, column.comment_id]
  ref_columns = [table.comments.column.tenant_id, table.comments.column.id]
 }
 check "comment_operations_state" { expr = "state IN ('pending','inflight','done','failed','unknown')" }
}
table "comments" {
 schema = schema.public
 column "id" { type = text }
 column "tenant_id" { type = text }
 column "scope_uri" { type = text }
 column "target_uri" { type = text }
 column "phase" { type = text }
 column "version" { type = bigint }
 column "request_id" { type = text }
 column "request_hash" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id, column.id] }
 index "comments_request_key" {
  unique = true
  columns = [column.tenant_id, column.request_id]
 }
 index "comments_scope_phase" { columns = [column.tenant_id, column.scope_uri, column.phase, column.id] }
 index "comments_target" { columns = [column.tenant_id, column.target_uri] }
 check "comments_version_positive" { expr = "version > 0" }
 check "comments_outcome_phase" { expr = "(phase = 'final') = (payload->'outcome' IS NOT NULL AND payload->'outcome' <> 'null'::jsonb)" }
 check "comments_phase" { expr = "phase IN ('active', 'awaiting_outcome', 'final')" }
}
table "comment_sequences" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "sequence" { type = bigint }
 primary_key { columns = [column.tenant_id] }
}
table "comment_messages" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "comment_id" { type = text }
 column "id" { type = text }
 column "sequence" { type = bigint }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id, column.id] }
 index "comment_messages_order" {
  unique = true
  columns = [column.tenant_id, column.comment_id, column.sequence]
 }
 foreign_key "comment_messages_root" {
  columns = [column.tenant_id, column.comment_id]
  ref_columns = [table.comments.column.tenant_id, table.comments.column.id]
 }
}
table "comment_events" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "sequence" { type = bigint }
 column "comment_id" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id, column.sequence] }
 foreign_key "comment_events_root" {
  columns = [column.tenant_id, column.comment_id]
  ref_columns = [table.comments.column.tenant_id, table.comments.column.id]
 }
}
