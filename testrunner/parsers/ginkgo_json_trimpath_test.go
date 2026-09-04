package parsers

import "testing"

// Builds run with -trimpath report ginkgo CodeLocations as import paths rather
// than filesystem paths. relToWorkDir must strip the module prefix so the
// result stays relative to the work dir, as it already does for absolute paths.
func TestRelToWorkDirStripsModulePath(t *testing.T) {
	const (
		workDir    = "/Users/dev/src/example.com/proj"
		modulePath = "example.com/proj"
	)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "trimpath location loses the module prefix",
			input: modulePath + "/testrunner/runner_test.go",
			want:  "testrunner/runner_test.go",
		},
		{
			name:  "absolute location is made relative to the work dir",
			input: workDir + "/testrunner/runner_test.go",
			want:  "testrunner/runner_test.go",
		},
		{
			name:  "a path from another module is left alone",
			input: "example.com/other/pkg/thing_test.go",
			want:  "example.com/other/pkg/thing_test.go",
		},
		{
			name:  "the module path alone is not treated as a prefix",
			input: modulePath,
			want:  modulePath,
		},
		{
			name:  "empty stays empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &GinkgoJSON{workDir: workDir, modulePath: modulePath}
			if got := p.relToWorkDir(tt.input); got != tt.want {
				t.Errorf("relToWorkDir(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Without a resolvable go.mod there is no prefix to strip, and import-path-shaped
// locations must pass through untouched rather than being mangled.
func TestRelToWorkDirWithoutModulePath(t *testing.T) {
	p := &GinkgoJSON{workDir: "/Users/dev/src/example.com/proj"}
	const input = "example.com/proj/testrunner/runner_test.go"
	if got := p.relToWorkDir(input); got != input {
		t.Errorf("relToWorkDir(%q) = %q, want it unchanged", input, got)
	}
}
