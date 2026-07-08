package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

type todoRunContextResponse struct {
	Backends       []todoRunBackendOption `json:"backends"`
	Efforts        []string               `json:"efforts"`
	DefaultBackend string                 `json:"defaultBackend,omitempty"`
	// Tools is the catalog the run dialog's tool-permissions control renders, so
	// the per-tool Auto/Ask/Off choices map to real agent tools.
	Tools []todoRunToolOption `json:"tools"`
	// InputSchemas maps a run mode to the JSON Schema of its prompt inputs (the
	// test/lint options reflected from the engine structs, yaml-keyed like
	// fixture fences) so the dashboard can render the schema-driven run form
	// (clicky PromptDialog/JsonSchemaForm). Modes with no inputs are omitted.
	InputSchemas map[string]json.RawMessage `json:"inputSchemas,omitempty"`
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

var runCaptainWhoami = captaincli.RunWhoami

func supportedTodoRunBackends() []runBackendSpec {
	return []runBackendSpec{
		{
			Backend:      captainai.BackendClaudeCmux,
			Label:        "Claude cmux",
			Provider:     "anthropic",
			Agent:        "claude",
			DefaultModel: "claude-agent-sonnet",
			Driver:       drivers.ClaudeCmux,
		},
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
			Backend:      captainai.BackendCodexCmux,
			Label:        "Codex cmux",
			Provider:     "openai",
			Agent:        "codex",
			DefaultModel: "gpt-5.5",
			Driver:       drivers.CodexCmux,
		},
		{
			Backend:      captainai.BackendCodexAgent,
			Label:        "Codex Agent",
			Provider:     "openai",
			Agent:        "codex",
			DefaultModel: "gpt-5.5",
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
		DefaultBackend: string(captainai.BackendClaudeCmux),
		Tools:          todoRunToolCatalog(),
		InputSchemas:   todoRunInputSchemas(),
	}
}

// todoRunInputSchemas collects each mode's prompt input schema; a mode with no
// inputs (plan, verify) is simply absent.
func todoRunInputSchemas() map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	for _, mode := range []types.RunMode{types.ModeRun, types.ModePlan} {
		raw, err := todoprompt.InputSchema(mode)
		if err != nil {
			logger.Warnf("todo run context: input schema for %s: %v", mode, err)
			continue
		}
		if raw != nil {
			out[string(mode)] = json.RawMessage(raw)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func todoRunBackendOptionFor(spec runBackendSpec) todoRunBackendOption {
	status := whoamiStatus(spec.Backend)
	configured := status.Ready()
	models := todoRunModelsFromWhoami(spec, status)
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
	result, err := runCaptainWhoami(captaincli.WhoamiOptions{
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

func todoRunModelsFromWhoami(spec runBackendSpec, status captaincli.AdapterStatus) []todoRunModelOption {
	candidates := make([]todoRunModelOption, 0, len(status.Models)+len(status.ModelDetails))
	if len(status.ModelDetails) > 0 {
		for _, model := range status.ModelDetails {
			candidates = append(candidates, todoRunModelOption{
				ID:         model.ID,
				Provider:   spec.Provider,
				Label:      firstNonEmpty(model.Name, modelLabel(model.ID)),
				Reasoning:  true,
				Configured: true,
			})
		}
	} else {
		for _, id := range status.Models {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			candidates = append(candidates, todoRunModelOption{
				ID:         id,
				Provider:   spec.Provider,
				Label:      modelLabel(id),
				Reasoning:  true,
				Configured: true,
			})
		}
	}

	out := make([]todoRunModelOption, 0, len(candidates))
	seen := map[string]bool{}
	for _, model := range candidates {
		originalID := model.ID
		model.ID = normalizeTodoRunModelForBackend(string(spec.Backend), model.ID)
		if model.ID == "" || seen[model.ID] {
			continue
		}
		seen[model.ID] = true
		model.Label = firstNonEmpty(model.Label, modelLabel(originalID))
		out = append(out, model)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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
	label = strings.TrimPrefix(label, "claude-")
	label = strings.TrimPrefix(label, "codex-")
	if formatted := claudeTierModelLabel(label); formatted != "" {
		return formatted
	}
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

func claudeTierModelLabel(id string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(id)), "-")
	if len(parts) == 0 {
		return ""
	}
	tier := parts[0]
	switch tier {
	case "fable", "opus", "sonnet", "haiku":
	default:
		return ""
	}
	label := titleWord(tier)
	version := []string{}
	i := 1
	for i < len(parts) && isModelVersionPart(parts[i]) {
		version = append(version, parts[i])
		i++
	}
	if len(version) > 0 {
		label += " " + strings.Join(version, ".")
	}
	if i < len(parts) {
		rest := make([]string, 0, len(parts)-i)
		for _, part := range parts[i:] {
			rest = append(rest, titleWord(part))
		}
		label += " " + strings.Join(rest, " ")
	}
	return label
}

func isModelVersionPart(part string) bool {
	if part == "" {
		return false
	}
	for _, r := range part {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func titleWord(part string) string {
	if part == "" {
		return part
	}
	return strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
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
		model = normalizeTodoRunModelForBackend(backend, model)
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

func normalizeTodoRunModelForBackend(backend, model string) string {
	return captainai.NormalizeModelForBackend(captainai.Backend(backend), model)
}

func defaultBackendForDriver(kind drivers.Kind) string {
	switch kind {
	case drivers.ClaudeCmux:
		return string(captainai.BackendClaudeCmux)
	case drivers.ClaudeHeadless:
		return string(captainai.BackendClaudeAgent)
	case drivers.CodexCmux:
		return string(captainai.BackendCodexCmux)
	case drivers.CodexHeadless:
		return string(captainai.BackendCodexAgent)
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
	case captainai.BackendClaudeAgent, captainai.BackendClaudeCLI, captainai.BackendClaudeCmux:
		lower := strings.ToLower(strings.TrimSpace(model))
		if lower == "claude" {
			return nil
		}
		for _, prefix := range []string{"fable", "opus", "sonnet", "haiku", "claude-", "claude-agent-", "claude-code-"} {
			if lower == strings.TrimSuffix(prefix, "-") || strings.HasPrefix(lower, prefix) {
				return nil
			}
		}
	case captainai.BackendCodexAgent, captainai.BackendCodexCmux:
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
