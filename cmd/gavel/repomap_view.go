package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/repomap"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type RepomapViewOptions struct {
	Path string `json:"path" flag:"path" help:"Path to file or directory to view merged configuration for" default:"."`
}

func (opts RepomapViewOptions) GetName() string {
	return "view"
}

func (opts RepomapViewOptions) Help() api.Text {
	return clicky.Text(`View the merged ArchConf configuration for a given path.

This command shows the effective architecture configuration after merging
user-defined arch.yaml rules with embedded defaults. It displays which
configuration file was loaded and the complete merged configuration.

EXAMPLES:
  # View config for current directory
  gavel repomap view .

  # View config for specific file
  gavel repomap view ./main.go

  # View config for specific directory
  gavel repomap view ./cmd`)
}

type ArchConfView struct {
	ConfigSource string
	Config       *repomap.ArchConf
}

func (v ArchConfView) Pretty() api.Text {
	t := clicky.Text("")

	if v.ConfigSource != "" {
		t = t.Append("Configuration loaded from: ", "text-muted").Append(v.ConfigSource).NewLine().NewLine()
	} else {
		t = t.Append("No user configuration found, using embedded defaults only", "text-yellow-600").NewLine().NewLine()
	}

	t = t.Append("Merged Configuration:", "font-bold").NewLine().NewLine()
	t = t.Append(v.Config.Pretty())

	return t
}

func init() {
	clicky.AddCommand(repomapCmd, RepomapViewOptions{}, runRepomapView)
}

func runRepomapView(opts RepomapViewOptions) (any, error) {
	path := opts.Path
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", absPath)
	}

	userConf, err := repomap.GetConfForFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	configSource := findConfigSource(absPath)

	defaultConf, err := repomap.LoadDefaultArchConf()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded defaults: %w", err)
	}

	merged := defaultConf.Merge(userConf)

	return ArchConfView{
		ConfigSource: configSource,
		Config:       &merged,
	}, nil
}

func findConfigSource(path string) string {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}

	path, _ = filepath.Abs(path)

	for {
		archFile := filepath.Join(path, "arch.yaml")
		if stat, err := os.Stat(archFile); err == nil && !stat.IsDir() {
			return archFile
		}

		if repomap.IsGitRoot(path) {
			break
		}

		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}

	return ""
}
