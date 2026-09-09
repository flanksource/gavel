package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestServeCommandRegistered(t *testing.T) {
	serve := findCommand(rootCmd, "serve")
	require.NotNil(t, serve, "`gavel serve` must be registered on rootCmd")

	// PRListOptions is bound, so the dashboard flags carry over.
	for _, name := range []string{"dev", "dev-dir", "port", "menu-bar"} {
		assert.NotNilf(t, serve.Flags().Lookup(name), "serve should expose --%s", name)
	}

	// `serve` always implies the web UI — the --ui toggle is hidden, not absent.
	ui := serve.Flags().Lookup("ui")
	require.NotNil(t, ui, "--ui flag should still be bound (from PRListOptions)")
	assert.True(t, ui.Hidden, "--ui should be hidden on the serve command")
}

func TestServeDefaultsToOrgWideWhenNoRepos(t *testing.T) {
	got := applyServeDefaults(PRListOptions{})
	assert.True(t, got.All, "serve with no repo args should default to org-wide (--all)")
	assert.True(t, got.UI, "serve should serve the web UI")
}

func TestServeRepoArgsOverrideAllDefault(t *testing.T) {
	got := applyServeDefaults(PRListOptions{Repos: []string{"flanksource/gavel"}})
	assert.False(t, got.All, "explicit repo args must not be widened to org-wide")
	assert.Equal(t, []string{"flanksource/gavel"}, got.Repos)
}

func TestServeMenuBarDoesNotForceUI(t *testing.T) {
	got := applyServeDefaults(PRListOptions{MenuBar: true})
	assert.False(t, got.UI, "--menu-bar mode should not also force the web UI")
	assert.True(t, got.All, "menu-bar serve with no repos should still default to org-wide")
}
