// Git state of the dashboard's project repositories, kept current by the git
// state tracker (git/gitstate) so navigation reads never run git. Ref-level
// rows (repos, worktrees, branches) are rewritten when a cheap fingerprint
// changes; git_range_stats rows are pure functions of two commits and are
// never updated once written.

table "git_repos" {
  schema = schema.public

  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  // Realpath of the primary checkout.
  column "root_dir" {
    null = false
    type = text
  }
  column "common_dir" {
    null = false
    type = text
  }
  column "base_branch" {
    null = true
    type = text
  }
  column "base_sha" {
    null = true
    type = text
  }
  column "current_branch" {
    null = true
    type = text
  }
  column "base_checked_out" {
    null    = false
    type    = boolean
    default = false
  }
  // Hash of `git worktree list --porcelain` + `git for-each-ref refs/heads`
  // + the base; a ref scan with an unchanged fingerprint writes nothing else.
  column "refs_fingerprint" {
    null = true
    type = text
  }
  column "refs_scanned_at" {
    null = true
    type = timestamptz
  }
  // Advances on every change to the repo's rows; pushed to the UI.
  column "generation" {
    null    = false
    type    = bigint
    default = 0
  }
  column "error" {
    null = true
    type = text
  }
  column "error_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key { columns = [column.id] }
  index "git_repos_root_dir_key" {
    unique  = true
    columns = [column.root_dir]
  }
}

table "git_worktrees" {
  schema = schema.public

  column "repo_id" {
    null = false
    type = uuid
  }
  column "path" {
    null = false
    type = text
  }
  column "branch" {
    null = true
    type = text
  }
  column "head_sha" {
    null = true
    type = text
  }
  column "is_primary" {
    null    = false
    type    = boolean
    default = false
  }
  column "detached" {
    null    = false
    type    = boolean
    default = false
  }
  column "prunable" {
    null    = false
    type    = boolean
    default = false
  }
  // Position in `git worktree list`; the primary checkout is 0.
  column "ordinal" {
    null    = false
    type    = integer
    default = 0
  }
  column "last_commit_at" {
    null = true
    type = timestamptz
  }
  column "staged" {
    null    = false
    type    = integer
    default = 0
  }
  column "unstaged" {
    null    = false
    type    = integer
    default = 0
  }
  column "both" {
    null    = false
    type    = integer
    default = 0
  }
  column "untracked" {
    null    = false
    type    = integer
    default = 0
  }
  column "conflict" {
    null    = false
    type    = integer
    default = 0
  }
  column "adds" {
    null    = false
    type    = integer
    default = 0
  }
  column "dels" {
    null    = false
    type    = integer
    default = 0
  }
  column "touched_at" {
    null = true
    type = timestamptz
  }
  // The uncommitted files ([]status.FileStatus), replaced whole on every
  // status scan whose fingerprint changed.
  column "files" {
    null    = false
    type    = jsonb
    default = sql("'[]'::jsonb")
  }
  column "status_fingerprint" {
    null = true
    type = text
  }
  column "status_scanned_at" {
    null = true
    type = timestamptz
  }
  column "status_error" {
    null = true
    type = text
  }
  // Set when the worktree disappears from `git worktree list`; the row is
  // kept so sessions that recorded it can still render it as removed.
  column "removed_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo_id, column.path] }
  foreign_key "git_worktrees_repo_id_fkey" {
    columns     = [column.repo_id]
    ref_columns = [table.git_repos.column.id]
    on_delete   = CASCADE
  }
}

table "git_branches" {
  schema = schema.public

  column "repo_id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = text
  }
  column "head_sha" {
    null = false
    type = text
  }
  column "worktree_path" {
    null = true
    type = text
  }
  column "last_commit_at" {
    null = true
    type = timestamptz
  }

  primary_key { columns = [column.repo_id, column.name] }
  foreign_key "git_branches_repo_id_fkey" {
    columns     = [column.repo_id]
    ref_columns = [table.git_repos.column.id]
    on_delete   = CASCADE
  }
}

table "git_range_stats" {
  schema = schema.public

  column "repo_id" {
    null = false
    type = uuid
  }
  column "base_sha" {
    null = false
    type = text
  }
  column "head_sha" {
    null = false
    type = text
  }
  column "merge_base" {
    null = false
    type = text
  }
  column "ahead" {
    null = false
    type = integer
  }
  column "behind" {
    null = false
    type = integer
  }
  column "commits" {
    null = false
    type = integer
  }
  column "files" {
    null = false
    type = integer
  }
  column "adds" {
    null = false
    type = integer
  }
  column "dels" {
    null = false
    type = integer
  }
  // The files of merge_base..head_sha ([]git.CommitFile without repomap
  // enrichment); NULL until the range's file list is first opened.
  column "file_list" {
    null = true
    type = jsonb
  }
  column "computed_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }

  primary_key { columns = [column.repo_id, column.base_sha, column.head_sha] }
  foreign_key "git_range_stats_repo_id_fkey" {
    columns     = [column.repo_id]
    ref_columns = [table.git_repos.column.id]
    on_delete   = CASCADE
  }
}
