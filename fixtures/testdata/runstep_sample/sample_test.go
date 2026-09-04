package sample

import "testing"

// A tiny package the runner-step fixtures point at. It lives under testdata so
// the go tool never discovers it during a normal `go test ./...` of gavel; the
// `yaml test` fixture step runs it explicitly via the test engine.

func TestSampleAdd(t *testing.T) {
	if 2+2 != 4 {
		t.Fatal("arithmetic is broken")
	}
}

func TestSampleConcat(t *testing.T) {
	if "a"+"b" != "ab" {
		t.Fatal("concatenation is broken")
	}
}
