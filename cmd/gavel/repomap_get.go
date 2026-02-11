package main

import (
	"fmt"
	"os"

	"github.com/flanksource/gavel/repomap"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type RepomapGetOptions struct {
	Path string `json:"path" flag:"path" help:"Path to file or directory to get FileMap for" stdin:"true"`
}

func (opts RepomapGetOptions) GetName() string {
	return "get"
}

func (opts RepomapGetOptions) Help() api.Text {
	return clicky.Text(`Get the FileMap for a specific file path.

This command returns the file mapping including detected language, scopes,
and technologies based on the merged configuration rules.

EXAMPLES:
  # Get FileMap for a specific file
  gavel repomap get ./main.go

  # Get FileMap for current file
  gavel repomap get .

  # Get FileMap for a test file
  gavel repomap get ./cmd/repomap_test.go`)
}

func init() {
	clicky.AddCommand(repomapCmd, RepomapGetOptions{}, runRepomapGet)
}

func runRepomapGet(opts RepomapGetOptions) (any, error) {
	if opts.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", opts.Path)
	}

	fileMap, err := repomap.GetFileMap(opts.Path, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get file map: %w", err)
	}

	return fileMap, nil
}
