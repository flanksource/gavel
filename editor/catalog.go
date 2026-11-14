package editor

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flanksource/captain/pkg/aichat"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
)

// runtimeCatalog serves captain's model and runtime catalogs (GET
// /api/chat/models and /api/chat/runtimes), which the new-session dialog's
// RuntimeBar chooses from. No chat turn is exposed.
func runtimeCatalog() http.Handler {
	service := aichat.NewService(aichat.ServiceOptions{Profile: aichat.RuntimeProfileProviderFunc(savedRuntimeProfile)})
	return service.Handler()
}

func savedRuntimeProfile(_ context.Context, options ...aichat.RuntimeProfileOption) (aichat.RuntimeProfile, error) {
	if selection := aichat.ApplyRuntimeProfileOptions(options...); len(selection.Presets) > 0 {
		return aichat.RuntimeProfile{}, aichat.RequestError(http.StatusBadRequest, "runtime presets are not available in gavel edit")
	}
	config, _, err := captainconfig.Load()
	if err != nil {
		return aichat.RuntimeProfile{}, fmt.Errorf("load Captain settings: %w", err)
	}
	if err := config.AI.Validate(); err != nil {
		return aichat.RuntimeProfile{}, fmt.Errorf("validate Captain settings: %w", err)
	}
	composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Saved: &config.AI})
	if err != nil {
		return aichat.RuntimeProfile{}, fmt.Errorf("compose the saved Captain runtime: %w", err)
	}
	return aichat.RuntimeProfile{Composed: composed, Saved: &config.AI}, nil
}
