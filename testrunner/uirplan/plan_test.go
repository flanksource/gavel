package uirplan

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExecuteRejectsUnknownSkippableInventoryID(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper is a shell script")
	}
	binary := filepath.Join(t.TempDir(), "uir")
	script := `#!/bin/sh
cat <<'EOF'
{"version":1,"plan_id":"plan-1","inventory_digest":"inventory-1","mode":"selected","skippable":["go test\u0000example.com/module/pkg\u0000TestNotInventoried"]}
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
