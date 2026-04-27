package main

import (
	"testing"

	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestCollectFailedReturnsLeafFailuresOnly(t *testing.T) {
	tree := []parsers.Test{{
		Name: "pkg",
		Children: []parsers.Test{
			{Name: "passing", Passed: true},
			{Name: "failing", Failed: true, Message: "bad"},
		},
	}, {
		Name:   "standalone-fail",
		Failed: true,
	}}

	got := collectFailed(tree)
	if len(got) != 2 {
		t.Fatalf("want 2 failing leaves, got %d", len(got))
	}
	names := []string{got[0].Name, got[1].Name}
	wantNames := map[string]bool{"failing": true, "standalone-fail": true}
	for _, n := range names {
		if !wantNames[n] {
			t.Errorf("unexpected leaf %q", n)
		}
	}
}

func TestPrintFailureDetailsRespectsShowStdoutOnFailure(t *testing.T) {
	// Exercise the filter contract: OnFailure keeps stdout on a failing
	// test (the bug), Never drops it. Since printFailureDetails operates
	// on a copy via range, we test the filter itself via ShouldShow.
	if !testrunner.OutputOnFailure.ShouldShow(true) {
		t.Error("OnFailure must show streams for failing tests")
	}
	if testrunner.OutputOnFailure.ShouldShow(false) {
		t.Error("OnFailure must not show streams for passing tests")
	}
	if !testrunner.OutputAlways.ShouldShow(false) {
		t.Error("Always must show streams even on pass")
	}
	if testrunner.OutputNever.ShouldShow(true) {
		t.Error("Never must not show streams even on fail")
	}
}
