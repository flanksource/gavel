package verify

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkLoadGavelConfig measures the config load procfile.gather performs for
// every project on every status poll, against an empty directory and against a
// checkout that carries a .gavel.yaml (so the git-root merge is exercised).
// "cold" clears the cache before every load — the full read, parse, merge and
// validate — while "warm" is the steady state of an unchanged chain: read every
// file, compare bytes, deep-copy the cached result.
func BenchmarkLoadGavelConfig(b *testing.B) {
	for _, target := range []struct {
		name string
		dir  func(*testing.B) string
	}{
		{"empty", func(b *testing.B) string { return b.TempDir() }},
		{"repo", benchRepoDir},
	} {
		b.Run(target.name+"/cold", func(b *testing.B) {
			dir := target.dir(b)
			for b.Loop() {
				resetGavelConfigCache()
				if _, err := LoadGavelConfig(dir); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(target.name+"/warm", func(b *testing.B) {
			dir := target.dir(b)
			if _, err := LoadGavelConfig(dir); err != nil {
				b.Fatal(err)
			}
			for b.Loop() {
				if _, err := LoadGavelConfig(dir); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkDefaultGavelConfig isolates the baseline construction cost from the
// file reads and merges layered on top of it.
func BenchmarkDefaultGavelConfig(b *testing.B) {
	for b.Loop() {
		_ = DefaultGavelConfig()
	}
}

func resetGavelConfigCache() {
	gavelConfigCache.Lock()
	defer gavelConfigCache.Unlock()
	clear(gavelConfigCache.entries)
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
