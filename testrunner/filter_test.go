package testrunner

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestFilterFrameworks(t *testing.T) {
	detected := []Framework{parsers.GoTest, parsers.Ginkgo, parsers.Jest, parsers.Vitest}

	t.Run("narrow to single", func(t *testing.T) {
		got, err := filterFrameworks(detected, []string{"jest"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0] != parsers.Jest {
			t.Fatalf("got %v, want [jest]", got)
		}
	})

	t.Run("go aliases resolve", func(t *testing.T) {
		for _, alias := range []string{"go", "gotest", "go-test", "go test"} {
			got, err := filterFrameworks(detected, []string{alias})
			if err != nil {
				t.Fatalf("alias %q: unexpected error: %v", alias, err)
			}
			if len(got) != 1 || got[0] != parsers.GoTest {
				t.Fatalf("alias %q: got %v, want [go test]", alias, got)
			}
		}
	})

	t.Run("dedupes", func(t *testing.T) {
		got, err := filterFrameworks(detected, []string{"jest", "jest"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 framework, got %v", got)
		}
	})

	t.Run("unknown hard-fails with known list", func(t *testing.T) {
		_, err := filterFrameworks(detected, []string{"jst"})
		if err == nil {
			t.Fatal("expected error for typo")
		}
		if !strings.Contains(err.Error(), "unknown framework") {
			t.Errorf("error should mention unknown framework, got %q", err)
		}
		if !strings.Contains(err.Error(), "jest") {
			t.Errorf("error should list known frameworks, got %q", err)
		}
	})

	t.Run("not-detected hard-fails", func(t *testing.T) {
		_, err := filterFrameworks(detected, []string{"playwright"})
		if err == nil {
			t.Fatal("expected error when requested framework isn't detected")
		}
		if !strings.Contains(err.Error(), "not detected") {
			t.Errorf("error should mention not detected, got %q", err)
		}
	})
}
