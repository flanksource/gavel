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

	// A Ginkgo suite beside a plain-testing benchmark: the suite's bootstrap is
	// the package's runnable test, so benchmarks are opt-in and never displace it.
	ginkgoSuite := filepath.Join(tmpDir, "ginkgosuite")
	if err := os.MkdirAll(ginkgoSuite, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ginkgoSuite, "suite_test.go"), []byte(`package ginkgosuite_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestSuite(t *testing.T) { RunSpecs(t, "suite") }
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ginkgoSuite, "bench_test.go"), []byte(`package ginkgosuite_test

import "testing"

func BenchmarkZ(b *testing.B) {}
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
			name:  "Ginkgo suite with benchmarks keeps its specs runnable under an explicit pattern",
			bench: "^$",
			pkg:   "./ginkgosuite",
			want:  []string{"-bench=^$"},
		},
		{
			name:  "Ginkgo suite with benchmarks is not treated as bench-only",
			bench: "",
			pkg:   "./ginkgosuite",
			want:  nil,
		},
		{
			name:  "Ginkgo suite with bench=true runs the benchmark beside its specs",
			bench: "true",
			pkg:   "./ginkgosuite",
			want:  []string{"-bench=."},
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
