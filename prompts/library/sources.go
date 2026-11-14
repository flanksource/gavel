package library

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/internal/safefs"
)

func Context(ctx context.Context, directory string) (context.Context, error) {
	sources := captaincli.PromptSources{FS: []captaincli.PromptFS{{Label: Label, FS: FS}}}
	local := filepath.Join(directory, ".captain", "prompts")
	if info, err := safefs.Stat(directory, filepath.Join(".captain", "prompts")); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("prompt source %s must be a directory", local)
		}
		sources.Dirs = []string{local}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect prompt directory: %w", err)
	}
	return captaincli.ContextWithPromptSources(ctx, sources), nil
}
