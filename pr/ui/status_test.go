package ui

import (
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
)

func TestFilterIgnoredOrgs(t *testing.T) {
	all := []github.Org{
		{Login: "alpha"},
		{Login: "bravo"},
		{Login: "charlie"},
	}

	t.Run("empty ignored is a passthrough", func(t *testing.T) {
		got := filterIgnoredOrgs(all, nil)
		assert.Equal(t, all, got)
	})

	t.Run("drops matching logins", func(t *testing.T) {
		got := filterIgnoredOrgs(all, []string{"bravo"})
		assert.Equal(t, []github.Org{{Login: "alpha"}, {Login: "charlie"}}, got)
	})

	t.Run("preserves order", func(t *testing.T) {
		got := filterIgnoredOrgs(all, []string{"charlie"})
		assert.Equal(t, []github.Org{{Login: "alpha"}, {Login: "bravo"}}, got)
	})

	t.Run("ignore-all returns empty (not nil)", func(t *testing.T) {
		got := filterIgnoredOrgs(all, []string{"alpha", "bravo", "charlie"})
		assert.Equal(t, []github.Org{}, got, "downstream JSON encoders render empty slice as [], nil as null")
	})

	t.Run("unknown logins in ignored are harmless", func(t *testing.T) {
		got := filterIgnoredOrgs(all, []string{"does-not-exist"})
		assert.Equal(t, all, got)
	})
}
