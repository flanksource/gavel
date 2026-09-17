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

func todoRunPresetPreviewFor(presets *runtimeprofiles.PresetResolution) []runtimeprofiles.Preset {
	if presets == nil {
		return nil
	}
	return presets.Presets
}

func todoRunPresetOptions(ctx context.Context, host *lifecycle.Host) ([]runtimeprofiles.Preset, error) {
	catalog, err := host.RuntimeCatalog(ctx)
	if err != nil {
		return nil, fmt.Errorf("load runtime presets: %w", err)
	}
	host.Catalog = func(context.Context) (*runtimeprofiles.Catalog, error) { return catalog, nil }
	presets, err := catalog.ListPresets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtime presets: %w", err)
	}
	return presets, nil
}
