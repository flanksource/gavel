package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api/registry"
	captaindb "github.com/flanksource/captain/pkg/database"
)

func sessionTerminalAgent(ctx context.Context, dir, sessionID string) (string, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", fmt.Errorf("sessionId is required to locate the terminal")
	}
	store, ok := todoCaptainSessionStore(ctx, dir)
	if !ok {
		return "", fmt.Errorf("Captain session store is unavailable for %s", sessionID)
	}
	resolved, err := resolveCaptainSession(ctx, store, sessionID)
	if err != nil {
		return "", err
	}
	if resolved.run != nil {
		runs, err := store.ListPromptRuns(ctx, captaindb.PromptRunFilter{SessionID: &resolved.run.ID})
		if err != nil {
			return "", err
		}
		if len(runs) > 0 {
			return terminalAgentFromProvider(runs[0].Runtime.Resolved.Provider)
		}
	}
	if resolved.transcript != nil {
		return terminalAgentFromProvider(resolved.transcript.AgentType)
	}
	return "", fmt.Errorf("session %s has no recorded terminal identity", sessionID)
}

func terminalAgentFromProvider(name string) (string, error) {
	provider, ok := registry.ProviderByName(name)
	if !ok || provider.AgentName == "" {
		return "", fmt.Errorf("session has unknown terminal provider %q", name)
	}
	return provider.AgentName, nil
}
