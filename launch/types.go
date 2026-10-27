package launch

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/resultstore"
)

type Request struct {
	Workspace     *shell.Setup      `json:"workspace,omitempty"`
	PreparationID string            `json:"preparationId,omitempty"`
	Config        Configuration     `json:"config"`
	Labels        map[string]string `json:"labels,omitempty"`
}

type Configuration struct {
	Test map[string]any `json:"test,omitempty"`
	Lint *LintConfig    `json:"lint,omitempty"`
}

type LintConfig struct {
	Files   []string       `json:"files,omitempty"`
	Changed bool           `json:"changed,omitempty"`
	Since   string         `json:"since,omitempty"`
	Linters []LinterConfig `json:"linters,omitempty"`
}

type LinterConfig struct {
	ID      string         `json:"id"`
	Linter  string         `json:"linter"`
	Options map[string]any `json:"options,omitempty"`
}

type PromptConfig struct {
	Prompt        string         `json:"prompt"`
	PromptVersion string         `json:"promptVersion"`
	Variables     map[string]any `json:"variables,omitempty"`
	Presets       []string       `json:"presets,omitempty"`
	Spec          *api.Spec      `json:"spec,omitempty"`
	Concurrency   int            `json:"concurrency,omitempty"`
}

type SchemaProvider interface {
	Schema(string) (map[string]any, error)
}

type Publish func(resultstore.Run) error

type Execution interface {
	Run(context.Context, Publish) error
}

type Executor interface {
	Prepare(context.Context, PreparedWorkspace, Configuration) (Execution, error)
}
