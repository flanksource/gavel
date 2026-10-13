# A git-agent task exists before its prompt run, so it needs an independent
# ownership link. Attempts inherit through captain_git_agent_tasks.id.
table "git_agent_task_owners" {
  schema = schema.public
  column "task_id" {
    null = false
    type = uuid
  }
  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  primary_key { columns = [column.task_id] }
  foreign_key "git_agent_task_owners_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "git_agent_task_owners_user_id_idx" { columns = [column.user_id] }
  index "git_agent_task_owners_org_id_idx" { columns = [column.org_id] }
}
