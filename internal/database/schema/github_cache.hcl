table "http_cache_entries" {
  schema = schema.public

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

  primary_key { columns = [column.url, column.method] }
  index "idx_http_cache_entries_fetched_at" { columns = [column.fetched_at] }
}

table "workflow_run_caches" {
  schema = schema.public

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

  primary_key { columns = [column.run_id] }
  index "idx_workflow_run_caches_repo" { columns = [column.repo] }
}

table "job_log_caches" {
  schema = schema.public

  column "job_id" {
    null = false
    type = bigint
  }
  column "repo" {
    null = true
    type = varchar(255)
  }
  column "logs_gz" {
    null = true
    type = bytea
  }
  column "fetched_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.job_id] }
}

table "workflow_def_caches" {
  schema = schema.public

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

  primary_key { columns = [column.repo, column.workflow_id, column.sha] }
}

table "seen_prs" {
  schema = schema.public

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

  primary_key { columns = [column.repo, column.number] }
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

table "grite_issue_caches" {
  schema = schema.public

  column "repo" {
    null = false
    type = varchar(512)
  }
  column "issue_id" {
    null = false
    type = varchar(64)
  }
  column "title" {
    null = true
    type = text
  }
  column "state" {
    null = true
    type = varchar(16)
  }
  column "labels" {
    null = true
    type = bytea
  }
  column "events_json" {
    null = true
    type = bytea
  }
  column "created_ts" {
    null = true
    type = bigint
  }
  column "updated_ts" {
    null = true
    type = bigint
  }
  column "comment_count" {
    null = true
    type = bigint
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo, column.issue_id] }
  index "idx_grite_issue_caches_state" { columns = [column.state] }
}

table "grite_sync_cursors" {
  schema = schema.public

  column "repo" {
    null = false
    type = varchar(512)
  }
  column "last_event_ts" {
    null = true
    type = bigint
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo] }
}

table "commit_stat_caches" {
  schema = schema.public

  column "repo" {
    null = false
    type = varchar(512)
  }
  column "issue_id" {
    null = false
    type = varchar(128)
  }
  column "commits" {
    null = true
    type = bigint
  }
  column "files" {
    null = true
    type = bigint
  }
  column "adds" {
    null = true
    type = bigint
  }
  column "dels" {
    null = true
    type = bigint
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo, column.issue_id] }
}

table "commit_stat_cursors" {
  schema = schema.public

  column "repo" {
    null = false
    type = varchar(512)
  }
  column "synced_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo] }
}

table "test_run_caches" {
  schema = schema.public

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

  primary_key { columns = [column.workspace_dir, column.run_id] }
  index "idx_test_run_caches_kind" { columns = [column.kind] }
  index "idx_test_run_caches_started_ts" { columns = [column.started_ts] }
}

table "test_run_cursors" {
  schema = schema.public

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

  primary_key { columns = [column.workspace_dir] }
}
