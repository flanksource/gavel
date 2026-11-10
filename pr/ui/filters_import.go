// Moves the hidden projects and hidden GitHub orgs an earlier gavel kept in its config
// files into the local org's named filters, then drops them from the files.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/flanksource/gavel/filters"
	"github.com/flanksource/gavel/tenant"
)

// ImportFileFilters moves projects.json's "hidden": true entries into the local
// hidden-projects filter and ui.settings.json's "ignoredOrgs" into the local
// hidden-github-orgs filter, then rewrites each file without those fields. A file
// without them is left untouched, so running it again changes nothing.
func ImportFileFilters(ctx context.Context, store *filters.Store) error {
	if err := importHiddenProjects(ctx, store); err != nil {
		return fmt.Errorf("import hidden projects from %s: %w", projectsPath, err)
	}
	if err := importIgnoredOrgs(ctx, store); err != nil {
		return fmt.Errorf("import ignored orgs from %s: %w", settingsPath, err)
	}
	return nil
}

func importHiddenProjects(ctx context.Context, store *filters.Store) error {
	var entries []map[string]json.RawMessage
	found, err := readJSONFile(projectsPath, &entries)
	if !found || err != nil {
		return err
	}
	changed := false
	for _, entry := range entries {
		raw, ok := entry["hidden"]
		if !ok {
			continue
		}
		var hidden bool
		var project Project
		if err := errors.Join(json.Unmarshal(raw, &hidden), json.Unmarshal(entry["dir"], &project.Dir)); err != nil {
			return fmt.Errorf("project %s: %w", entry["name"], err)
		}
		if hidden {
			if _, err := store.Add(ctx, tenant.Local(), filters.HiddenProjects, project.ResolvedDir()); err != nil {
				return fmt.Errorf("project %s: %w", entry["name"], err)
			}
		}
		delete(entry, "hidden")
		changed = true
	}
	if !changed {
		return nil
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return writeProjectsFile(data)
}

func importIgnoredOrgs(ctx context.Context, store *filters.Store) error {
	var settings map[string]json.RawMessage
	found, err := readJSONFile(settingsPath, &settings)
	if !found || err != nil {
		return err
	}
	raw, ok := settings["ignoredOrgs"]
	if !ok {
		return nil
	}
	var orgs []string
	if err := json.Unmarshal(raw, &orgs); err != nil {
		return err
	}
	for _, org := range orgs {
		if _, err := store.Add(ctx, tenant.Local(), filters.HiddenGitHubOrgs, org); err != nil {
			return err
		}
	}
	delete(settings, "ignoredOrgs")
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0o644)
}

// readJSONFile decodes path into into, reporting false for a missing file.
func readJSONFile(path string, into any) (bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, into)
}
