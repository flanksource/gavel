// Serves the sandbox size of an organization to platform admins: the CPU and memory its substrate
// environments get, which they set for it or return to serve's default.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
)

// PlatformSize is an amount of CPU and memory.
type PlatformSize struct {
	CPUMillicores int64 `json:"cpuMillicores"`
	MemoryBytes   int64 `json:"memoryBytes"`
}

// PlatformSandboxSize is an organization's sandbox size: Size is its own, nil without one, and
// Default what it gets without one. DefaultError says why Default is missing, when it is.
type PlatformSandboxSize struct {
	Size         *PlatformSize `json:"size"`
	Default      *PlatformSize `json:"default,omitempty"`
	DefaultError string        `json:"defaultError,omitempty"`
}

func (p *PlatformAPI) sandboxSize(w http.ResponseWriter, r *http.Request) {
	org, ok := p.sizedOrg(w, r)
	if !ok {
		return
	}
	p.writeSandboxSize(w, r.Context(), org)
}

func (p *PlatformAPI) setSandboxSize(w http.ResponseWriter, r *http.Request) {
	org, ok := p.sizedOrg(w, r)
	if !ok {
		return
	}
	var size PlatformSize
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&size); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("the sandbox size must be JSON like {\"cpuMillicores\": 2000, \"memoryBytes\": 4294967296}: %w", err))
		return
	}
	if !p.storeSandboxSize(w, r.Context(), org, &substrate.Size{CPUMillicores: size.CPUMillicores, MemoryBytes: size.MemoryBytes}) {
		return
	}
	p.writeSandboxSize(w, r.Context(), org)
}

func (p *PlatformAPI) resetSandboxSize(w http.ResponseWriter, r *http.Request) {
	org, ok := p.sizedOrg(w, r)
	if !ok {
		return
	}
	if !p.storeSandboxSize(w, r.Context(), org, nil) {
		return
	}
	p.writeSandboxSize(w, r.Context(), org)
}

// sizedOrg is the organization the request names, answering why when serve has no sandbox sizes
// or the ID isn't one.
func (p *PlatformAPI) sizedOrg(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	if p.Sizes == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("an organization's sandbox size is the size of its substrate environments: serve runs without --sandbox substrate"))
		return uuid.Nil, false
	}
	org, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("organization ID %q is not a UUID", r.PathValue("id")))
		return uuid.Nil, false
	}
	return org, true
}

// storeSandboxSize sets org's size, nil for the default, answering why when it can't.
func (p *PlatformAPI) storeSandboxSize(w http.ResponseWriter, ctx context.Context, org uuid.UUID, size *substrate.Size) bool {
	err := sandbox.SetOrgSize(ctx, p.Store.DB, org, size)
	switch {
	case errors.Is(err, sandbox.ErrInvalidOrgSize):
		writeJSONError(w, http.StatusBadRequest, err)
	case errors.Is(err, sandbox.ErrOrgNotFound):
		writeJSONError(w, http.StatusNotFound, err)
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
	default:
		return true
	}
	return false
}

// writeSandboxSize answers with org's size and the default, which substrate's base template fills.
func (p *PlatformAPI) writeSandboxSize(w http.ResponseWriter, ctx context.Context, org uuid.UUID) {
	size, err := sandbox.OrgSize(ctx, p.Store.DB, org)
	if errors.Is(err, sandbox.ErrOrgNotFound) {
		writeJSONError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	var result PlatformSandboxSize
	if size != nil {
		result.Size = &PlatformSize{CPUMillicores: size.CPUMillicores, MemoryBytes: size.MemoryBytes}
	}
	ctx, cancel := context.WithTimeout(ctx, platformSourceTimeout)
	defer cancel()
	if defaults, err := p.Sizes.DefaultSize(ctx); err != nil {
		result.DefaultError = err.Error()
	} else {
		result.Default = &PlatformSize{CPUMillicores: defaults.CPUMillicores, MemoryBytes: defaults.MemoryBytes}
	}
	writeJSON(w, result)
}
