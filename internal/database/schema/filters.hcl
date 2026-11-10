# Named filters a user keeps, such as the projects and GitHub orgs they hide. Any
# list endpoint can reference a filter by name. Local mode has no user, so the
# local org's filters have a null owner; the unique key treats that null as one owner.
table "user_filters" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("gen_random_uuid()")
  }
  column "org_id" {
    null = false
    type = uuid
  }
  column "user_id" {
    null = true
    type = uuid
  }
  column "name" {
    null = false
    type = text
  }
  column "kind" {
    null = false
    type = text
  }
  column "values" {
    null    = false
    type    = sql("text[]")
    default = sql("'{}'::text[]")
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
  foreign_key "user_filters_org_id_fkey" {
    columns     = [column.org_id]
    ref_columns = [table.orgs.column.id]
    on_delete   = CASCADE
    on_update   = NO_ACTION
  }
  foreign_key "user_filters_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
    on_update   = NO_ACTION
  }
  index "user_filters_owner_name_key" {
    unique = true
    on { column = column.org_id }
    on { expr = "COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid)" }
    on { column = column.name }
  }
  index "user_filters_user_id_idx" { columns = [column.user_id] }
  check "user_filters_name_nonempty" { expr = "name <> '' AND name = btrim(name)" }
  check "user_filters_kind_check" {
    expr = "kind = ANY (ARRAY['workspace'::text, 'github-org'::text])"
  }
}
