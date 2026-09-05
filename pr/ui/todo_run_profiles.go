package ui

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/todos/lifecycle"
)

type todoRunRuntimeProfileOption struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Model       string                     `json:"model,omitempty"`
	Presets     []string                   `json:"presets"`
	Source      runtimeprofiles.SourceInfo `json:"source"`
}

type todoRunProfilePreview struct {
	Profile runtimeprofiles.Profile  `json:"profile"`
	Presets []runtimeprofiles.Preset `json:"presets"`
}

func todoRunProfilePreviewFor(profile *runtimeprofiles.Resolution) *todoRunProfilePreview {
	if profile == nil {
		return nil
	}
	return &todoRunProfilePreview{Profile: profile.Profile, Presets: profile.Presets}
}

func todoRunProfileOptions(ctx context.Context, host *lifecycle.Host) ([]todoRunRuntimeProfileOption, error) {
	catalog, err := host.RuntimeCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("load runtime profiles: %w", err)
	}
	host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) { return catalog, nil }
	profiles, err := catalog.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtime profiles: %w", err)
	}
	options := make([]todoRunRuntimeProfileOption, 0, len(profiles))
	for _, profile := range profiles {
		options = append(options, todoRunRuntimeProfileOption{
			ID: profile.ID, Name: profile.Name, Description: profile.Description,
			Model: profile.Spec.Name, Presets: profile.Presets, Source: profile.Source,
		})
	}
	return options, nil
}
