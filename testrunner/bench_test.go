package testrunner

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/runners"
)

func TestAugmentBenchArgs(t *testing.T) {
	tmpDir := t.TempDir()

	benchOnly := filepath.Join(tmpDir, "benchonly")
	if err := os.MkdirAll(benchOnly, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(benchOnly, "main_test.go"), []byte(`package benchonly

import "testing"

func TestMain(m *testing.M) { m.Run() }
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(benchOnly, "bench_test.go"), []byte(`package benchonly

import "testing"

func BenchmarkFoo(b *testing.B) {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	mixed := filepath.Join(tmpDir, "mixed")
	if err := os.MkdirAll(mixed, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mixed, "x_test.go"), []byte(`package mixed

import "testing"

func TestX(t *testing.T) {}
func BenchmarkY(b *testing.B) {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	noTests := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(noTests, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(noTests, "x_test.go"), []byte(`package empty

import "testing"

func TestX(t *testing.T) {}
`), 0644); err != nil {
		t.Fatal(err)
	}

	runner := runners.NewGoTest(tmpDir)
	newOrchestrator := func(bench string) *TestOrchestrator {
		return &TestOrchestrator{
			RunOptions: RunOptions{WorkDir: tmpDir, Bench: bench},
		}
	}

	tests := []struct {
		name      string
		bench     string
		pkg       string
		extraArgs []string
		want      []string
	}{
		{
			name:  "bench-only package auto-enables bench with -run=^$",
			bench: "",
			pkg:   "./benchonly",
			want:  []string{"-bench=.", "-run=^$"},
		},
		{
			name:  "package with tests only is untouched when bench flag off",
			bench: "",
			pkg:   "./empty",
			want:  nil,
		},
		{
			name:  "package with tests only runs benches when flag explicitly enabled",
			bench: ".",
			pkg:   "./empty",
			want:  nil, // no benchmarks in this package; nothing to add
		},
		{
			name:  "mixed package with bench=true runs bench without -run=^$",
			bench: "true",
			pkg:   "./mixed",
			want:  []string{"-bench=."},
		},
		{
			name:  "mixed package with custom pattern",
			bench: "BenchmarkY",
			pkg:   "./mixed",
			want:  []string{"-bench=BenchmarkY"},
		},
		{
			name:      "user-supplied -bench in extraArgs is preserved as-is",
			bench:     "",
			pkg:       "./benchonly",
			extraArgs: []string{"-bench=Custom"},
			want:      []string{"-bench=Custom"},
		},
		{
			name:      "user-supplied -run prevents injection of -run=^$",
			bench:     ".",
			pkg:       "./benchonly",
			extraArgs: []string{"-run=^TestMain$"},
			want:      []string{"-run=^TestMain$", "-bench=."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newOrchestrator(tt.bench)
			got := o.augmentBenchArgs(runner, tt.pkg, tt.extraArgs)
			if !slices.Equal(got, tt.want) {
				t.Errorf("augmentBenchArgs(%q, bench=%q, extra=%v) = %v, want %v",
					tt.pkg, tt.bench, tt.extraArgs, got, tt.want)
			}
		})
	}

	t.Run("bench-only package produces args compatible with go test", func(t *testing.T) {
		o := newOrchestrator("")
		got := o.augmentBenchArgs(runner, "./benchonly", nil)
		joined := strings.Join(got, " ")
		if !strings.Contains(joined, "-bench=") {
			t.Errorf("expected -bench= in args, got %q", joined)
		}
	})
}
