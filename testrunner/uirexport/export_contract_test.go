package uirexport

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildUsesV1UnlessAffectedTestMetadataIsRequested(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "go.mod"), []byte("module example.com/export\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := Options{WorkDir: repo, StartedAt: time.Now(), FinishedAt: time.Now(), Provenance: Provenance{GitCommit: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	legacy, err := Build(nil, base)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.Version != LegacyContractVersion {
		t.Fatalf("ordinary export version = %d", legacy.Version)
	}
	base.ContractVersion = ContractVersion
	base.SelectionMode = "full"
	base.Inventory = []InventoryTest{{ID: "id", Framework: "go test", Package: "example.com/export", Name: "TestOne"}}
	affected, err := Build(nil, base)
	if err != nil {
		t.Fatal(err)
	}
	if affected.Version != ContractVersion || len(affected.Runs[0].Inventory) != 1 {
		t.Fatalf("affected export = %+v", affected)
	}
}
