// Tests each organization's substrate template against an in-process Substrate: a copy of the base
// template at the organization's size running serve's runner image, made on first use, followed when
// the base, the size or the runner image changes.
package sandbox

import (
	"context"
	"encoding/hex"
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
	return &OrgTemplates{DB: db, Client: client, Atespace: "ate-env", Base: "gavel-runner", Size: size, Gavel: "beta"}, server
}

func TestAnOrganizationGetsItsOwnTemplateAtTheBaseSize(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()

	got, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.Regexp(t, "^gavel-beta-"+hex.EncodeToString(org[:])+"-[0-9a-f]{8}$", got.Name)
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

func TestSweepDeletesTheTemplatesAnOrganizationNoLongerRunsFrom(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	old, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	require.NoError(t, SetOrgSize(ctx, templates.DB, org, &substrate.Size{CPUMillicores: 4000, MemoryBytes: 8 << 30}))
	current, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	server.AddTemplate("ate-env", "someone-elses", baseSize)
	server.AddTemplate("ate-env", "gavel-beta-not-an-organization", baseSize)

	swept, err := templates.Sweep(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{old.Name}, swept)
	_, ok := server.Template("ate-env", old.Name)
	assert.False(t, ok, "the template of the organization's old size is deleted")
	for _, kept := range []string{current.Name, "gavel-runner", "someone-elses", "gavel-beta-not-an-organization"} {
		_, ok := server.Template("ate-env", kept)
		assert.True(t, ok, "%s is kept: it is the organization's current template, or not an organization's", kept)
	}
}

func TestSweepKeepsATemplateAnEnvironmentStillRuns(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	old, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	options := server.Options()
	options.Template = old.Name
	running, err := substrate.Dial(options)
	require.NoError(t, err)
	defer running.Close()
	_, err = running.Create(ctx, substrate.Environment{ID: "gavel-run-1"})
	require.NoError(t, err)
	require.NoError(t, SetOrgSize(ctx, templates.DB, org, &substrate.Size{CPUMillicores: 4000, MemoryBytes: 8 << 30}))

	swept, err := templates.Sweep(ctx)
	require.NoError(t, err)
	assert.Empty(t, swept)
	_, ok := server.Template("ate-env", old.Name)
	assert.True(t, ok, "env can't delete an environment whose template is gone, so the template stays while it runs")
}

func TestSweepDeletesTheTemplatesOfOrganizationsGavelNoLongerHas(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	current, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	gone := uuid.New()
	removed := "gavel-beta-" + hex.EncodeToString(gone[:]) + "-0123abcd"
	server.AddTemplate("ate-env", removed, baseSize)

	swept, err := templates.Sweep(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{removed}, swept)
	_, ok := server.Template("ate-env", current.Name)
	assert.True(t, ok, "an organization that hasn't run in a while keeps its template: without a golden snapshot it costs nothing")
}

// Each gavel release ships its runner's digest, and its runs execute that runner: the organization's
// copy of the base template runs it, and the previous release's copy is swept.
func TestAnOrganizationsTemplateRunsServesRunnerImage(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	templates.Image = "ghcr.io/flanksource/gavel-runner@sha256:aaaa"

	previous, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	created, ok := server.Template("ate-env", previous.Name)
	require.True(t, ok)
	assert.Equal(t, "ghcr.io/flanksource/gavel-runner@sha256:aaaa", created.GetContainers()[0].GetImage())

	templates.Image = "ghcr.io/flanksource/gavel-runner@sha256:bbbb"
	got, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	assert.NotEqual(t, previous.Name, got.Name, "a new runner image is a new template: a template can't change")
	created, ok = server.Template("ate-env", got.Name)
	require.True(t, ok)
	assert.Equal(t, "ghcr.io/flanksource/gavel-runner@sha256:bbbb", created.GetContainers()[0].GetImage())

	swept, err := templates.Sweep(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{previous.Name}, swept)
}

// Without a runner image of its own, serve's runs execute the base template's.
func TestWithoutARunnerImageAnOrganizationsTemplateRunsTheBases(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})

	got, err := templates.Resolve(context.Background(), org)
	require.NoError(t, err)
	base, _ := server.Template("ate-env", "gavel-runner")
	created, ok := server.Template("ate-env", got.Name)
	require.True(t, ok)
	assert.Equal(t, base.GetContainers()[0].GetImage(), created.GetContainers()[0].GetImage())
}

// gavel-beta and gavel-prod share a substrate cluster: each sweeps only its own organizations'
// copies, so neither deletes the other's as belonging to an organization it doesn't have.
func TestSweepKeepsAnotherGavelsTemplates(t *testing.T) {
	templates, server, org := newOrgTemplates(t, substrate.Size{})
	ctx := context.Background()
	_, err := templates.Resolve(ctx, org)
	require.NoError(t, err)
	theirs := uuid.New()
	prods := "gavel-prod-" + hex.EncodeToString(theirs[:]) + "-0123abcd"
	server.AddTemplate("ate-env", prods, baseSize)

	swept, err := templates.Sweep(ctx)
	require.NoError(t, err)
	assert.Empty(t, swept)
	_, ok := server.Template("ate-env", prods)
	assert.True(t, ok, "gavel-prod's organization's template is gavel-prod's to sweep")
}

// A gavel's name is at most 15 characters, so its organizations' template names fit a DNS label.
func TestTheLongestGavelNameStillGivesADNSLabel(t *testing.T) {
	templates, _, org := newOrgTemplates(t, substrate.Size{})
	templates.Gavel = "gavel-staging01"

	got, err := templates.Resolve(context.Background(), org)
	require.NoError(t, err)
	assert.Len(t, got.Name, 63)
}
