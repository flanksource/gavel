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
