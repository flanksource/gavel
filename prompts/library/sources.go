package library

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	captaincli "github.com/flanksource/captain/pkg/cli"
)

func Context(ctx context.Context, directory string) (context.Context, error) {
	sources := captaincli.PromptSources{FS: []captaincli.PromptFS{{Label: Label, FS: FS}}}
	local := filepath.Join(directory, ".captain", "prompts")
	if _, err := os.Stat(local); err == nil {
		sources.Dirs = []string{local}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect prompt directory: %w", err)
	}
	return captaincli.ContextWithPromptSources(ctx, sources), nil
}
