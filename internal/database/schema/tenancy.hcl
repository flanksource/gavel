# Organization-scoped tables for hosted mode: memberships, SSH keys, repositories,
# GitHub App installations, per-org settings and sandboxed runs.
table "org_memberships" {
  schema = schema.public
  column "user_id" {
    null = false
    type = uuid
  }
  column "org_id" {
    null = false
    type = uuid
  }
  column "role" {
    null = false
    type = text
  }
  column "synced_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.user_id, column.org_id] }
  foreign_key "org_memberships_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
  foreign_key "org_memberships_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
  index "org_memberships_org_id_idx" { columns = [column.org_id] }
}

# SSH keys belong to users, not orgs: one key authenticates its user for every org they're in.
table "user_ssh_keys" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = text
    default = ""
  }
  # SHA256 form, as produced by gossh.FingerprintSHA256.
  column "fingerprint" {
    null = false
    type = text
  }
  # The authorized_keys line.
  column "public_key" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  column "last_used_at" {
    null = true
    type = timestamptz
  }
  primary_key { columns = [column.id] }
  foreign_key "user_ssh_keys_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
  index "user_ssh_keys_fingerprint_key" {
    unique = true
    columns = [column.fingerprint]
  }
  index "user_ssh_keys_user_id_idx" { columns = [column.user_id] }
}

# A GitHub App installation is linked to exactly one gavel org.
table "github_installations" {
  schema = schema.public
  # GitHub's installation id.
  column "id" {
    null = false
    type = bigint
  }
  column "org_id" {
    null = false
    type = uuid
  }
  column "account_login" {
    null = false
    type = text
  }
  column "account_type" {
    null = false
    type = text
  }
  column "created_by_user_id" {
    null = true
    type = uuid
  }
  column "created_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  column "suspended_at" {
    null = true
    type = timestamptz
  }
  primary_key { columns = [column.id] }
  foreign_key "github_installations_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = RESTRICT
    on_update = NO_ACTION
  }
  foreign_key "github_installations_created_by_user_id_fkey" {
    columns = [column.created_by_user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "github_installations_org_id_idx" { columns = [column.org_id] }
}

# Repositories an org can run against: pushed over SSH, or reached through a GitHub App installation.
table "repos" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "org_id" {
    null = false
    type = uuid
  }
  # <repo> for pushed repositories, <owner>/<repo> for GitHub.
  column "name" {
    null = false
    type = text
  }
  column "source" {
    null = false
    type = text
  }
  column "github_repo_id" {
    null = true
    type = bigint
  }
  column "installation_id" {
    null = true
    type = bigint
  }
  column "default_branch" {
    null = false
    type = text
    default = "main"
  }
  column "created_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "repos_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = RESTRICT
    on_update = NO_ACTION
  }
  foreign_key "repos_installation_id_fkey" {
    columns = [column.installation_id]
    ref_columns = [table.github_installations.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "repos_org_source_name_key" {
    unique = true
    columns = [column.org_id, column.source, column.name]
  }
  index "repos_github_repo_id_key" {
    unique = true
    columns = [column.github_repo_id]
    where = "github_repo_id IS NOT NULL"
  }
  check "repos_source_valid" { expr = "source IN ('push', 'github')" }
}

# Per-org dashboard settings; replaces ~/.config/gavel/ui.settings.json in hosted mode.
table "org_settings" {
  schema = schema.public
  column "org_id" {
    null = false
    type = uuid
  }
  # The ui.SearchConfig JSON.
  column "search_config" {
    null = false
    type = jsonb
    default = sql("'{}'::jsonb")
  }
  column "updated_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.org_id] }
  foreign_key "org_settings_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = CASCADE
    on_update = NO_ACTION
  }
}

# Runs executed in a Kubernetes Job or container. The run reports back through the run API.
table "sandbox_runs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "org_id" {
    null = false
    type = uuid
  }
  column "user_id" {
    null = true
    type = uuid
  }
  column "repo_id" {
    null = true
    type = uuid
  }
  column "kind" {
    null = false
    type = text
  }
  column "ref" {
    null = false
    type = text
    default = ""
  }
  column "sha" {
    null = false
    type = text
    default = ""
  }
  column "todo_id" {
    null = true
    type = uuid
  }
  column "status" {
    null = false
    type = text
  }
  column "exit_code" {
    null = true
    type = integer
  }
  column "backend" {
    null = false
    type = text
  }
  # Job name or container id.
  column "backend_ref" {
    null = false
    type = text
    default = ""
  }
  # Blob key of the run log.
  column "log_key" {
    null = false
    type = text
    default = ""
  }
  # The RunSpec the run API serves to the job (sandbox.RunSpec).
  column "spec" {
    null    = false
    type    = jsonb
    default = sql("'{}'::jsonb")
  }
  column "created_at" {
    null = false
    type = timestamptz
    default = sql("now()")
  }
  column "started_at" {
    null = true
    type = timestamptz
  }
  column "finished_at" {
    null = true
    type = timestamptz
  }
  primary_key { columns = [column.id] }
  foreign_key "sandbox_runs_org_id_fkey" {
    columns = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete = RESTRICT
    on_update = NO_ACTION
  }
  foreign_key "sandbox_runs_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  foreign_key "sandbox_runs_repo_id_fkey" {
    columns = [column.repo_id]
    ref_columns = [table.repos.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "sandbox_runs_org_created_idx" {
    on {
      column = column.org_id
    }
    on {
      column = column.created_at
      desc   = true
    }
  }
  check "sandbox_runs_kind_valid" { expr = "kind IN ('test', 'lint', 'push-check', 'agent')" }
  check "sandbox_runs_status_valid" { expr = "status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled')" }
}
