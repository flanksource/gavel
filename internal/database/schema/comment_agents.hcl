table "comment_routes" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "id" { type = text }
 column "scope_uri" { type = text }
 column "target_prefix" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id,column.id] }
 index "comment_routes_prefix" {
  unique = true
  columns = [column.tenant_id,column.scope_uri,column.target_prefix]
 }
}
table "comment_agent_sessions" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "scope_uri" { type = text }
 column "group_key" { type = text }
 column "session_id" { type = text }
 column "recorded_at" { type = timestamptz }
 primary_key { columns = [column.tenant_id,column.scope_uri,column.group_key] }
}
