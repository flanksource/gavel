# The project catalog when gavel runs with PostgreSQL: the GitHub repos each
# workspace directory backs, or a web project with neither. Names are the public key (comment scopes are
# gavel://project/<name>), unique per tenant. repos holds only registered repos;
# repos inferred from the checkout's remotes are derived on read.
table "projects" {
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
  column "name" {
    null = false
    type = text
  }
  column "kind" {
    null    = false
    type    = text
    default = "dir"
  }
  column "dir" {
    null = true
    type = text
  }
  column "repos" {
    null    = false
    type    = jsonb
    default = sql("'[]'::jsonb")
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.id] }
  index "projects_tenant_name_key" {
    unique  = true
    columns = [column.tenant_id, column.name]
  }
  check "projects_name_nonempty" { expr = "btrim(name) <> ''" }
  check "projects_kind" { expr = "kind = ANY (ARRAY['dir'::text, 'web'::text])" }
  check "projects_kind_dir" { expr = "(kind = 'web') = (dir IS NULL)" }
  check "projects_dir_nonempty" { expr = "dir IS NULL OR btrim(dir) <> ''" }
  check "projects_repos_array" { expr = "jsonb_typeof(repos) = 'array'" }
}

# One row per tenant whose projects.json was imported into projects, so the
# file is read once: deleting every project afterwards does not re-import it.
table "project_imports" {
  schema = schema.public
  column "tenant_id" {
    null = false
    type = text
  }
  column "source" {
    null = false
    type = text
  }
  column "imported_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key { columns = [column.tenant_id] }
}
