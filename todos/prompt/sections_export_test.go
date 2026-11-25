package prompt

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

// Sections is the renderer both the lifecycle prompts and the one-shot merge
// prompt use. These assert the two properties a second caller depends on:
// numbering only when there is more than one todo, and a verbatim fixture when
// the caller is going to rewrite it.
func TestSectionsNumbersOnlyGroups(t *testing.T) {
	one, err := Sections([]*types.TODO{newTestTODO("solo", "Fix it")}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(one, "## 1. ") {
		t.Fatalf("a single todo must not be numbered:\n%s", one)
	}

	group, err := Sections([]*types.TODO{newTestTODO("a", "First"), newTestTODO("b", "Second")}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## 1. a", "## 2. b"} {
		if !strings.Contains(group, want) {
			t.Fatalf("group section is missing %q:\n%s", want, group)
		}
	}
}

func TestSectionsRawFixtureShowsTheFixtureSource(t *testing.T) {
	todo := newTestTODO("a", "First")
	todo.VerificationMarkdown = "```yaml test\npackages: ./todos\n```"

	projected, err := Sections([]*types.TODO{todo}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(projected, "packages: ./todos") {
		t.Fatalf("the command projection must not leak the fixture source:\n%s", projected)
	}

	raw, err := Sections([]*types.TODO{todo}, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "packages: ./todos") {
		t.Fatalf("a rewriting caller must see the fixture verbatim:\n%s", raw)
	}
}
