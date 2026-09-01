package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	todospec "github.com/flanksource/gavel/todos/spec"
	"github.com/flanksource/gavel/verify"
)

type todoRunContextResponse struct {
	// Modes is one row per (provider, mode) Captain can actually run, decorated
	// with that runtime's models, auth state and binary. Runtimes below is
	// Captain's own family catalog, which the display projection consumes.
	Modes           []todoRunModeOption  `json:"modes"`
	Runtimes        []api.RuntimeFamily  `json:"runtimes"`
	Models          []todoRunModelOption `json:"models"`
	Efforts         []string             `json:"efforts"`
	DefaultMode     string               `json:"defaultMode,omitempty"`
	DefaultProvider string               `json:"defaultProvider,omitempty"`
	// Tools is the catalog the run dialog's tool-permissions control renders, so
	// the per-tool Auto/Ask/Off choices map to real agent tools.
	Tools []todoRunToolOption `json:"tools"`
	// InputSchemas maps a run mode to the JSON Schema of its prompt inputs (the
	// test/lint options reflected from the engine structs, yaml-keyed like
	// fixture fences) so the dashboard can render the schema-driven run form
	// (clicky PromptDialog/JsonSchemaForm). Modes with no inputs are omitted.
	InputSchemas map[string]json.RawMessage `json:"inputSchemas,omitempty"`
	// PromptDefaults is the (mode, model) each named prompt actually resolves to,
	// keyed by prompt name.
	//
	// DefaultMode is Captain's account-wide default and knows nothing about a
	// prompt's frontmatter, so a dialog seeded from it sends that mode as if the
	// operator had chosen it — which outranks the frontmatter it was supposed to
	// defer to. `todos-triage.prompt` pins `model: claude` and declares a per-tool
	// policy only the Claude transports carry, so under a codex `ai.mode` the run
	// was rejected for a pairing nobody configured. These come from
	// todos/spec.Resolve, the same resolution the run itself performs.
	PromptDefaults map[string]todoRunPromptDefault `json:"promptDefaults,omitempty"`
}

// todoRunPromptDefault is one prompt's resolved runtime.
type todoRunPromptDefault struct {
	Mode  string `json:"mode,omitempty"`
	Model string `json:"model,omitempty"`
}

type todoRunModeOption struct {
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

type todoRunModelOption = captaincli.PromptModelCatalogEntry

var runCaptainWhoami = captaincli.RunWhoami

func (s *Server) handleTodoRunContext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// The prompt defaults are read from the workspace's .gavel.yaml layers, so the
	// dialog reflects the project the operator is looking at rather than the
	// process's own directory. An unknown dir resolves against the cwd, matching
	// how the CLI reads configuration.
	context, err := todoRunContext(strings.TrimSpace(r.URL.Query().Get("dir")))
	if err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("load run providers from Captain: %w", err))
		return
	}
	json.NewEncoder(w).Encode(context) //nolint:errcheck
}

func todoRunContext(workDir string) (todoRunContextResponse, error) {
	who, err := todoRunWhoami()
	if err != nil {
		return todoRunContextResponse{}, err
	}
	models := captaincli.PromptModelCatalog(who.Adapters)
	modes := make([]todoRunModeOption, 0, len(who.Adapters))
	for _, status := range who.Adapters {
		provider, ok := api.ProviderByName(strings.TrimSpace(status.Provider))
		mode := api.RuntimeMode(strings.TrimSpace(status.Mode))
		if !ok || !supportedTodoRunFamily(provider.AgentName) {
			continue
		}
		if _, err := drivers.Parse(string(mode)); err != nil {
			continue
		}
		modes = append(modes, todoRunModeOptionFor(provider, mode, status, who.ProviderDefaults[provider.Name].Model, models))
	}
	if len(modes) == 0 {
		return todoRunContextResponse{}, fmt.Errorf("captain returned no supported TODO run providers")
	}
	defaultMode := defaultTodoRunMode(who, modes)
	promptDefaults, err := todoRunPromptDefaults(workDir)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	return todoRunContextResponse{
		Modes:           modes,
		Runtimes:        who.Runtimes,
		Models:          models,
		Efforts:         todoRunEfforts(),
		DefaultMode:     defaultMode,
		DefaultProvider: who.DefaultProvider,
		Tools:           todoRunToolCatalog(),
		InputSchemas:    todoRunInputSchemas(),
		PromptDefaults:  promptDefaults,
	}, nil
}

// todoRunPromptDefaults resolves every catalogued prompt to the (mode, model)
// its run would use, so the dialog can seed itself from the prompt rather than
// from an account-wide default that overrides the prompt's own frontmatter.
//
// A prompt that fails to resolve is reported rather than skipped: it would fail
// the same way when run, and a silently absent default sends the operator back
// to the account default — the exact substitution this exists to prevent.
func todoRunPromptDefaults(workDir string) (map[string]todoRunPromptDefault, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return nil, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	catalog, err := todoprompt.NewCatalog(cfg.Todos)
	if err != nil {
		return nil, err
	}
	defaults := make(map[string]todoRunPromptDefault)
	for _, name := range catalog.Names() {
		resolved, err := todospec.Resolve(todospec.Input{WorkDir: workDir, Prompt: name})
		if err != nil {
			return nil, fmt.Errorf("resolve the %s prompt's runtime: %w", name, err)
		}
		mode, model, err := resolveTodoRunRuntime(
			resolved.Driver, string(resolved.Spec.Mode), resolved.Spec.Name)
		if err != nil {
			return nil, fmt.Errorf("resolve the %s prompt's runtime mode: %w", name, err)
		}
		defaults[name] = todoRunPromptDefault{Mode: mode, Model: model}
	}
	return defaults, nil
}

func todoRunModeOptionFor(
	provider *api.ModelProvider,
	runtimeMode api.RuntimeMode,
	status captaincli.AdapterStatus,
	captainDefaultModel string,
	catalog []captaincli.PromptModelCatalogEntry,
) todoRunModeOption {
	configured := status.Ready()
	mode := string(runtimeMode)
	models := todoRunModelsForMode(catalog, provider.Name, mode, configured)
	label := strings.ToUpper(provider.AgentName[:1]) + provider.AgentName[1:] + " " + mechanismLabel(mode)
	return todoRunModeOption{
		ID:            mode,
		Label:         label,
		Provider:      provider.Name,
		Agent:         provider.AgentName,
		DefaultModel:  defaultModelFromCatalog(models, captainDefaultModel),
		Driver:        mode,
		Mechanisms:    []todoRunMechanismItem{{Value: mode, Label: mechanismLabel(mode), Driver: mode}},
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

func supportedTodoRunFamily(family string) bool {
	return family == "claude" || family == "codex"
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

func todoRunModelsForMode(
	catalog []captaincli.PromptModelCatalogEntry,
	provider string,
	runtimeMode string,
	configured bool,
) []todoRunModelOption {
	out := make([]todoRunModelOption, 0)
	for _, model := range catalog {
		if model.Provider != provider || !slices.Contains(model.Modes, runtimeMode) {
			continue
		}
		model.Configured = configured
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

// resolveTodoRunRuntime returns the (mode, model) pair a run would execute on,
// falling back to the driver's own mechanism when the spec names none.
func resolveTodoRunRuntime(kind drivers.Kind, mode, model string) (string, string, error) {
	mode = strings.TrimSpace(mode)
	model = strings.TrimSpace(model)
	if mode == "" {
		mode = string(kind)
	}
	if model == "" {
		model = "claude"
	}
	resolved, err := registry.ResolveModel(api.Model{Name: model, Mode: registry.RuntimeMode(mode)})
	if err != nil {
		return "", "", err
	}
	return string(resolved.Mode), resolved.Name, nil
}

// providerKey names the family a spec's model belongs to. It reads the resolved
// descriptor when there is one and derives it from the name otherwise — the
// provider is a resolution result, never something the spec authored, so there
// is no field to read it from directly.
func providerKey(model api.Model) string {
	if model.Provider != nil {
		return model.Provider.Name
	}
	provider, err := api.ProviderFor(model.Name)
	if err != nil {
		return ""
	}
	return provider.Name
}

func mechanismLabel(mechanism string) string {
	switch mechanism {
	case "cmux":
		return "cmux (TUI)"
	case "agent":
		return "agent"
	case "cli":
		return "cli"
	case "api":
		return "API"
	default:
		return mechanism
	}
}
