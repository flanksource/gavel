package verify

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkLoadGavelConfig measures the uncached per-call cost of the config
// load that procfile.gather performs for every project on every status poll.
func BenchmarkLoadGavelConfig(b *testing.B) {
	dir := b.TempDir()
	for b.Loop() {
		if _, err := LoadGavelConfig(dir); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDefaultGavelConfig isolates the baseline construction cost from the
// file reads and merges layered on top of it.
func BenchmarkDefaultGavelConfig(b *testing.B) {
	for b.Loop() {
		_ = DefaultGavelConfig()
	}
}

// BenchmarkLoadGavelConfigRealRepo measures the cost against a checkout that
// actually carries a .gavel.yaml, so the git-root merge is exercised — the
// shape procfile.gather hits for every configured project on every poll.
func BenchmarkLoadGavelConfigRealRepo(b *testing.B) {
	dir := benchRepoDir(b)
	for b.Loop() {
		if _, err := LoadGavelConfig(dir); err != nil {
			b.Fatal(err)
		}
	}
}

func benchRepoDir(b *testing.B) string {
	b.Helper()
	dir, err := filepath.Abs("..")
	if err != nil {
		b.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gavel.yaml")); err != nil {
		b.Skipf("no .gavel.yaml at %s: %v", dir, err)
	}
	return dir
}
