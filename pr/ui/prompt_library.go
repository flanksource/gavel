package ui

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/prompts/library"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
)

const libraryPromptPrefix = "library."

type promptLibrary struct {
	Context context.Context
	Dir     string
	Sources []captaincli.PromptSourceInfo
	Prompts []captaincli.PromptSummary
}

func loadPromptLibrary(ctx context.Context, dir string) (promptLibrary, error) {
	ctx, err := library.Context(ctx, dir)
	if err != nil {
		return promptLibrary{}, err
	}
	sources, err := captaincli.DiscoverPromptSources(ctx, dir)
	if err != nil {
		return promptLibrary{}, err
	}
	prompts, err := captaincli.DiscoverPrompts(ctx, dir, captaincli.PromptListOptions{})
	return promptLibrary{Context: ctx, Dir: dir, Sources: sources, Prompts: prompts}, err
}

func libraryPromptName(summary captaincli.PromptSummary) string {
	return strings.TrimSuffix(filepath.Base(summary.RelPath), ".prompt")
}

func discoveredPromptEntries(ctx context.Context, scope, dir string) []promptCatalogEntry {
	lib, err := loadPromptLibrary(ctx, dir)
	if err != nil {
		return []promptCatalogEntry{{ID: libraryPromptPrefix + "discovery", Title: "Prompt library discovery", Owner: "captain", Source: "none", ParseError: err.Error(), Layers: []promptCatalogLayer{}}}
	}
	names := map[string]bool{}
	for _, summary := range lib.Prompts {
		names[libraryPromptName(summary)] = true
	}
	entries := make([]promptCatalogEntry, 0, len(names))
	for name := range names {
		entries = append(entries, lib.catalogEntry(name, scope))
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries
}

func (lib promptLibrary) catalogEntry(name, scope string) promptCatalogEntry {
	entry := promptCatalogEntry{ID: libraryPromptPrefix + name, Title: name, Owner: "captain", Source: "builtin"}
	detail, err := captaincli.InspectPrompt(lib.Context, lib.Dir, name)
	if err != nil {
		entry.ParseError = err.Error()
		return entry
	}
	entry.Title, entry.Description = detail.Name, detail.Description
	entry.Path, entry.Raw, entry.Version, entry.ParseError = detail.Path, detail.Content, promptSourceVersion(detail.Content), detail.ParseError
	if detail.SourceKind == "local" {
		entry.Source = "file"
	}
	entry.Layers = lib.layers(name, scope, detail.RelPath)
	for _, summary := range lib.Prompts {
		if libraryPromptName(summary) == name && summary.SourceKind == "embedded" {
			if summary.SourceID == "gavel" {
				entry.Owner = promptOwnerGavel
			}
			builtin, err := captaincli.InspectPrompt(lib.Context, lib.Dir, summary.ID)
			if err != nil {
				entry.ParseError = err.Error()
				return entry
			}
			entry.DefaultRaw = builtin.Content
			break
		}
	}
	entry.Variables = templateVariables(detail.Content)
	if detail.ParseError == "" {
		spec, body, _, err := promptregistry.ParsePromptSource(detail.Content)
		if err != nil {
			entry.ParseError = err.Error()
		} else {
			entry.Spec, entry.Body = spec, body
			entry.Effective = catalogRuntime(spec.Model, "prompt frontmatter")
		}
	}
	entry.Provenance = map[string]string{"body": detail.Source, "model": detail.Source}
	return entry
}

func (lib promptLibrary) layers(name, scope, rel string) []promptCatalogLayer {
	layers := make([]promptCatalogLayer, 0, len(lib.Sources))
	query := "scope=global"
	if scope != "global" {
		query = "project=" + url.QueryEscape(scope)
	}
	sources := append([]captaincli.PromptSourceInfo(nil), lib.Sources...)
	sort.SliceStable(sources, func(i, j int) bool { return sources[i].Writable && !sources[j].Writable })
	for i := len(sources) - 1; i >= 0; i-- {
		source := sources[i]
		layer := promptCatalogLayer{Origin: source.Label, Path: source.Label, Scope: query + "&source=" + url.QueryEscape(source.ID), Editable: source.Writable, Source: "none", FilePath: rel, SaveSources: []string{"file"}}
		found := false
		for _, summary := range lib.Prompts {
			if summary.SourceID == source.ID && libraryPromptName(summary) == name {
				layer.Source, layer.FilePath, layer.Path = "file", summary.RelPath, summary.Path
				found = true
			}
		}
		if source.Writable || found {
			layers = append(layers, layer)
		}
	}
	return layers
}

func (lib promptLibrary) detail(name, sourceID string) (captaincli.PromptDetail, bool, error) {
	for _, summary := range lib.Prompts {
		if libraryPromptName(summary) == name && summary.SourceID == sourceID {
			detail, err := captaincli.InspectPrompt(lib.Context, lib.Dir, summary.ID)
			return detail, true, err
		}
	}
	detail, err := captaincli.InspectPrompt(lib.Context, lib.Dir, name)
	return detail, false, err
}

func (lib promptLibrary) source(id string) (captaincli.PromptSourceInfo, error) {
	for _, source := range lib.Sources {
		if source.ID == id {
			return source, nil
		}
	}
	return captaincli.PromptSourceInfo{}, fmt.Errorf("unknown prompt source %q", id)
}
