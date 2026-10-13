# Internal UUIDs remain independent of Clerk's external IDs. The issuer is part
# of each identity so development/production Clerk instances never collide.
table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "issuer" {
    null = false
    type = text
  }
  column "clerk_user_id" {
    null = false
    type = text
  }
  column "name" {
    null = false
    type = text
    default = ""
  }
  column "email" {
    null = false
    type = text
    default = ""
  }
  column "email_verified" {
    null = false
    type = boolean
    default = false
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
  index "users_clerk_identity_key" {
    unique = true
    columns = [column.issuer, column.clerk_user_id]
  }
  check "users_identity_nonempty" { expr = "issuer <> '' AND clerk_user_id <> ''" }
}

table "orgs" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "issuer" {
    null = false
    type = text
  }
  column "clerk_org_id" {
    null = false
    type = text
  }
  # Slug is metadata, not the tenant's stable identity. OAuth may omit it.
  column "slug" {
    null = true
    type = text
  }
  column "name" {
    null = false
    type = text
    default = ""
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
  index "orgs_clerk_identity_key" {
    unique = true
    columns = [column.issuer, column.clerk_org_id]
  }
  index "orgs_issuer_slug_key" {
    unique = true
    columns = [column.issuer, column.slug]
    where = "slug IS NOT NULL"
  }
  check "orgs_identity_nonempty" { expr = "issuer <> '' AND clerk_org_id <> ''" }
  check "orgs_slug_nonempty" { expr = "slug IS NULL OR (slug <> '' AND slug = btrim(slug))" }
}

# Captain owns its schema. These Gavel-owned sidecars attach ownership without
# modifying upstream tables or risking removal by Captain's migrator. Child
# turns, runs, plans, calls, messages, events and artifacts inherit via session.
# They hold Captain IDs without foreign keys (Gavel installs no DDL on Captain's
# tables); every migration deletes rows whose Captain resource is gone.
table "session_owners" {
  schema = schema.public
  column "session_id" {
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
  primary_key { columns = [column.session_id] }
  foreign_key "session_owners_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "session_owners_user_id_idx" { columns = [column.user_id] }
  index "session_owners_org_id_idx" { columns = [column.org_id] }
}

table "api_token_owners" {
  schema = schema.public
  column "token_id" {
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
  primary_key { columns = [column.token_id] }
  foreign_key "api_token_owners_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "api_token_owners_user_id_idx" { columns = [column.user_id] }
  index "api_token_owners_org_id_idx" { columns = [column.org_id] }
}

table "runtime_profile_owners" {
  schema = schema.public
  column "profile_id" {
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
  primary_key { columns = [column.profile_id] }
  foreign_key "runtime_profile_owners_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "runtime_profile_owners_user_id_idx" { columns = [column.user_id] }
  index "runtime_profile_owners_org_id_idx" { columns = [column.org_id] }
}

table "runtime_preset_owners" {
  schema = schema.public
  column "preset_id" {
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
  primary_key { columns = [column.preset_id] }
  foreign_key "runtime_preset_owners_user_id_fkey" {
    columns = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete = SET_NULL
    on_update = NO_ACTION
  }
  index "runtime_preset_owners_user_id_idx" { columns = [column.user_id] }
  index "runtime_preset_owners_org_id_idx" { columns = [column.org_id] }
}
