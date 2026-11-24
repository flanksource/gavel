// Tests the platform API's sandbox size of an organization: platform admins read and set the
// CPU and memory its substrate environments get, or return it to serve's default.
package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/captain/pkg/substrate/substratetest"
	"github.com/flanksource/gavel/platform"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// platformSend sends body as JSON to url with method and decodes the answer into into.
func platformSend(t *testing.T, method, url string, body any, into any) int {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&payload).Encode(body))
	}
	req, err := http.NewRequest(method, url, &payload)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(into))
	return resp.StatusCode
}

// sizedPlatform is the platform API on db with org templates copied from gavel-runner, of 2 CPU
// and 4 GiB, whose default CPU is 1 core.
func sizedPlatform(t *testing.T) (*PlatformAPI, uuid.UUID) {
	t.Helper()
	db, _, org := openOrgDB(t)
	server := substratetest.Start(t)
	server.AddTemplate("ate-env", "gavel-runner", substrate.Size{CPUMillicores: 2000, MemoryBytes: 4 << 30})
	client, err := substrate.Dial(server.Options())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	templates := &sandbox.OrgTemplates{DB: db, Client: client, Atespace: "ate-env", Base: "gavel-runner", Size: substrate.Size{CPUMillicores: 1000}}
	return &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}, Sizes: templates}, org
}

func TestPlatformSetsAndResetsAnOrganizationsSandboxSize(t *testing.T) {
	p, org := sizedPlatform(t)
	url := servePlatform(t, p, org, true) + "/api/platform/orgs/" + org.String() + "/sandbox-size"
	defaults := &PlatformSize{CPUMillicores: 1000, MemoryBytes: 4 << 30}

	var got PlatformSandboxSize
	require.Equal(t, http.StatusOK, platformGet(t, url, &got))
	assert.Equal(t, PlatformSandboxSize{Default: defaults}, got, "an organization without a size of its own gets serve's default, filled from the base template")

	own := PlatformSize{CPUMillicores: 4000, MemoryBytes: 8 << 30}
	require.Equal(t, http.StatusOK, platformSend(t, http.MethodPut, url, own, &got))
	assert.Equal(t, PlatformSandboxSize{Size: &own, Default: defaults}, got)
	stored, err := sandbox.OrgSize(t.Context(), p.Store.DB, org)
	require.NoError(t, err)
	assert.Equal(t, &substrate.Size{CPUMillicores: 4000, MemoryBytes: 8 << 30}, stored)

	require.Equal(t, http.StatusOK, platformSend(t, http.MethodDelete, url, nil, &got))
	assert.Equal(t, PlatformSandboxSize{Default: defaults}, got)
}

func TestPlatformRefusesASandboxSizeItCantSet(t *testing.T) {
	p, org := sizedPlatform(t)
	base := servePlatform(t, p, org, true) + "/api/platform/orgs/"

	var failure map[string]string
	assert.Equal(t, http.StatusBadRequest, platformSend(t, http.MethodPut, base+org.String()+"/sandbox-size", PlatformSize{CPUMillicores: 4000}, &failure))
	assert.Contains(t, failure["error"], "an organization's sandbox size needs both CPU and memory")

	missing := uuid.New()
	assert.Equal(t, http.StatusNotFound, platformSend(t, http.MethodPut, base+missing.String()+"/sandbox-size", PlatformSize{CPUMillicores: 1000, MemoryBytes: 1 << 30}, &failure))
	assert.Contains(t, failure["error"], missing.String())
	assert.Equal(t, http.StatusNotFound, platformGet(t, base+missing.String()+"/sandbox-size", &failure))

	assert.Equal(t, http.StatusBadRequest, platformGet(t, base+"acme/sandbox-size", &failure))
	assert.Contains(t, failure["error"], `organization ID "acme"`)

	assert.Equal(t, http.StatusForbidden, platformSend(t, http.MethodPut, servePlatform(t, p, org, false)+"/api/platform/orgs/"+org.String()+"/sandbox-size",
		PlatformSize{CPUMillicores: 1000, MemoryBytes: 1 << 30}, &failure), "only platform admins size an organization")
}

func TestPlatformSandboxSizeWithoutSubstrate(t *testing.T) {
	db, _, org := openOrgDB(t)
	p := &PlatformAPI{Store: platform.Store{DB: db, Issuer: "https://clerk.example"}}

	var failure map[string]string
	assert.Equal(t, http.StatusNotFound, platformGet(t, servePlatform(t, p, org, true)+"/api/platform/orgs/"+org.String()+"/sandbox-size", &failure))
	assert.Contains(t, failure["error"], "--sandbox substrate")
}
