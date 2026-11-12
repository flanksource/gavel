table "comment_agent_sessions" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "scope_uri" { type = text }
 column "group_key" { type = text }
 column "session_id" { type = text }
 column "recorded_at" { type = timestamptz }
 primary_key { columns = [column.tenant_id,column.scope_uri,column.group_key] }
}
