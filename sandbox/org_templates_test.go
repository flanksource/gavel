// Tests each organization's substrate template against an in-process Substrate: a copy of the base
// template at the organization's size, made on first use, followed when the base or the size changes.
package sandbox

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var baseSize = substrate.Size{CPUMillicores: 2000, MemoryBytes: 4 << 30}

// newOrgTemplates is an organization in gavel's database and org templates of size on an
// in-process Substrate.
func newOrgTemplates(t *testing.T, size substrate.Size) (*OrgTemplates, *substratetest.Server, uuid.UUID) {
	t.Helper()
	_, db, _, org := newDispatcher(t, &memoryExecutor{})
	templates, server := orgTemplatesOn(t, db, size)
	return templates, server, org
}

// orgTemplatesOn are org templates for db's organizations, copied from the base template
// gavel-runner, of baseSize, on an in-process Substrate.
func orgTemplatesOn(t *testing.T, db *gorm.DB, size substrate.Size) (*OrgTemplates, *substratetest.Server) {
	t.Helper()
	server := substratetest.Start(t)
	server.AddTemplate("ate-env", "gavel-runner", baseSize)
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return &OrgTemplates{DB: db, Client: client, Atespace: "ate-env", Base: "gavel-runner", Size: size}, server
}

func TestAnOrganizationGetsItsOwnTemplateAtTheBaseSize(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()

	got, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.Regexp(t, "^gavel-org-"+org.String()+"-[0-9a-f]{8}$", got.Name)
	assert.LessOrEqual(t, len(got.Name), 63, "a template's name is a DNS label")
	assert.Equal(t, baseSize, got.Size)
	created, ok := server.Template("ate-env", got.Name)
	require.True(t, ok, "the organization's template is made on first use")
	assert.Equal(t, int64(2000), created.GetResources().GetLimits().GetCpuMillicores())

	again, err := templates.Resolve(ctx, org)
	require.NoError(t, err, "the organization's template already being there is no error")
	assert.Equal(t, got, again)
}

func TestTheDefaultSizeFillsWhatItLeavesUnsetFromTheBase(t *testing.T) {
	templates, _, org := newOrgTemplates(t, substrate.Size{CPUMillicores: 1000})

	got, err := templates.Resolve(context.Background(), org)
	require.NoError(t, err)
	assert.Equal(t, substrate.Size{CPUMillicores: 1000, MemoryBytes: 4 << 30}, got.Size)
}

func TestAnOrganizationsOwnSizeWinsOverTheDefault(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{CPUMillicores: 1000, MemoryBytes: 2 << 30})
	ctx := context.Background()
	other := uuid.New()
	require.NoError(t, templates.DB.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, other, other.String()).Error)
	byDefault, err := templates.Resolve(ctx, other)
	require.NoError(t, err)

	require.NoError(t, SetOrgSize(ctx, templates.DB, org, &substrate.Size{CPUMillicores: 4000, MemoryBytes: 8 << 30}))
	got, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.Equal(t, substrate.Size{CPUMillicores: 4000, MemoryBytes: 8 << 30}, got.Size)
	created, ok := server.Template("ate-env", got.Name)
	require.True(t, ok)
	assert.Equal(t, int64(8<<30), created.GetResources().GetLimits().GetMemoryBytes())
	assert.Equal(t, substrate.Size{CPUMillicores: 1000, MemoryBytes: 2 << 30}, byDefault.Size, "other organizations keep the default")

	require.NoError(t, SetOrgSize(ctx, templates.DB, org, nil))
	reset, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.Equal(t, substrate.Size{CPUMillicores: 1000, MemoryBytes: 2 << 30}, reset.Size, "without its own size the organization has the default")
	assert.NotEqual(t, got.Name, reset.Name, "each size has its own template, since a template can't change")
}

func TestSetOrgSizeRefusesASizeWithoutCPUOrMemory(t *testing.T) {
	templates, _, org := newOrgTemplates(t, substrate.Size{})

	err := SetOrgSize(context.Background(), templates.DB, org, &substrate.Size{CPUMillicores: 1000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "an organization's sandbox size needs both CPU and memory")
}

func TestSetOrgSizeSaysWhichOrganizationIsMissing(t *testing.T) {
	templates, _, _ := newOrgTemplates(t, substrate.Size{})
	missing := uuid.New()

	err := SetOrgSize(context.Background(), templates.DB, missing, &baseSize)
	require.ErrorIs(t, err, ErrOrgNotFound)
	assert.Contains(t, err.Error(), missing.String())
}

func TestAnOrganizationsTemplateFollowsANewBaseTemplate(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	before, err := templates.Resolve(ctx, org)
	require.NoError(t, err)

	// An operator registers the runner's next image under the same name.
	require.NoError(t, templates.Client.DeleteTemplate(ctx, "gavel-runner"))
	server.AddTemplate("ate-env", "gavel-runner", baseSize)
	after, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.NotEqual(t, before.Name, after.Name, "a template copied from the old base runs the old image")
	_, ok := server.Template("ate-env", after.Name)
	assert.True(t, ok)
}

func TestResolveSaysWhichOrganizationAndBaseTemplateWhenTheBaseIsMissing(t *testing.T) {
	templates, _, org := newOrgTemplates(t, substrate.Size{})
	templates.Base = "no-such-template"

	_, err := templates.Resolve(context.Background(), org)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization "+org.String())
	assert.Contains(t, err.Error(), "actor template ate-env/no-such-template")
}
