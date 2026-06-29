package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/drivers"
)

type todoRunContextResponse struct {
	Backends       []todoRunBackendOption `json:"backends"`
	Efforts        []string               `json:"efforts"`
	DefaultBackend string                 `json:"defaultBackend,omitempty"`
	// Tools is the catalog the run dialog's tool-permissions control renders, so
	// the per-tool Auto/Ask/Off choices map to real agent tools.
	Tools []todoRunToolOption `json:"tools"`
}

type todoRunToolOption struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Group       string `json:"group,omitempty"`
	DefaultMode string `json:"defaultMode,omitempty"`
}

// todoRunToolCatalog maps the default agent toolset (drivers.DefaultTools) onto
// the tool-preferences catalog, grouped for the picker. It is the single source
// of truth the dashboard mirrors via providers.ts's fallback.
func todoRunToolCatalog() []todoRunToolOption {
	group := map[string]string{
		"Read": "Files", "Edit": "Files", "Write": "Files",
		"Bash": "Shell",
		"Glob": "Search", "Grep": "Search",
	}
	tools := drivers.DefaultTools()
	out := make([]todoRunToolOption, 0, len(tools))
	for _, name := range tools {
		// Bash defaults to ask (brokered), matching the run's default posture where
		// command execution is surfaced for approval; the rest auto-run.
		mode := "enabled"
		if name == "Bash" {
			mode = "ask"
		}
		out = append(out, todoRunToolOption{Name: name, Label: name, Group: group[name], DefaultMode: mode})
	}
	return out
}

type todoRunBackendOption struct {
	ID            string                 `json:"id"`
	Label         string                 `json:"label"`
	Provider      string                 `json:"provider"`
	Agent         string                 `json:"agent"`
	DefaultModel  string                 `json:"defaultModel"`
	Driver        string                 `json:"driver"`
	Mechanisms    []todoRunMechanismItem `json:"mechanisms"`
	Models        []todoRunModelOption   `json:"models"`
	Configured    bool                   `json:"configured"`
	Type          string                 `json:"type,omitempty"`
	AuthMethod    string                 `json:"authMethod,omitempty"`
	AuthDetail    string                 `json:"authDetail,omitempty"`
	Binary        string                 `json:"binary,omitempty"`
	BinaryMissing string                 `json:"binaryMissing,omitempty"`
	ModelError    string                 `json:"modelError,omitempty"`
}

type todoRunMechanismItem struct {
	Value  string `json:"value"`
	Label  string `json:"label"`
	Driver string `json:"driver"`
}

type todoRunModelOption struct {
	ID         string `json:"id"`
	Provider   string `json:"provider"`
	Label      string `json:"label"`
	Reasoning  bool   `json:"reasoning"`
	Configured bool   `json:"configured"`
}

type runBackendSpec struct {
	Backend      captainai.Backend
	Label        string
	Provider     string
	Agent        string
	DefaultModel string
	Driver       drivers.Kind
}

func supportedTodoRunBackends() []runBackendSpec {
	return []runBackendSpec{
		{
			Backend:      captainai.BackendClaudeAgent,
			Label:        "Claude Agent",
			Provider:     "anthropic",
			Agent:        "claude",
			DefaultModel: "claude-agent-sonnet",
			Driver:       drivers.ClaudeHeadless,
		},
		{
			Backend:      captainai.BackendClaudeCLI,
			Label:        "Claude CLI",
			Provider:     "anthropic",
			Agent:        "claude",
			DefaultModel: "claude-agent-sonnet",
			Driver:       drivers.ClaudeHeadless,
		},
		{
			Backend:      captainai.BackendCodexCLI,
			Label:        "Codex CLI",
			Provider:     "openai",
			Agent:        "codex",
			DefaultModel: "gpt-5-codex",
			Driver:       drivers.CodexHeadless,
		},
	}
}

func (s *Server) handleTodoRunContext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode(todoRunContext()) //nolint:errcheck
}

func todoRunContext() todoRunContextResponse {
	specs := supportedTodoRunBackends()
	backends := make([]todoRunBackendOption, 0, len(specs))
	for _, spec := range specs {
		backends = append(backends, todoRunBackendOptionFor(spec))
	}
	return todoRunContextResponse{
		Backends:       backends,
		Efforts:        todoRunEfforts(),
		DefaultBackend: string(captainai.BackendClaudeAgent),
		Tools:          todoRunToolCatalog(),
	}
}

func todoRunBackendOptionFor(spec runBackendSpec) todoRunBackendOption {
	status := whoamiStatus(spec.Backend)
	configured := status.Ready()
	models := todoRunModelsForBackend(spec)
	return todoRunBackendOption{
		ID:            string(spec.Backend),
		Label:         spec.Label,
		Provider:      spec.Provider,
		Agent:         spec.Agent,
		DefaultModel:  defaultModelFromCatalog(models, spec.DefaultModel),
		Driver:        string(spec.Driver),
		Mechanisms:    []todoRunMechanismItem{{Value: spec.Driver.Mechanism(), Label: mechanismLabel(spec.Driver.Mechanism()), Driver: string(spec.Driver)}},
		Models:        models,
		Configured:    configured,
		Type:          status.Type,
		AuthMethod:    status.AuthMethod,
		AuthDetail:    status.AuthDetail,
		Binary:        status.Binary,
		BinaryMissing: status.BinaryMissing,
		ModelError:    status.ModelError,
	}
}

func whoamiStatus(backend captainai.Backend) captaincli.AdapterStatus {
	result, err := captaincli.RunWhoami(captaincli.WhoamiOptions{
		Backend: string(backend),
		Models:  true,
		Limit:   50,
	})
	if err != nil {
		return captaincli.AdapterStatus{Backend: string(backend), Type: backend.Kind(), ModelError: err.Error()}
	}
	who, ok := result.(captaincli.WhoamiResult)
	if !ok || len(who.Adapters) == 0 {
		return captaincli.AdapterStatus{Backend: string(backend), Type: backend.Kind(), ModelError: "captain whoami returned no adapter status"}
	}
	return who.Adapters[0]
}

func todoRunModelsForBackend(spec runBackendSpec) []todoRunModelOption {
	result, err := captaincli.RunAIModels(captaincli.AIModelsOptions{Backend: string(spec.Backend), Limit: 50})
	if err == nil {
		if rows, ok := result.(captaincli.AIModelsResult); ok && len(rows.Rows) > 0 {
			models := make([]todoRunModelOption, 0, len(rows.Rows))
			for _, row := range rows.Rows {
				models = append(models, todoRunModelOption{
					ID:         row.Model,
					Provider:   spec.Provider,
					Label:      modelLabel(row.Model),
					Reasoning:  true,
					Configured: true,
				})
			}
			return models
		}
	}
	return fallbackModelsForBackend(spec)
}

func fallbackModelsForBackend(spec runBackendSpec) []todoRunModelOption {
	models := []string{spec.DefaultModel}
	switch spec.Backend {
	case captainai.BackendClaudeAgent, captainai.BackendClaudeCLI:
		models = []string{"claude-agent-opus", "claude-agent-sonnet", "claude-agent-haiku"}
	case captainai.BackendCodexCLI:
		models = []string{"gpt-5-codex"}
	}
	out := make([]todoRunModelOption, 0, len(models))
	for _, id := range models {
		out = append(out, todoRunModelOption{
			ID:         id,
			Provider:   spec.Provider,
			Label:      modelLabel(id),
			Reasoning:  true,
			Configured: true,
		})
	}
	return out
}

func defaultModelFromCatalog(models []todoRunModelOption, fallback string) string {
	for _, model := range models {
		if model.ID == fallback {
			return fallback
		}
	}
	if len(models) > 0 {
		return models[0].ID
	}
	return fallback
}

func modelLabel(id string) string {
	label := strings.TrimPrefix(id, "claude-agent-")
	label = strings.TrimPrefix(label, "claude-code-")
	label = strings.TrimPrefix(label, "codex-")
	label = strings.ReplaceAll(label, "-", " ")
	if label == "" {
		return id
	}
	parts := strings.Fields(label)
	for i, part := range parts {
		switch strings.ToLower(part) {
		case "gpt":
			parts[i] = "GPT"
		case "api":
			parts[i] = "API"
		case "cli":
			parts[i] = "CLI"
		default:
			if len(part) > 0 {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}
	return strings.Join(parts, " ")
}

func todoRunEfforts() []string {
	efforts := api.AllEfforts()
	out := make([]string, 0, len(efforts))
	for _, effort := range efforts {
		out = append(out, string(effort))
	}
	return out
}

func validTodoRunEffort(effort string) bool {
	return api.Effort(effort).Valid() && effort != ""
}

func resolveTodoRunBackendModel(kind drivers.Kind, backend, model string) (string, string, error) {
	backend = strings.TrimSpace(backend)
	model = strings.TrimSpace(model)
	if backend == "" {
		backend = defaultBackendForDriver(kind)
	}
	if backend != "" {
		if err := validateBackendForDriver(kind, backend); err != nil {
			return "", "", err
		}
		if model == "" || model == kind.Agent() {
			model = defaultModelForBackend(backend)
		}
		if err := validateModelForBackend(backend, model); err != nil {
			return "", "", err
		}
	} else {
		if model == "" {
			model = kind.Agent()
		}
		if got, _ := claude.ResolveAgent(model); got != kind.Agent() {
			return "", "", fmt.Errorf("driver %s expects a %s model but %q resolves to %s", kind, kind.Agent(), model, got)
		}
	}
	return backend, model, nil
}

func defaultBackendForDriver(kind drivers.Kind) string {
	switch kind {
	case drivers.ClaudeHeadless:
		return string(captainai.BackendClaudeAgent)
	case drivers.CodexHeadless:
		return string(captainai.BackendCodexCLI)
	default:
		return ""
	}
}

func defaultModelForBackend(backend string) string {
	for _, spec := range supportedTodoRunBackends() {
		if string(spec.Backend) == backend {
			return spec.DefaultModel
		}
	}
	return ""
}

func validateBackendForDriver(kind drivers.Kind, backend string) error {
	b := captainai.Backend(backend)
	if !b.Valid() {
		return fmt.Errorf("invalid backend %q (valid: %s)", backend, captainai.BackendList())
	}
	for _, spec := range supportedTodoRunBackends() {
		if spec.Backend == b && spec.Driver == kind {
			return nil
		}
	}
	return fmt.Errorf("backend %q is not supported by driver %q", backend, kind)
}

func validateModelForBackend(backend, model string) error {
	switch captainai.Backend(backend) {
	case captainai.BackendClaudeAgent, captainai.BackendClaudeCLI:
		switch strings.ToLower(model) {
		case "claude", "sonnet", "opus", "haiku":
			return nil
		}
		if strings.HasPrefix(strings.ToLower(model), "claude-agent-") || strings.HasPrefix(strings.ToLower(model), "claude-code-") {
			return nil
		}
	case captainai.BackendCodexCLI:
		if model == "" || strings.EqualFold(model, "codex") {
			return nil
		}
		lower := strings.ToLower(model)
		if strings.HasPrefix(lower, "codex") || strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "o4") {
			return nil
		}
	}
	return fmt.Errorf("model %q is not valid for backend %q", model, backend)
}

func mechanismLabel(mechanism string) string {
	switch mechanism {
	case "cmux":
		return "cmux (TUI)"
	case "headless":
		return "headless"
	case "sdk":
		return "SDK"
	case "api":
		return "API"
	default:
		return mechanism
	}
}
