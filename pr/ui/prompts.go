package ui

import (
	"net/http"
	"sync"

	"github.com/flanksource/clicky/prompt"
)

// promptManagerOnce installs the process-wide prompt manager the first time the
// dashboard mounts its routes. Installing it makes prompt.HasInteractiveSink()
// true, so commit checks (and any other clicky prompt) route their questions to
// the dashboard instead of failing in the non-TTY server context. A single
// in-memory manager is shared by every Server/MultiServer in the process.
var promptManagerOnce sync.Once

func ensurePromptManager() *prompt.Manager {
	promptManagerOnce.Do(func() {
		if prompt.GlobalManager() == nil {
			prompt.SetDefault(prompt.NewManager(prompt.NewMemory(prompt.MemoryConfig{})))
		}
	})
	return prompt.GlobalManager()
}

// registerPromptRoutes mounts the generic prompt API under /api/todos so the
// dashboard can list, stream, and answer pending prompts (commit questions, tool
// approvals, AI elicitation) scoped to a todo (owner) or session (label).
func registerPromptRoutes(mux *http.ServeMux) {
	ensurePromptManager().RegisterHandlers(mux, "/api/todos")
}
