package testrunner

import (
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestTestStreamerIgnoresSendsAfterDone(t *testing.T) {
	updates := make(chan []parsers.Test, 4)
	streamer := NewTestStreamer(updates)

	streamer.SetPackageOutline([]parsers.Test{{
		Name:        "./pkg/foo",
		PackagePath: "./pkg/foo",
		Framework:   parsers.GoTest,
		Pending:     true,
	}})
	streamer.Done()

	streamer.SetPackageOutline([]parsers.Test{{
		Name:        "./pkg/bar",
		PackagePath: "./pkg/bar",
		Framework:   parsers.GoTest,
		Pending:     true,
	}})
	streamer.UpdateFixtures([]parsers.Test{{Name: "fixture", Framework: "fixture"}})
	streamer.CompletePackage("./pkg/foo", parsers.GoTest, nil)
	streamer.Done()
}

func TestSharedTestStreamerLeavesSharedChannelOpen(t *testing.T) {
	updates := make(chan []parsers.Test, 4)
	streamer := NewSharedTestStreamer(updates)

	streamer.SetPackageOutline([]parsers.Test{{
		Name:        "./pkg/foo",
		PackagePath: "./pkg/foo",
		Framework:   parsers.GoTest,
		Pending:     true,
	}})
	streamer.Done()

	select {
	case _, ok := <-updates:
		if !ok {
			t.Fatal("shared streamer closed the shared updates channel")
		}
	default:
	}

	select {
	case updates <- nil:
	default:
		t.Fatal("expected shared updates channel to remain writable after Done")
	}
}
