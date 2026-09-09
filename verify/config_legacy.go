package verify

import "fmt"

// LegacyField describes a .gavel.yaml path that no longer exists.
type LegacyField struct {
	// ReplacedBy is the path that carries this setting now. Empty means the
	// field was removed outright.
	ReplacedBy string
	// Note explains a removal. It is only read when ReplacedBy is empty.
	Note string
}

// LegacyConfigFields maps every .gavel.yaml path retired by commit ef35046c
// ("Introduce operation-scoped AI specs and prompt resolution") to where it
// went. That commit replaced the scalar `*Prompt` and `*model` fields with
// PromptSpec objects and deleted the top-level `verify:` section. It was a
// declared BREAKING CHANGE with no migration shim, so a config written before
// it still parses — its values are simply dropped on the floor.
//
// Without this table the only signal is "unknown field commit.summaryPrompt",
// which is indistinguishable from a typo and says nothing about the setting
// still existing one rename away. Removals are listed too rather than omitted:
// "this is gone" is an answer, and leaving the key out would render the same
// bare warning.
//
// TestLegacyConfigFields_MatchesLiveSchema checks every ReplacedBy against the
// live config structs, and checks that no removed path has quietly come back,
// so the table cannot outlive the fields it describes.
var LegacyConfigFields = map[string]LegacyField{
	// Prompt overrides became operation-scoped PromptSpec objects.
	"commit.messagePrompt":      {ReplacedBy: "commit.message"},
	"commit.groupingPrompt":     {ReplacedBy: "commit.grouping"},
	"commit.summaryPrompt":      {ReplacedBy: "commit.summary"},
	"commit.prContentPrompt":    {ReplacedBy: "pr.content"},
	"todos.runPrompt":           {ReplacedBy: "todos.run"},
	"todos.planPrompt":          {ReplacedBy: "todos.plan"},
	"status.summaryPrompt":      {ReplacedBy: "status.summary"},
	"test.outlineSummaryPrompt": {ReplacedBy: "test.outlineSummary"},

	// Non-prompt renames from the same change.
	"commit.linkedDeps": {ReplacedBy: "commit.precommit"},
	"commit.model":      {ReplacedBy: "ai.model"},
	"commit.groupModel": {ReplacedBy: "commit.grouping.model"},

	// Dropped when commit-time compatibility analysis was removed.
	"commit.compatibility":              {Note: "compatibility analysis no longer runs"},
	"commit.compatibilityPrompt":        {Note: "compatibility analysis no longer runs"},
	"commit.functionalityRemovedPrompt": {Note: "compatibility analysis no longer runs"},

	// The whole section went away. Only the root key is ever reported, because
	// the field walker stops descending once a node is unknown.
	"verify": {Note: "use the top-level ai: spec and todos.verify"},

	// Retired when the ad-hoc run-mode/driver model gave way to the lifecycle.
	"todos.driver":    {ReplacedBy: "ai.model"},
	"todos.prompts":   {Note: "declare a lifecycle step under todos.<step>"},
	"todos.groupBy":   {Note: "grouping was removed; runs dispatch per todo"},
	"todos.approvals": {Note: "set permissions.mode: default on the step"},

	"checks.maxIterations": {ReplacedBy: "todos.run.workflow.verify.maxIterations"},
}

// legacyFieldHints renders the table into the path -> hint form the decoder
// takes. Built once: the map is immutable and every config layer load would
// otherwise rebuild it.
var legacyFieldHints = buildLegacyFieldHints()

func buildLegacyFieldHints() map[string]string {
	hints := make(map[string]string, len(LegacyConfigFields))
	for path, field := range LegacyConfigFields {
		hints[path] = field.hint()
	}
	return hints
}

func (f LegacyField) hint() string {
	if f.ReplacedBy != "" {
		return fmt.Sprintf("renamed to %s", f.ReplacedBy)
	}
	if f.Note != "" {
		return fmt.Sprintf("removed; %s", f.Note)
	}
	return "removed"
}
