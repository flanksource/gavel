table "http_cache_entries" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "http_cache_entries_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "http_cache_entries_owner_idx" { columns = [column.org_id, column.user_id] }

  column "url" {
    null = false
    type = varchar(1024)
  }
  column "method" {
    null = false
    type = varchar(8)
  }
  column "e_tag" {
    null = true
    type = varchar(256)
  }
  column "last_modified" {
    null = true
    type = varchar(64)
  }
  column "status_code" {
    null = true
    type = bigint
  }
  column "body" {
    null = true
    type = bytea
  }
  column "headers" {
    null = true
    type = bytea
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }
  column "expires_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.url, column.method] }
  index "idx_http_cache_entries_fetched_at" { columns = [column.fetched_at] }
}

table "workflow_run_caches" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "workflow_run_caches_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "workflow_run_caches_owner_idx" { columns = [column.org_id, column.user_id] }

  column "run_id" {
    null = false
    type = bigint
  }
  column "repo" {
    null = true
    type = varchar(255)
  }
  column "status" {
    null = true
    type = varchar(32)
  }
  column "conclusion" {
    null = true
    type = varchar(32)
  }
  column "payload" {
    null = true
    type = bytea
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.run_id] }
  index "idx_workflow_run_caches_repo" { columns = [column.repo] }
}

table "job_log_caches" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "job_log_caches_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "job_log_caches_owner_idx" { columns = [column.org_id, column.user_id] }

  column "job_id" {
    null = false
    type = bigint
  }
  column "repo" {
    null = true
    type = varchar(255)
  }
  # Inline logs from before job logs moved to the blob store; GetJobLogs moves a row's logs
  # to the store on first read.
  column "logs_gz" {
    null = true
    type = bytea
  }
  # Key of the gzipped logs in the organization's blob store; empty for a row whose logs are
  # still inline.
  column "log_key" {
    null    = false
    type    = text
    default = ""
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.job_id] }
}

table "workflow_def_caches" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "workflow_def_caches_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "workflow_def_caches_owner_idx" { columns = [column.org_id, column.user_id] }

  column "repo" {
    null = false
    type = varchar(255)
  }
  column "workflow_id" {
    null = false
    type = bigint
  }
  column "sha" {
    null = false
    type = varchar(64)
  }
  column "path" {
    null = true
    type = varchar(512)
  }
  column "yaml" {
    null = true
    type = text
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.repo, column.workflow_id, column.sha] }
}

table "seen_prs" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "seen_prs_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "seen_prs_owner_idx" { columns = [column.org_id, column.user_id] }

  column "repo" {
    null = false
    type = varchar(255)
  }
  column "number" {
    null = false
    type = bigint
  }
  column "seen_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.repo, column.number] }
}

# The PR dashboard poller's results, persisted so triage and the TODO pages can
# search PRs without a live host call. Repo keeps the host's display spelling;
# lookups compare it case-insensitively against todo_pr_links' lower-cased key.
table "pull_requests" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "pull_requests_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "pull_requests_owner_idx" { columns = [column.org_id, column.user_id] }

  column "provider" {
    null = false
    type = text
  }
  column "repo" {
    null = false
    type = varchar(255)
  }
  column "number" {
    null = false
    type = bigint
  }
  column "url" {
    null = false
    type = varchar(1024)
  }
  column "title" {
    null = false
    type = text
  }
  column "author" {
    null = true
    type = varchar(255)
  }
  column "state" {
    null = true
    type = varchar(32)
  }
  column "is_draft" {
    null    = false
    type    = boolean
    default = false
  }
  column "head" {
    null = true
    type = varchar(255)
  }
  column "base" {
    null = true
    type = varchar(255)
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "synced_at" {
    null = false
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.provider, column.repo, column.number] }
  index "idx_pull_requests_updated_at" { columns = [column.org_id, column.updated_at] }
  index "idx_pull_requests_title_search" {
    type = GIN
    on {
      expr = "to_tsvector('simple'::regconfig, title)"
    }
  }

  check "pull_requests_provider_check" {
    expr = "provider = ANY (ARRAY['github'::text, 'azure'::text])"
  }
}

table "favicon_caches" {
  schema = schema.public

  column "homepage" {
    null = false
    type = varchar(1024)
  }
  column "icon_url" {
    null = true
    type = varchar(1024)
  }
  column "data" {
    null = true
    type = bytea
  }
  column "mime_type" {
    null = true
    type = varchar(64)
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }
  column "expires_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.homepage] }
  index "idx_favicon_caches_fetched_at" { columns = [column.fetched_at] }
  index "idx_favicon_caches_expires_at" { columns = [column.expires_at] }
}

table "test_run_caches" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "test_run_caches_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "test_run_caches_owner_idx" { columns = [column.org_id, column.user_id] }

  column "workspace_dir" {
    null = false
    type = varchar(512)
  }
  column "run_id" {
    null = false
    type = varchar(255)
  }
  column "path" {
    null = true
    type = varchar(1024)
  }
  column "kind" {
    null = true
    type = varchar(16)
  }
  column "repo" {
    null = true
    type = varchar(255)
  }
  column "sha" {
    null = true
    type = varchar(64)
  }
  column "started_ts" {
    null = true
    type = bigint
  }
  column "ended_ts" {
    null = true
    type = bigint
  }
  column "passed" {
    null = true
    type = bigint
  }
  column "failed" {
    null = true
    type = bigint
  }
  column "skipped" {
    null = true
    type = bigint
  }
  column "warned" {
    null = true
    type = bigint
  }
  column "total" {
    null = true
    type = bigint
  }
  column "lint_violations" {
    null = true
    type = bigint
  }
  column "lint_linters" {
    null = true
    type = bigint
  }
  column "frameworks" {
    null = true
    type = bytea
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.workspace_dir, column.run_id] }
  index "idx_test_run_caches_kind" { columns = [column.kind] }
  index "idx_test_run_caches_started_ts" { columns = [column.started_ts] }
}

table "test_run_cursors" {
  schema = schema.public

  column "user_id" {
    null = true
    type = uuid
  }
  column "org_id" {
    null    = false
    type    = uuid
    default = sql("'00000000-0000-0000-0000-000000000001'::uuid")
  }
  foreign_key "test_run_cursors_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "test_run_cursors_owner_idx" { columns = [column.org_id, column.user_id] }

  column "workspace_dir" {
    null = false
    type = varchar(512)
  }
  column "last_run_ts" {
    null = true
    type = bigint
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.org_id, column.workspace_dir] }
}
