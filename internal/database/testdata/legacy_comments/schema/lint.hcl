table "comment_lint_runs" {
 schema = schema.public
 column "tenant_id" { type = text }
 column "id" { type = text }
 column "request_id" { type = text }
 column "payload" { type = jsonb }
 primary_key { columns = [column.tenant_id,column.id] }
 index "comment_lint_runs_request" {
  unique = true
  columns = [column.tenant_id,column.request_id]
 }
}
