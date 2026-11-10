package database

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentitySchemaBundle(t *testing.T) {
	identity := readEmbeddedSchemaFile(t, "schema/identity.hcl")
	for _, name := range []string{"users", "orgs", "session_owners", "api_token_owners", "runtime_profile_owners", "runtime_preset_owners"} {
		assert.Contains(t, identity, `table "`+name+`"`)
	}
	assert.Contains(t, identity, "users_clerk_identity_key")
	assert.Contains(t, identity, "orgs_clerk_identity_key")
	assert.Contains(t, identity, `column "slug"`)
	// Gavel installs no DDL on Captain's tables: ownership sidecars hold Captain
	// IDs without foreign keys, and every migration sweeps rows whose Captain
	// resource is gone.
	for _, file := range []string{"schema/identity.hcl", "schema/git_agent_ownership.hcl"} {
		assert.NotContains(t, readEmbeddedSchemaFile(t, file), "table.captain_", file)
	}
	for _, exclusion := range schemaExclusions {
		assert.NotContains(t, exclusion, "_captain_", "no SQL-installed constraint references Captain")
	}
	swept := map[string]string{}
	for _, sweep := range ownershipSweeps {
		swept[sweep.Sidecar] = sweep.Resource
	}
	assert.Equal(t, map[string]string{
		"session_owners": "captain_sessions", "api_token_owners": "captain_api_tokens", "git_agent_task_owners": "captain_git_agent_tasks",
		"runtime_profile_owners": "captain_runtime_profiles", "runtime_preset_owners": "captain_runtime_presets",
	}, swept)
	// Public favicon data has no authentication identity and remains shared.
	cache := readEmbeddedSchemaFile(t, "schema/github_cache.hcl")
	assert.Contains(t, cache, "http_cache_entries_user_id_fkey")
	assert.NotContains(t, cache, "_org_id_fkey", "org FKs on existing tables are installed by 151_tenancy_local_org.sql")
}

func TestTenancySchemaBundle(t *testing.T) {
	prepare := readEmbeddedSchemaFile(t, "schema/095_tenancy_prepare.sql")
	assert.Contains(t, prepare, "-- phase: pre")
	assert.Contains(t, prepare, "violations_file_path_fkey")

	localOrg := readEmbeddedSchemaFile(t, "schema/151_tenancy_local_org.sql")
	assert.Contains(t, localOrg, "-- phase: post")
	assert.Contains(t, localOrg, "'00000000-0000-0000-0000-000000000001', 'local', 'local', 'local', 'Local'")

	list := regexp.MustCompile(`(?s)FOREACH tbl IN ARRAY ARRAY\[(.*?)\]`).FindStringSubmatch(localOrg)
	require.Len(t, list, 2, "151 lists its tables in one FOREACH array")
	tables := regexp.MustCompile(`'(\w+)'`).FindAllStringSubmatch(list[1], -1)
	require.NotEmpty(t, tables)
	for _, table := range tables {
		// Atlas drops constraints it inspects but the HCL doesn't declare, so each
		// SQL-installed FK must be excluded or it churns on every migration.
		assert.Contains(t, schemaExclusions, table[1]+"."+table[1]+"_org_id_fkey")
	}

	tenancy := readEmbeddedSchemaFile(t, "schema/tenancy.hcl")
	for _, name := range []string{"org_memberships", "user_ssh_keys", "repos", "github_installations", "org_settings", "org_model_keys", "sandbox_runs"} {
		assert.Contains(t, tenancy, `table "`+name+`"`)
	}
}
