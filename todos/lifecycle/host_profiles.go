package lifecycle

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/verify"
)

type ProfileSelection struct {
	Requested string
	Pinned    string
	Default   string
}

func (h *Host) resolveProfileLayers(ctx context.Context, in LayerInput) (runtimeprofiles.ResolveResult, error) {
	options := runtimeprofiles.ResolveOptions{
		RequestedProfile: in.RuntimeProfile.Requested,
		PinnedProfile:    in.RuntimeProfile.Pinned,
		DefaultProfile:   in.RuntimeProfile.Default,
	}
	for _, layer := range Layers(in) {
		switch layer.Scope {
		case api.SpecLayerGlobal, api.SpecLayerContext:
			options.BaseLayers = append(options.BaseLayers, layer)
		default:
			options.SurfaceLayers = append(options.SurfaceLayers, layer)
		}
	}
	assembled, err := runtimeprofiles.NewResolver(h.RuntimeCatalog).Layers(ctx, options)
	if err != nil {
		return runtimeprofiles.ResolveResult{}, err
	}
	layers := RestrictHostPermissions(assembled.Layers)
	if err := ValidateRequestPermissions(layers); err != nil {
		return runtimeprofiles.ResolveResult{}, err
	}
	resolved, err := api.ResolveSpecLayers(layers...)
	return runtimeprofiles.ResolveResult{Profile: assembled.Profile, Resolved: resolved}, err
}

func (h *Host) RuntimeCatalog(ctx context.Context) (*runtimeprofiles.Catalog, error) {
	if h.Catalog != nil {
		return h.Catalog(ctx)
	}
	options := runtimeprofiles.DefaultCatalogOptions{Cwd: h.WorkDir}
	if provider, ok := h.Provider.(interface{ Captain() *captaindb.DB }); ok {
		options.Read = func(context.Context) (*captaindb.DB, error) {
			if db := provider.Captain(); db != nil {
				return db, nil
			}
			return nil, fmt.Errorf("lifecycle runtime catalog: provider has no Captain database")
		}
	}
	return runtimeprofiles.NewDefaultCatalog(ctx, options)
}

func (h *Host) profileSelection(step string, requested, pinned string) ProfileSelection {
	profile := strings.TrimSpace(stepPromptSpec(h.Config.Todos, step).RuntimeProfile)
	if profile == "" {
		profile = h.Config.Todos.RuntimeProfile
	}
	return ProfileSelection{Requested: requested, Pinned: pinned, Default: profile}
}

func stepPromptSpec(cfg verify.TodosConfig, step string) verify.PromptSpec {
	switch step {
	case "run":
		return cfg.Run
	case "plan":
		return cfg.Plan
	case "triage":
		return cfg.Triage
	case StepVerify:
		return cfg.Verify
	default:
		return cfg.Steps[step]
	}
}
