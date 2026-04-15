package ui

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/flanksource/gavel/github"
)

func TestParseRouteRequest(t *testing.T) {
	cases := []struct {
		name       string
		target     string
		wantOK     bool
		wantPath   []string
		wantFormat string
		wantExport bool
		wantFilter prRouteFilters
	}{
		{
			name:   "root path",
			target: "/",
			wantOK: true,
		},
		{
			name:   "prs tab",
			target: "/prs",
			wantOK: true,
		},
		{
			name:     "prs with repo/number",
			target:   "/prs/flanksource/gavel/42",
			wantOK:   true,
			wantPath: []string{"flanksource", "gavel", "42"},
		},
		{
			name:       "prs json export",
			target:     "/prs.json",
			wantOK:     true,
			wantFormat: "json",
			wantExport: true,
		},
		{
			name:       "single PR markdown export",
			target:     "/prs/flanksource/gavel/42.md",
			wantOK:     true,
			wantPath:   []string{"flanksource", "gavel", "42"},
			wantFormat: "markdown",
			wantExport: true,
		},
		{
			name:   "filters in query string",
			target: "/prs?state=open,draft&checks=failing&repos=flanksource/gavel&authors=alice,bob",
			wantOK: true,
			wantFilter: prRouteFilters{
				State:   []string{"open", "draft"},
				Checks:  []string{"failing"},
				Repos:   []string{"flanksource/gavel"},
				Authors: []string{"alice", "bob"},
			},
		},
		{
			name:   "unknown tab",
			target: "/lint",
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tc.target, nil)
			req, ok := parseRouteRequest(r)
			if ok != tc.wantOK {
				t.Fatalf("ok: got %v want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if !reflect.DeepEqual(req.NodePath, tc.wantPath) && !(len(req.NodePath) == 0 && len(tc.wantPath) == 0) {
				t.Errorf("NodePath: got %v want %v", req.NodePath, tc.wantPath)
			}
			if req.Format != tc.wantFormat {
				t.Errorf("Format: got %q want %q", req.Format, tc.wantFormat)
			}
			if req.IsExport != tc.wantExport {
				t.Errorf("IsExport: got %v want %v", req.IsExport, tc.wantExport)
			}
			if !reflect.DeepEqual(req.PRFilters, tc.wantFilter) &&
				!(isEmptyFilters(req.PRFilters) && isEmptyFilters(tc.wantFilter)) {
				t.Errorf("Filters: got %+v want %+v", req.PRFilters, tc.wantFilter)
			}
		})
	}
}

func isEmptyFilters(f prRouteFilters) bool {
	return len(f.State) == 0 && len(f.Checks) == 0 && len(f.Repos) == 0 && len(f.Authors) == 0
}

func TestFilterPRNodes(t *testing.T) {
	prs := []*PRViewNode{
		{Repo: "org/a", Number: 1, Author: "alice", State: "OPEN", CheckStatus: &github.CheckSummary{Failed: 1}},
		{Repo: "org/b", Number: 2, Author: "bob", State: "OPEN", IsDraft: true, CheckStatus: &github.CheckSummary{}},
		{Repo: "org/a", Number: 3, Author: "alice", State: "MERGED"},
	}

	got := filterPRNodes(prs, prRouteFilters{State: []string{"draft"}})
	if len(got) != 1 || got[0].Number != 2 {
		t.Errorf("draft filter: got %+v", got)
	}

	got = filterPRNodes(prs, prRouteFilters{Checks: []string{"failing"}})
	if len(got) != 1 || got[0].Number != 1 {
		t.Errorf("failing checks filter: got %+v", got)
	}

	got = filterPRNodes(prs, prRouteFilters{Repos: []string{"org/a"}})
	if len(got) != 2 {
		t.Errorf("repo filter: got %d want 2", len(got))
	}

	got = filterPRNodes(prs, prRouteFilters{Authors: []string{"bob"}})
	if len(got) != 1 || got[0].Author != "bob" {
		t.Errorf("author filter: got %+v", got)
	}
}

func TestAnnotatePRRoutePaths(t *testing.T) {
	nodes := []*PRViewNode{
		{Repo: "flanksource/gavel", Number: 42},
		{Repo: "flanksource/commons", Number: 7},
	}
	annotatePRRoutePaths(nodes)
	if nodes[0].RoutePath != "flanksource/gavel/42" {
		t.Errorf("first route path: got %q", nodes[0].RoutePath)
	}
	if nodes[1].RoutePath != "flanksource/commons/7" {
		t.Errorf("second route path: got %q", nodes[1].RoutePath)
	}
}

func TestParseRouteRequestAcceptHeader(t *testing.T) {
	t.Run("html accept renders SPA", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/prs", nil)
		r.Header.Set("Accept", "text/html,application/xhtml+xml")
		req, ok := parseRouteRequest(r)
		if !ok || req.IsExport {
			t.Errorf("expected SPA render, got ok=%v export=%v format=%q", ok, req.IsExport, req.Format)
		}
	})
	t.Run("json accept triggers export", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/prs", nil)
		r.Header.Set("Accept", "application/json")
		req, ok := parseRouteRequest(r)
		if !ok || !req.IsExport || req.Format != "json" {
			t.Errorf("expected json export, got ok=%v export=%v format=%q", ok, req.IsExport, req.Format)
		}
	})
	t.Run("markdown accept triggers export", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/prs", nil)
		r.Header.Set("Accept", "text/markdown")
		req, ok := parseRouteRequest(r)
		if !ok || !req.IsExport || req.Format != "markdown" {
			t.Errorf("expected markdown export, got ok=%v export=%v format=%q", ok, req.IsExport, req.Format)
		}
	})
}

func TestFindPRNode(t *testing.T) {
	nodes := []*PRViewNode{
		{Repo: "flanksource/gavel", Number: 42, RoutePath: "flanksource/gavel/42"},
		{Repo: "flanksource/commons", Number: 7, RoutePath: "flanksource/commons/7"},
	}
	got := findPRNode(nodes, []string{"flanksource", "gavel", "42"})
	if got == nil || got.Number != 42 {
		t.Errorf("expected to find #42, got %+v", got)
	}
	if findPRNode(nodes, []string{"unknown", "1"}) != nil {
		t.Error("expected nil for unknown path")
	}
}
