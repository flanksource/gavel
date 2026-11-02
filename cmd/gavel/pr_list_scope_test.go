package main

import (
	"reflect"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/ui"
)

func TestScopePRSearch(t *testing.T) {
	startupRepos := []string{"acme/startup"}
	cases := map[string]struct {
		startup github.PRSearchOptions
		cfg     ui.SearchConfig
		want    github.PRSearchOptions
	}{
		"cleared repos return an org-wide serve to the whole org": {
			startup: github.PRSearchOptions{All: true, Repos: startupRepos},
			cfg:     ui.SearchConfig{IgnoredOrgs: []string{"noise"}},
			want:    github.PRSearchOptions{All: true, IgnoredOrgs: []string{"noise"}},
		},
		"selected repos replace the startup repos": {
			startup: github.PRSearchOptions{All: true, Repos: startupRepos},
			cfg:     ui.SearchConfig{Repos: []string{"acme/widget"}},
			want:    github.PRSearchOptions{Repos: []string{"acme/widget"}},
		},
		"a configured org overrides the startup org": {
			startup: github.PRSearchOptions{All: true, Org: "acme"},
			cfg:     ui.SearchConfig{Org: "other"},
			want:    github.PRSearchOptions{All: true, Org: "other"},
		},
	}
	for name, c := range cases {
		if got := scopePRSearch(c.startup, c.cfg); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: scopePRSearch() = %+v, want %+v", name, got, c.want)
		}
	}
}
