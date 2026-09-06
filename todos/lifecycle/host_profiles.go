package lifecycle

import (
	"context"
	"errors"
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
	assembled, err := h.profileLayers(ctx, in)
	if err != nil {
		return runtimeprofiles.ResolveResult{}, err
	}
	resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{Layers: assembled.Layers, Saved: h.savedDefaults(), RequireModel: in.RequireModel})
	return runtimeprofiles.ResolveResult{Profile: assembled.Profile, Resolved: resolved}, runtimeConfigurationError(err)
}

func (h *Host) profileLayers(ctx context.Context, in LayerInput) (runtimeprofiles.LayerResult, error) {
	if err := api.ValidateSpecLayers(projectLayers(in)...); err != nil {
		return runtimeprofiles.LayerResult{}, &ConfigurationError{Err: err}
	}
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
		var selection *runtimeprofiles.SelectionError
		var ownedLayers *runtimeprofiles.OwnedLayersError
		if errors.As(err, &selection) && (selection.Origin != runtimeprofiles.SelectionRequested || errors.As(err, &ownedLayers) || errors.Is(err, runtimeprofiles.ErrCatalogUnavailable)) {
			err = &ConfigurationError{Err: err}
		}
		return runtimeprofiles.LayerResult{}, err
	}
	assembled.Layers = RestrictHostPermissions(assembled.Layers)
	if err := ValidateRequestPermissions(assembled.Layers); err != nil {
		return runtimeprofiles.LayerResult{}, err
	}
	return assembled, nil
}

func (h *Host) RuntimeCatalog(ctx context.Context) (*runtimeprofiles.Catalog, error) {
	if h.Catalog != nil {
		catalog, err := h.Catalog(ctx)
		if err != nil {
			return nil, &ConfigurationError{Err: err}
		}
		return catalog, nil
	}
	options := runtimeprofiles.DefaultCatalogOptions{Cwd: h.WorkDir, Config: h.Saved}
	if provider, ok := h.Provider.(interface{ Captain() *captaindb.DB }); ok {
		options.Read = func(context.Context) (*captaindb.DB, error) {
			if db := provider.Captain(); db != nil {
				return db, nil
			}
			return nil, fmt.Errorf("lifecycle runtime catalog: provider has no Captain database")
		}
	}
	catalog, err := runtimeprofiles.NewDefaultCatalog(ctx, options)
	if err != nil {
		return nil, &ConfigurationError{Err: err}
	}
	return catalog, nil
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
