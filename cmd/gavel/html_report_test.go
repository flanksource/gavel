package main

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestRenderGavelHTMLReportShowsPackageTableAndFailureTrace(t *testing.T) {
	s := testui.Snapshot{Tests: []parsers.Test{{
		Name:        "./pkg/foo",
		PackagePath: "./pkg/foo",
		Framework:   parsers.Ginkgo,
		Children: parsers.Tests{{
			Name:        "does a thing",
			PackagePath: "./pkg/foo",
			Framework:   parsers.Ginkgo,
			Failed:      true,
			Duration:    1500 * time.Millisecond,
			Message:     "expected true",
			File:        "foo_test.go",
			Line:        42,
		}, {
			Name:        "passes",
			PackagePath: "./pkg/foo",
			Framework:   parsers.Ginkgo,
			Passed:      true,
			Duration:    500 * time.Millisecond,
		}},
	}}}

	html := renderGavelHTMLReport(s)
	for _, want := range []string{
		"<h2>Packages</h2>",
		"./pkg/foo",
		"ginkgo",
		"expected true",
		"foo_test.go:42",
		"details",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("HTML missing %q:\n%s", want, html)
		}
	}
}
