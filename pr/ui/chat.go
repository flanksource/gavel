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
	"github.com/flanksource/gavel/ai/chatprompt"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

func newGavelChatServer(root *cobra.Command, cwd string, db *captaindb.DB) (*aichat.Service, error) {
	threads, err := aichat.NewDatabaseThreadStore(db)
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat thread store: %w", err)
	}
	authority, err := aichat.NewDatabaseExecutionAuthority(db)
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat execution authority: %w", err)
	}
	provider, err := newGavelChatTools(root)
	if err != nil {
		return nil, err
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

// newGavelChatTools derives the assistant's tools from the entity command tree:
// every generated operation is a tool, and a read (GET) runs without asking.
func newGavelChatTools(root *cobra.Command) (*clickyaichat.CobraToolProvider, error) {
	provider, err := clickyaichat.NewCobraToolProvider(clickyaichat.CobraToolProviderOptions{
		Root: root,
		Strategies: []api.PermissionStrategy{
			api.HTTPVerbStrategy{}, api.MCPHintStrategy{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create Gavel chat tools: %w", err)
	}
	return provider, nil
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
		cfg, err := verify.LoadGavelConfig(cwd)
		if err != nil {
			return aichat.RuntimeProfile{}, err
		}
		layers, err := cfg.Chat.Assistant.Layers(verify.PromptResolveOptions{Base: cfg.AI, DefaultPrompt: chatprompt.Default, Dir: cwd, Name: "chat.assistant"})
		if err != nil {
			return aichat.RuntimeProfile{}, err
		}
		layers = append(layers, api.SpecLayer{
			Name: "gavel", Scope: api.SpecLayerGlobal,
			Spec: api.Spec{Setup: &shell.Setup{Cwd: cwd}},
		})
		composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{
			Layers: layers,
			Saved:  &config.AI,
		})
		if err != nil {
			return aichat.RuntimeProfile{}, fmt.Errorf("compose Gavel chat runtime profile: %w", err)
		}
		return aichat.RuntimeProfile{Composed: composed, Saved: &config.AI}, nil
	})
}
