package uirplan

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestInventoryRecognizesTestingAliasesAndIncludesPackagesWithoutTests(t *testing.T) {
	repo := t.TempDir()
	files := map[string]string{
		"go.mod":       "module example.com/inventory\n\ngo 1.22\n",
		"pkg/value.go": "package pkg\n",
		"pkg/value_test.go": `package pkg
import alias "testing"
func TestAlias(t *alias.T) {}
type fake struct{}
func TestFake(t *fake) {}
`,
		"pkg/dot_test.go": `package pkg
import . "testing"
func TestDot(t *T) {}
`,
		"notests/value.go": "package notests\n",
	}
	for name, content := range files {
		path := filepath.Join(repo, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Inventory(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, test := range got.Tests {
		names = append(names, test.Name)
	}
	if !slices.Equal(names, []string{"TestAlias", "TestDot"}) {
		t.Fatalf("inventory names = %v", names)
	}
	if !slices.Contains(got.Targets, "./notests") {
		t.Fatalf("targets omit package without tests: %v", got.Targets)
	}
}

func TestExecuteRejectsUnknownSkippableInventoryID(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper is a shell script")
	}
	binary := filepath.Join(t.TempDir(), "uir")
	script := `#!/bin/sh
cat <<'EOF'
{"version":2,"plan_id":"plan-1","inventory_digest":"inventory-1","mode":"selected","skippable":["go test\u0000example.com/module/pkg\u0000TestNotInventoried"]}
EOF
`
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := Execute(context.Background(), binary, t.TempDir(), Request{
		Version: ContractVersion,
		Inventory: []Test{{
			ID:      "go test\x00example.com/module/pkg\x00TestKnown",
			Package: "example.com/module/pkg",
			Name:    "TestKnown",
		}},
	})
	if err == nil || err.Error() != `uir tests plan returned unknown inventory id "go test\x00example.com/module/pkg\x00TestNotInventoried"` {
		t.Fatalf("Execute error = %v", err)
	}
}

func TestExecuteRejectsSkippableIDsInRunAll(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper is a shell script")
	}
	binary := filepath.Join(t.TempDir(), "uir")
	script := `#!/bin/sh
cat <<'EOF'
{"version":2,"plan_id":"plan-1","inventory_digest":"inventory-1","mode":"run_all","skippable":["known"]}
EOF
`
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Execute(context.Background(), binary, t.TempDir(), Request{Version: ContractVersion, Inventory: []Test{{ID: "known"}}})
	if err == nil || err.Error() != "uir tests plan returned an invalid response" {
		t.Fatalf("Execute error = %v", err)
	}
}
