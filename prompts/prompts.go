// Package prompts is the registry of Gavel's overridable AI prompt templates.
//
// Each entry pairs a stable ID with the embedded default .prompt (dotprompt)
// source and the dotted .gavel.yaml path of its typed override field. Owning
// packages keep the //go:embed of their default and expose it through a
// Prompts() function returning these descriptors; the settings UI composes them
// to render one editor per prompt (showing the default and whether it is
// overridden). Resolution of an override stays at each call site — the typed
// verify.PromptOverride.Resolve runs against the embedded default — so this
// package is a leaf with no dependencies and never participates in rendering.
package prompts

// Stable prompt IDs. They are the JSON-Schema x-prompt-id that links an override
// field's schema node to its registry descriptor, so they MUST match the
// x-prompt-id stamped by the config schema generator.
const (
	Verify              = "verify"
	CommitMessage       = "commit.message"
	CommitFuncRemoved   = "commit.functionalityRemoved"
	CommitCompatibility = "commit.compatibility"
	CommitSummary       = "commit.summary"
	PRContent           = "pr.content"
	TodosRun            = "todos.run"
)

// Prompt describes one overridable AI prompt template for the settings UI.
type Prompt struct {
	// ID is stable and matches the schema node's x-prompt-id.
	ID string `json:"id"`
	// Title is a short human label (e.g. "Commit message").
	Title string `json:"title"`
	// Description explains what the prompt drives and when it runs.
	Description string `json:"description"`
	// ConfigPath is the dotted .gavel.yaml location of the typed override field,
	// e.g. "verify.promptTemplate".
	ConfigPath string `json:"configPath"`
	// Default is the embedded .prompt source used when the override is unset; the
	// UI shows it as the built-in default.
	Default string `json:"default"`
}
