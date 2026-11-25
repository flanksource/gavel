# One .gavel.yaml configuration layer per row: the tenant's own (project and path
# null), a project's root (path null), or a sub-folder of a project (path, slash
# separated, relative to the project). A row is either a file synced from disk
# (file is the .gavel.yaml or .gavel.<name>.yaml name, git_sha the repository HEAD
# it was read at, "none" outside git) or a layer authored in the database where no
# file exists (file and git_sha null). The settings service merges the layers of a
# target, nearest last, as captain merges spec layers.
table "gavel_settings" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "tenant_id" {
    null = false
    type = text
  }
  column "project_id" {
    null = true
    type = uuid
  }
  column "path" {
    null = true
    type = text
  }
  column "file" {
    null = true
    type = text
  }
  column "git_sha" {
    null = true
    type = text
  }
  column "payload" {
    null    = false
    type    = jsonb
    default = sql("'{}'::jsonb")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  foreign_key "gavel_settings_project_id_fkey" {
    columns     = [column.project_id]
    ref_columns = [table.projects.column.id]
    on_delete   = CASCADE
    on_update   = NO_ACTION
  }
  index "gavel_settings_layer_key" {
    unique = true
    on { column = column.tenant_id }
    on { expr = "COALESCE(project_id, '00000000-0000-0000-0000-000000000000'::uuid)" }
    on { expr = "COALESCE(path, '')" }
    on { expr = "COALESCE(file, '')" }
  }
  index "gavel_settings_project_idx" { columns = [column.project_id] }
  check "gavel_settings_path_in_project" { expr = "path IS NULL OR project_id IS NOT NULL" }
  check "gavel_settings_path_clean" { expr = "path IS NULL OR (path <> '' AND path !~ '(^/|^\\.\\.?(/|$)|/\\.\\.?(/|$)|/$)')" }
  check "gavel_settings_file_sha" { expr = "(file IS NULL) = (git_sha IS NULL)" }
}
