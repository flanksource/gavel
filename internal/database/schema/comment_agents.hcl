// One managed session a comments host launched for a scope's threads, with
// what it runs so it is listed after it ends and resumed with the same prompt.
// recorded_at is when it started; group_key is "" for a session of no group,
// else a group's session is its newest. A row without a prompt predates
// recording what sessions run and only names its group's session.
table "comment_agent_sessions" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "scope_uri" { type = text }
 column "group_key" { type = text }
 column "session_id" { type = text }
 column "recorded_at" { type = timestamptz }
 column "prompt" {
  type = text
  null = true
 }
 column "run_options" {
  type = jsonb
  null = true
 }
 column "label" {
  type = text
  null = true
 }
 column "cwd" {
  type = text
  null = true
 }
 column "source" {
  type = text
  null = true
 }
 column "model" {
  type = text
  null = true
 }
 column "last_delivered_at" {
  type = timestamptz
  null = true
 }
 primary_key { columns = [column.tenant_id,column.scope_uri,column.session_id] }
 index "comment_agent_sessions_group" { columns = [column.tenant_id,column.scope_uri,column.group_key,column.recorded_at] }
}
