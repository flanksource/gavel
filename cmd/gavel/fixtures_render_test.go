package main

import (
	"testing"

	clickytask "github.com/flanksource/clicky/task"
)

func TestFixtureLiveRendererClearsTheFinalFrame(t *testing.T) {
	var renderer clickytask.LiveRenderer = fixtureLiveRenderer{}
	if rendered := renderer.RenderFinal(nil); !rendered.IsEmpty() {
		t.Fatalf("fixture final renderer must be empty, got %q", rendered.String())
	}
}
