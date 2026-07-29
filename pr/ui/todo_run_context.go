package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
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
// the tool-preferences catalog, grouped for the picker.
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
	ID                string   `json:"id"`
	Provider          string   `json:"provider"`
	Label             string   `json:"label"`
	CapabilitiesKnown bool     `json:"capabilitiesKnown,omitempty"`
	Reasoning         bool     `json:"reasoning"`
	Temperature       *bool    `json:"temperature,omitempty"`
	SupportedEfforts  []string `json:"supportedEfforts,omitempty"`
	DefaultEffort     string   `json:"defaultEffort,omitempty"`
	Configured        bool     `json:"configured"`
}

type runBackendSpec struct {
	Backend  captainai.Backend
	Label    string
	Provider string
	Agent    string
	Driver   drivers.Kind
}

var runCaptainWhoami = captaincli.RunWhoami

func supportedTodoRunBackends() []runBackendSpec {
	return []runBackendSpec{
		{
			Backend:  captainai.BackendClaudeCmux,
			Label:    "Claude cmux",
			Provider: "anthropic",
			Agent:    "claude",
			Driver:   drivers.Cmux,
		},
		{
			Backend:  captainai.BackendClaudeAgent,
			Label:    "Claude Agent",
			Provider: "anthropic",
			Agent:    "claude",
			Driver:   drivers.Cli,
		},
		{
			Backend:  captainai.BackendClaudeCLI,
			Label:    "Claude CLI",
			Provider: "anthropic",
			Agent:    "claude",
			Driver:   drivers.Cli,
		},
		{
			Backend:  captainai.BackendCodexCmux,
			Label:    "Codex cmux",
			Provider: "openai",
			Agent:    "codex",
			Driver:   drivers.Cmux,
		},
		{
			Backend:  captainai.BackendCodexAgent,
			Label:    "Codex Agent",
			Provider: "openai",
			Agent:    "codex",
			Driver:   drivers.Cli,
		},
	}
}

func (s *Server) handleTodoRunContext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	context, err := todoRunContext()
	if err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("load run providers from Captain: %w", err))
		return
	}
	json.NewEncoder(w).Encode(context) //nolint:errcheck
}

func todoRunContext() (todoRunContextResponse, error) {
	specs := supportedTodoRunBackends()
	who, err := todoRunWhoami()
	if err != nil {
		return todoRunContextResponse{}, err
	}
	specsByBackend := make(map[captainai.Backend]runBackendSpec, len(specs))
	for _, spec := range specs {
		specsByBackend[spec.Backend] = spec
	}
	backends := make([]todoRunBackendOption, 0, len(who.Adapters))
	for _, status := range who.Adapters {
		backend := captainai.Backend(strings.TrimSpace(status.Backend))
		spec, ok := specsByBackend[backend]
		if !ok {
			continue
		}
		backends = append(backends, todoRunBackendOptionFor(spec, status, who.ProviderDefaults[spec.Provider].Model))
	}
	if len(backends) == 0 {
		return todoRunContextResponse{}, fmt.Errorf("captain returned no supported TODO run providers")
	}
	defaultBackend := ""
	if defaults, ok := who.ProviderDefaults[who.DefaultProvider]; ok {
		for _, backend := range backends {
			if backend.ID == defaults.Agent {
				defaultBackend = backend.ID
				break
			}
		}
	}
	return todoRunContextResponse{
		Backends:       backends,
		Efforts:        todoRunEfforts(),
		DefaultBackend: defaultBackend,
		Tools:          todoRunToolCatalog(),
		InputSchemas:   todoRunInputSchemas(),
	}, nil
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

func todoRunBackendOptionFor(spec runBackendSpec, status captaincli.AdapterStatus, captainDefaultModel string) todoRunBackendOption {
	configured := status.Ready()
	models := todoRunModelsFromWhoami(spec, status)
	mode := runtimeModeForTodoBackend(spec.Backend, spec.Driver)
	return todoRunBackendOption{
		ID:            string(spec.Backend),
		Label:         spec.Label,
		Provider:      spec.Provider,
		Agent:         spec.Agent,
		DefaultModel:  defaultModelFromCatalog(models, captainDefaultModel),
		Driver:        string(spec.Driver),
		Mechanisms:    []todoRunMechanismItem{{Value: mode, Label: mechanismLabel(mode), Driver: string(spec.Driver)}},
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

func todoRunWhoami() (captaincli.WhoamiResult, error) {
	result, err := runCaptainWhoami(captaincli.WhoamiOptions{
		Models: true,
	})
	if err != nil {
		return captaincli.WhoamiResult{}, err
	}
	who, ok := result.(captaincli.WhoamiResult)
	if !ok {
		return captaincli.WhoamiResult{}, fmt.Errorf("captain whoami returned %T, want WhoamiResult", result)
	}
	return who, nil
}

func todoRunModelsFromWhoami(spec runBackendSpec, status captaincli.AdapterStatus) []todoRunModelOption {
	candidates := make([]todoRunModelOption, 0, len(status.Models)+len(status.ModelDetails))
	if len(status.ModelDetails) > 0 {
		for _, model := range status.ModelDetails {
			reasoning := true
			var temperature *bool
			if model.CapabilitiesKnown {
				reasoning = model.Reasoning
				value := model.Temperature
				temperature = &value
			}
			efforts := make([]string, 0, len(model.SupportedEfforts))
			for _, effort := range model.SupportedEfforts {
				efforts = append(efforts, string(effort))
			}
			candidates = append(candidates, todoRunModelOption{
				ID:                model.ID,
				Provider:          spec.Provider,
				Label:             firstNonEmpty(model.Name, modelLabel(model.ID)),
				CapabilitiesKnown: model.CapabilitiesKnown,
				Reasoning:         reasoning,
				Temperature:       temperature,
				SupportedEfforts:  efforts,
				DefaultEffort:     string(model.DefaultEffort),
				Configured:        true,
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
		model.ID = strings.TrimSpace(model.ID)
		if model.ID == "" || seen[model.ID] {
			continue
		}
		seen[model.ID] = true
		model.Label = firstNonEmpty(model.Label, modelLabel(model.ID))
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
	return ""
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
	// The coding agent is derived from the model (empty → claude); the driver is
	// mechanism-only. With no explicit backend, pick the (agent, mechanism) backend.
	agent, _ := claude.ResolveAgent(model)
	if backend == "" {
		backend = defaultBackendForDriver(kind, agent)
	}
	if backend != "" {
		if err := validateBackendForDriver(kind, backend); err != nil {
			return "", "", err
		}
		if model == "" {
			provider, _, ok := registry.ProviderFor(captainai.Backend(backend))
			if !ok {
				return "", "", fmt.Errorf("captain has no provider for backend %q", backend)
			}
			model = provider.AgentName
		}
		model = normalizeTodoRunModelForBackend(backend, model)
		if err := validateModelForBackend(backend, model); err != nil {
			return "", "", err
		}
	} else if model == "" {
		model = agent
	}
	return backend, model, nil
}

func normalizeTodoRunModelForBackend(backend, model string) string {
	return captainai.NormalizeModelForBackend(captainai.Backend(backend), model)
}

// defaultBackendForDriver is the default captain backend for a (mechanism, agent)
// pair. The (agent, mechanism) → backend mapping is captain's Provider.BackendFor,
// reached here through drivers.BackendFor; the driver's cli mechanism resolves to
// the agent-SDK backend, matching the catalog's "Agent wins" default.
func defaultBackendForDriver(kind drivers.Kind, agent string) string {
	backend, err := drivers.BackendFor(agent, kind)
	if err != nil {
		return ""
	}
	return string(backend)
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

// validateModelForBackend checks the model belongs to the backend's provider
// family. The family classification is captain's — a model's provider (its
// registry descriptor) and the backend's provider must be the same — rather than
// a local strings.HasPrefix(lower, "gpt-") heuristic that duplicated it.
func validateModelForBackend(backend, model string) error {
	b := captainai.Backend(backend)
	want, _, ok := registry.ProviderFor(b)
	if !ok {
		return fmt.Errorf("invalid backend %q (valid: %s)", backend, captainai.BackendList())
	}
	// A bare agent sentinel ("claude"/"codex") is valid for its own family.
	if model == "" || strings.EqualFold(strings.TrimSpace(model), want.AgentName) {
		return nil
	}
	got, _, _, claimed := registry.ProviderForToken(normalizeTodoRunModelForBackend(backend, model))
	if claimed && got == want {
		return nil
	}
	return fmt.Errorf("model %q is not valid for backend %q", model, backend)
}

func runtimeModeForTodoBackend(backend captainai.Backend, driver drivers.Kind) string {
	if mode := backend.Mode(); mode != "" {
		return string(mode)
	}
	if mechanism := driver.Mechanism(); mechanism != "" {
		return mechanism
	}
	return string(backend)
}

func mechanismLabel(mechanism string) string {
	switch mechanism {
	case "cmux":
		return "cmux (TUI)"
	case "agent":
		return "agent"
	case "cli":
		return "cli"
	case "sdk":
		return "SDK"
	case "api":
		return "API"
	default:
		return mechanism
	}
}
