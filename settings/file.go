package settings

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/verify"
)

// File resolves configuration from disk only: every layer is a .gavel.yaml or
// snippet in the target's chain.
type File struct{}

var _ Service = File{}

func (File) Resolve(_ context.Context, target Target) (Resolved, error) {
	positions, layers, err := diskChain(target)
	if err != nil {
		return Resolved{}, err
	}
	return compose(positions, layers)
}

func (File) Save(_ context.Context, target Target, ref Ref, edit func(*verify.GavelConfig) error) error {
	if ref.File == "" {
		return fmt.Errorf("file-only settings have no database layer: save %s %q to a .gavel.yaml or .gavel.<name>.yaml", ref.Kind, ref.Path)
	}
	positions, _, err := diskChain(target)
	if err != nil {
		return err
	}
	position, err := positionFor(positions, ref)
	if err != nil {
		return err
	}
	return saveFile(position, ref.File, edit)
}

// saveFile applies edit to the config file name at position, creating it when absent.
func saveFile(position Position, name string, edit func(*verify.GavelConfig) error) error {
	if !verify.IsGavelConfigFile(name) {
		return fmt.Errorf("%q is not a %s or .gavel.<name>.yaml file", name, verify.GavelConfigFileName)
	}
	if !filepath.IsAbs(position.Dir) {
		return fmt.Errorf("%s %q of a web project has no directory to hold %s: save its database layer instead", position.Kind, position.Path, name)
	}
	path := filepath.Join(position.Dir, name)
	cfg, err := verify.LoadSingleGavelConfig(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := edit(&cfg); err != nil {
		return err
	}
	if err := verify.SaveGavelConfig(path, cfg); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}
