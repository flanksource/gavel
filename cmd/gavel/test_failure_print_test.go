package main

import (
	"testing"

	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestCollectLeavesReturnsLeafFailuresOnly(t *testing.T) {
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

	got := collectLeaves(tree, func(t parsers.Test) bool { return t.Failed })
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

func TestPassingLeavesForDisplaySelectsPassesAndMasksStdout(t *testing.T) {
	tree := []parsers.Test{{
		Name: "pkg",
		Children: []parsers.Test{
			{Name: "passing", Passed: true, Stdout: "hello", Stderr: "warn"},
			{Name: "failing", Failed: true, Stdout: "boom"},
			{Name: "skipped", Skipped: true},
		},
	}}

	// Default OnFailure: passing leaf is selected but its streams are masked,
	// since a pass evaluates ShouldShow(false) == false.
	masked := passingLeavesForDisplay(tree, testrunner.OutputOnFailure, testrunner.OutputOnFailure)
	if len(masked) != 1 || masked[0].Name != "passing" {
		t.Fatalf("want only the passing leaf, got %+v", masked)
	}
	if masked[0].Stdout != "" || masked[0].Stderr != "" {
		t.Errorf("OnFailure must mask streams for a passing test, got stdout=%q stderr=%q", masked[0].Stdout, masked[0].Stderr)
	}

	// Always: the passing leaf keeps its captured streams.
	kept := passingLeavesForDisplay(tree, testrunner.OutputAlways, testrunner.OutputAlways)
	if len(kept) != 1 || kept[0].Stdout != "hello" || kept[0].Stderr != "warn" {
		t.Fatalf("Always must keep passing streams, got %+v", kept)
	}
}

func TestOutputModeShouldShowContract(t *testing.T) {
	// Exercise the OutputMode filter contract that the failure-detail
	// printers rely on: OnFailure keeps streams for a failing test and
	// drops them for a passing one; Always always shows; Never never does.
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
