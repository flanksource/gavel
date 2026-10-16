package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

type todoRunContextResponse struct {
	RuntimePresets  []runtimeprofiles.Preset      `json:"runtimePresets"`
	RuntimeProfiles []todoRunRuntimeProfileOption `json:"runtimeProfiles"`
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
	// Lifecycle is the project's step vocabulary: what the dialog's step picker
	// offers. The definition is data the browser has no copy of, so it is
	// served rather than hardcoded on the client.
	Lifecycle todoRunLifecycle `json:"lifecycle"`
	// InputSchemas maps a lifecycle step to the JSON Schema of its prompt inputs
	// (the test/lint options reflected from the engine structs, yaml-keyed like
	// fixture fences) so the dashboard can render the schema-driven run form
	// (clicky PromptDialog/JsonSchemaForm). Steps with no inputs are omitted.
	InputSchemas map[string]json.RawMessage `json:"inputSchemas,omitempty"`
	// PromptDefaults retains the shared lifecycle snapshot for display and preview.
	// It never becomes the browser's explicit request layer.
	PromptDefaults map[string]todoRunPromptDefault `json:"promptDefaults,omitempty"`
}

// todoRunLifecycle is the lifecycle as the run dialog sees it.
type todoRunLifecycle struct {
	Name  string              `json:"name"`
	Steps []todoRunStepOption `json:"steps"`
}

// todoRunStepOption is one step the dialog can name.
type todoRunStepOption struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	// Prompt is the step's prompt reference (`todos.run`, `file:...`).
	Prompt string `json:"prompt"`
	// ReadOnly marks a step whose class never edits or commits: a plan or triage
	// pass, or the verify step, which runs the definition of done rather than an
	// agent turn.
	ReadOnly bool `json:"readOnly,omitempty"`
	// Auxiliary marks a step the lifecycle never picks on its own.
	Auxiliary bool `json:"auxiliary,omitempty"`
}

// todoRunPromptDefault is one step's resolved runtime.
type todoRunPromptDefault struct {
	Mode           string                         `json:"mode,omitempty"`
	Model          string                         `json:"model,omitempty"`
	Presets        []string                       `json:"presets,omitempty"`
	RuntimeProfile string                         `json:"runtimeProfile,omitempty"`
	Spec           api.Spec                       `json:"spec"`
	Trace          []api.SpecLayer                `json:"trace"`
	Provenance     map[string]api.FieldProvenance `json:"provenance,omitempty"`
	Warnings       []string                       `json:"warnings,omitempty"`
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
	// The lifecycle and the step defaults are read from the workspace's
	// .gavel.yaml layers, so the dialog reflects the project the operator is
	// looking at rather than the process's own directory. An unknown dir resolves
	// against the cwd, matching how the CLI reads configuration.
	dir := strings.TrimSpace(r.URL.Query().Get("dir"))
	if s.hosted {
		context, err := s.hostedRunContext(r.Context(), dir)
		if err != nil {
			writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("load the run context of %s: %w", dir, err))
			return
		}
		json.NewEncoder(w).Encode(context) //nolint:errcheck
		return
	}
	context, err := todoRunContext(r.Context(), dir)
	if err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("load run providers from Captain: %w", err))
		return
	}
	json.NewEncoder(w).Encode(context) //nolint:errcheck
}

func todoRunContext(ctx context.Context, workDir string) (todoRunContextResponse, error) {
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
		if _, ok := registry.ParseRuntimeMode(string(mode)); !ok {
			continue
		}
		modes = append(modes, todoRunModeOptionFor(provider, mode, status, models))
	}
	if len(modes) == 0 {
		return todoRunContextResponse{}, fmt.Errorf("captain returned no supported TODO run providers")
	}
	provider, err := openTodoProvider(ctx, workDir)
	if err != nil {
		return todoRunContextResponse{}, fmt.Errorf("open runtime profile provider: %w", err)
	}
	host, err := lifecycle.NewHost(provider, workDir, lifecycle.HostDashboard)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	definition := host.Def.Definition()
	presets, err := todoRunPresetOptions(ctx, host)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	promptDefaults, err := todoRunPromptDefaults(ctx, host)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	inputSchemas, err := todoRunInputSchemas(definition.Steps)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	return todoRunContextResponse{
		RuntimePresets:  presets,
		RuntimeProfiles: []todoRunRuntimeProfileOption{},
		Modes:           modes,
		Runtimes:        who.Runtimes,
		Models:          models,
		Efforts:         todoRunEfforts(),
		DefaultMode:     promptDefaults["run"].Mode,
		DefaultProvider: providerKey(promptDefaults["run"].Spec.Model),
		Tools:           todoRunToolCatalog(),
		Lifecycle:       todoRunLifecycleFor(definition),
		InputSchemas:    inputSchemas,
		PromptDefaults:  promptDefaults,
	}, nil
}

// hostedRunContext is a hosted repository's run context: its own lifecycle, and for each step
// the runtime the repository's configuration resolves it to. Those are the only runtimes its
// runs use; serve's own agents and credentials aren't what a sandbox offers.
func (s *Server) hostedRunContext(ctx context.Context, dir string) (todoRunContextResponse, error) {
	provider, source, err := s.todoProviderContext(ctx, todoSource{Dir: dir})
	if err != nil {
		return todoRunContextResponse{}, err
	}
	host, err := lifecycle.NewRepositoryHost(provider, source.Dir, lifecycle.HostDashboard)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	definition := host.Def.Definition()
	promptDefaults, err := todoRunPromptDefaults(ctx, host)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	// A step's model may be an alias, such as `claude`, with no mode: the run resolves it to a
	// concrete model and mode, and the dialog shows what the run will use.
	for step, resolved := range promptDefaults {
		model, err := api.ResolveModel(resolved.Spec.Model)
		if err != nil {
			return todoRunContextResponse{}, fmt.Errorf("resolve the %s step's model %q: %w", step, resolved.Spec.Model.Name, err)
		}
		if model.Name != "" {
			resolved.Model, resolved.Mode = model.Name, string(model.Mode)
			promptDefaults[step] = resolved
		}
	}
	inputSchemas, err := todoRunInputSchemas(definition.Steps)
	if err != nil {
		return todoRunContextResponse{}, err
	}
	modes, models := stepDefaultRuntimes(promptDefaults)
	return todoRunContextResponse{
		RuntimePresets:  []runtimeprofiles.Preset{},
		RuntimeProfiles: []todoRunRuntimeProfileOption{},
		Modes:           modes,
		Runtimes:        runtimeFamiliesOf(modes),
		Models:          models,
		Efforts:         todoRunEfforts(),
		DefaultMode:     promptDefaults["run"].Mode,
		DefaultProvider: providerKey(promptDefaults["run"].Spec.Model),
		Tools:           todoRunToolCatalog(),
		Lifecycle:       todoRunLifecycleFor(definition),
		InputSchemas:    inputSchemas,
		PromptDefaults:  promptDefaults,
	}, nil
}

// stepDefaultRuntimes lists each (provider, mode) the steps default to, each offering only the
// models the steps default to in it.
func stepDefaultRuntimes(defaults map[string]todoRunPromptDefault) ([]todoRunModeOption, []todoRunModelOption) {
	steps := make([]string, 0, len(defaults))
	for step := range defaults {
		steps = append(steps, step)
	}
	slices.Sort(steps)
	modes, models := []todoRunModeOption{}, []todoRunModelOption{}
	for _, step := range steps {
		resolved := defaults[step]
		provider, ok := api.ProviderByName(providerKey(api.Model{Name: resolved.Model}))
		if !ok || resolved.Mode == "" || resolved.Model == "" {
			continue
		}
		model := todoRunModelOption{
			ID: resolved.Model, Provider: provider.Name, Label: modelLabel(resolved.Model), Configured: true,
			Modes: []string{resolved.Mode}, Runtime: api.Model{Name: resolved.Model, Mode: api.RuntimeMode(resolved.Mode)},
		}
		mode := slices.IndexFunc(modes, func(option todoRunModeOption) bool {
			return option.Provider == provider.Name && option.ID == resolved.Mode
		})
		if mode < 0 {
			modes = append(modes, todoRunModeOption{
				ID: resolved.Mode, Label: modeLabel(provider, resolved.Mode), Provider: provider.Name, Agent: provider.AgentName,
				DefaultModel: resolved.Model, Driver: resolved.Mode, Configured: true,
				Mechanisms: []todoRunMechanismItem{{Value: resolved.Mode, Label: mechanismLabel(resolved.Mode), Driver: resolved.Mode}},
			})
			mode = len(modes) - 1
		}
		if !slices.ContainsFunc(modes[mode].Models, func(option todoRunModelOption) bool { return option.ID == model.ID }) {
			modes[mode].Models = append(modes[mode].Models, model)
		}
		listed := slices.IndexFunc(models, func(option todoRunModelOption) bool {
			return option.Provider == model.Provider && option.ID == model.ID
		})
		switch {
		case listed < 0:
			models = append(models, model)
		case !slices.Contains(models[listed].Modes, resolved.Mode):
			models[listed].Modes = append(models[listed].Modes, resolved.Mode)
		}
	}
	return modes, models
}

// runtimeFamiliesOf is Captain's runtime catalog narrowed to the providers modes run on.
func runtimeFamiliesOf(modes []todoRunModeOption) []api.RuntimeFamily {
	families := []api.RuntimeFamily{}
	for _, family := range api.RuntimeCatalog() {
		if slices.ContainsFunc(modes, func(mode todoRunModeOption) bool { return mode.Provider == family.Provider }) {
			families = append(families, family)
		}
	}
	return families
}

// modeLabel names a runtime as the run dialog lists it: `Claude agent`.
func modeLabel(provider *api.ModelProvider, mode string) string {
	return strings.ToUpper(provider.AgentName[:1]) + provider.AgentName[1:] + " " + mechanismLabel(mode)
}

// todoRunLifecycleFor projects the definition onto the dialog's step picker.
func todoRunLifecycleFor(definition lifecycle.Lifecycle) todoRunLifecycle {
	steps := make([]todoRunStepOption, 0, len(definition.Steps))
	for _, step := range definition.Steps {
		steps = append(steps, todoRunStepOption{
			Name:      step.Name,
			Label:     stepLabel(step.Name),
			Prompt:    step.Prompt,
			ReadOnly:  lifecycle.Class(step) != types.ModeRun,
			Auxiliary: step.Auxiliary,
		})
	}
	return todoRunLifecycle{Name: definition.Name, Steps: steps}
}

// stepLabel is a step name as a picker shows it: `shape-it` → `Shape it`.
func stepLabel(name string) string {
	words := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(name))
	if len(words) == 0 {
		return name
	}
	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]
	return strings.Join(words, " ")
}

// todoRunPromptDefaults resolves every lifecycle step to the (mode, model) its
// run would use, so the dialog can seed itself from the step rather than from
// an account-wide default that overrides the prompt's own frontmatter.
//
// A step that fails to resolve is reported rather than skipped: it would fail
// the same way when run, and a silently absent default sends the operator back
// to the account default — the exact substitution this exists to prevent.
func todoRunPromptDefaults(ctx context.Context, host *lifecycle.Host) (map[string]todoRunPromptDefault, error) {
	steps := host.Def.Definition().Steps
	defaults := make(map[string]todoRunPromptDefault, len(steps))
	for _, step := range steps {
		spec, err := host.StepDefaults(ctx, step)
		if err != nil {
			return nil, fmt.Errorf("resolve the %s step's runtime: %w", step.Name, err)
		}
		defaults[step.Name] = todoRunPromptDefault{
			Mode: string(spec.Spec.Mode), Model: spec.Spec.Name, Presets: spec.Presets, RuntimeProfile: spec.RuntimeProfile,
			Spec: spec.Spec, Trace: spec.Trace, Provenance: spec.Provenance, Warnings: spec.Warnings,
		}
	}
	return defaults, nil
}

func todoRunModeOptionFor(
	provider *api.ModelProvider,
	runtimeMode api.RuntimeMode,
	status captaincli.AdapterStatus,
	catalog []captaincli.PromptModelCatalogEntry,
) todoRunModeOption {
	configured := status.Ready()
	mode := string(runtimeMode)
	models := todoRunModelsForMode(catalog, provider.Name, mode, configured)
	return todoRunModeOption{
		ID:            mode,
		Label:         modeLabel(provider, mode),
		Provider:      provider.Name,
		Agent:         provider.AgentName,
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
