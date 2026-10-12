package merge

import (
	"fmt"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/verify"
)

// NewOptions resolves a merge from the workspace's configuration and the
// caller's flags. It is the single place the layering is assembled, so the CLI
// command and the dashboard's bulk action cannot resolve the prompt differently.
func NewOptions(dir string, flags bulk.MergeFlags) (Options, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return Options{}, fmt.Errorf("load .gavel.yaml for %s: %w", dir, err)
	}
	saved, err := loadSavedConfig()
	if err != nil {
		return Options{}, err
	}
	return Options{
		WorkDir:  dir,
		Into:     flags.Into,
		DryRun:   flags.DryRun,
		Base:     cfg.AI,
		Override: cfg.Todos.Merge,
		Request:  flags.Spec(),
		Saved:    saved,
		Runtime:  captaincli.AIRuntimeOptions{},
	}, nil
}
