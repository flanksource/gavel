package ui

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/captain/pkg/aichat"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaindb "github.com/flanksource/captain/pkg/database"
	clickyaichat "github.com/flanksource/clicky/aichat"
	"github.com/flanksource/commons-db/shell"
	"github.com/spf13/cobra"
)

const gavelChatSystemPrompt = "You are Gavel's development workflow assistant. Use the available TODO tools to inspect tracked work. " +
	"Prefer tool results over guesses, and ask before making changes."

func newGavelChatServer(root *cobra.Command, cwd string, db *captaindb.DB) (*aichat.Service, error) {
	threads, err := aichat.NewDatabaseThreadStore(db)
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat thread store: %w", err)
	}
	authority, err := aichat.NewDatabaseExecutionAuthority(db)
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat execution authority: %w", err)
	}
	provider, err := clickyaichat.NewCobraToolProvider(clickyaichat.CobraToolProviderOptions{
		Root: root,
		Strategies: []api.PermissionStrategy{
			api.HTTPVerbStrategy{}, api.MCPHintStrategy{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat tools: %w", err)
	}
	return aichat.NewService(aichat.ServiceOptions{
		Profile:        gavelChatRuntimeProfile(cwd),
		ToolStrategies: provider.Strategies(),
		ToolPolicy:     provider.ToolPolicy(),
		Tools:          provider,
		Threads:        aichat.FixedThreadStore(threads),
		Authority:      authority,
	}), nil
}

func gavelChatRuntimeProfile(cwd string) aichat.RuntimeProfileProvider {
	return aichat.RuntimeProfileProviderFunc(func(_ context.Context, options ...aichat.RuntimeProfileOption) (aichat.RuntimeProfile, error) {
		selection := aichat.ApplyRuntimeProfileOptions(options...)
		if len(selection.Presets) > 0 {
			return aichat.RuntimeProfile{}, aichat.RequestError(http.StatusBadRequest,
				fmt.Sprintf("runtime presets %q are not available in Gavel chat", strings.Join(selection.Presets, ",")))
		}
		config, _, err := captainconfig.Load()
		if err != nil {
			return aichat.RuntimeProfile{}, fmt.Errorf("load Captain chat settings: %w", err)
		}
		if err := config.AI.Validate(); err != nil {
			return aichat.RuntimeProfile{}, fmt.Errorf("validate Captain chat settings: %w", err)
		}
		composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{
			Layers: []api.SpecLayer{{
				Name: "gavel", Scope: api.SpecLayerGlobal,
				Spec: api.Spec{Setup: &shell.Setup{Cwd: cwd}},
			}},
			Saved: &config.AI,
		})
		if err != nil {
			return aichat.RuntimeProfile{}, fmt.Errorf("compose Gavel chat runtime profile: %w", err)
		}
		return aichat.RuntimeProfile{System: gavelChatSystemPrompt, Composed: composed, Saved: &config.AI}, nil
	})
}
