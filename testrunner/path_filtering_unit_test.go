package testrunner

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestApplyIgnorePatterns(t *testing.T) {
	tests := []struct {
		name   string
		pkgs   []string
		ignore []string
		want   []string
	}{
		{
			name:   "empty ignore returns input unchanged",
			pkgs:   []string{"./api", "./bench", "./bench/sub"},
			ignore: nil,
			want:   []string{"./api", "./bench", "./bench/sub"},
		},
		{
			name:   "bare directory ignore is recursive",
			pkgs:   []string{"./api", "./bench", "./bench/sub", "./bench/deep/inner"},
			ignore: []string{"./bench"},
			want:   []string{"./api"},
		},
		{
			name:   "trailing /... is treated the same as bare dir",
			pkgs:   []string{"./api", "./hack", "./hack/foo"},
			ignore: []string{"./hack/..."},
			want:   []string{"./api"},
		},
		{
			name:   "trailing /** is treated the same as bare dir",
			pkgs:   []string{"./api", "./hack", "./hack/foo"},
			ignore: []string{"./hack/**"},
			want:   []string{"./api"},
		},
		{
			name:   "non-matching pattern is a no-op",
			pkgs:   []string{"./api", "./db"},
			ignore: []string{"./does-not-exist"},
			want:   []string{"./api", "./db"},
		},
		{
			name:   "multiple patterns compose",
			pkgs:   []string{"./api", "./bench", "./hack/x", "./specs", "./tests/e2e", "./tests/unit"},
			ignore: []string{"./bench", "./hack", "./specs", "./tests/e2e"},
			want:   []string{"./api", "./tests/unit"},
		},
		{
			name:   "exact-match without recursion does not over-match siblings",
			pkgs:   []string{"./bench", "./benchmarks"},
			ignore: []string{"./bench"},
			want:   []string{"./benchmarks"},
		},
		{
			name:   "leading ./ is normalized when missing",
			pkgs:   []string{"./bench", "./api"},
			ignore: []string{"bench"},
			want:   []string{"./api"},
		},
		{
			name:   "blank pattern is ignored",
			pkgs:   []string{"./api"},
			ignore: []string{"   "},
			want:   []string{"./api"},
		},
		{
			name:   "root pattern is rejected (would match everything)",
			pkgs:   []string{"./api", "./db"},
			ignore: []string{"."},
			want:   []string{"./api", "./db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyIgnorePatterns(tt.pkgs, tt.ignore)
			// applyIgnorePatterns may return a nil slice for fully filtered input;
			// normalize for comparison.
			if got == nil {
				got = []string{}
			}
			want := tt.want
			if want == nil {
				want = []string{}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("applyIgnorePatterns(%v, %v) = %v, want %v", tt.pkgs, tt.ignore, got, want)
			}
		})
	}
}

func TestExpandRecursiveWildcards(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty input", in: nil, want: nil},
		{name: "./... becomes empty (whole-workdir discovery)", in: []string{"./..."}, want: nil},
		{name: "... becomes empty", in: []string{"..."}, want: nil},
		{name: ". becomes empty", in: []string{"."}, want: nil},
		{
			name: "./pkg/... is preserved without the suffix",
			in:   []string{"./pkg/..."},
			want: []string{"./pkg"},
		},
		{
			name: "mixed wildcards and concrete paths",
			in:   []string{"./...", "./api", "./pkg/..."},
			want: []string{"./api", "./pkg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandRecursiveWildcards(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandRecursiveWildcards(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFilterIgnoredGroups(t *testing.T) {
	base := filepath.FromSlash("/repo")
	groups := []testGroup{
		{workDir: filepath.FromSlash("/repo")},
		{workDir: filepath.FromSlash("/repo/hack/migrate")},
		{workDir: filepath.FromSlash("/repo/hack/generate-schemas")},
		{workDir: filepath.FromSlash("/repo/sub")},
	}

	tests := []struct {
		name     string
		ignore   []string
		wantDirs []string
	}{
		{
			name:     "no ignore returns all",
			ignore:   nil,
			wantDirs: []string{"/repo", "/repo/hack/migrate", "/repo/hack/generate-schemas", "/repo/sub"},
		},
		{
			name:     "bare dir ignore strips nested-module groups",
			ignore:   []string{"./hack"},
			wantDirs: []string{"/repo", "/repo/sub"},
		},
		{
			name:     "ignore that matches the base does not strip the base group",
			ignore:   []string{"./sub"},
			wantDirs: []string{"/repo", "/repo/hack/migrate", "/repo/hack/generate-schemas"},
		},
		{
			name:     "non-matching ignore is a no-op",
			ignore:   []string{"./does-not-exist"},
			wantDirs: []string{"/repo", "/repo/hack/migrate", "/repo/hack/generate-schemas", "/repo/sub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterIgnoredGroups(base, groups, tt.ignore)
			gotDirs := make([]string, 0, len(got))
			for _, g := range got {
				gotDirs = append(gotDirs, filepath.ToSlash(g.workDir))
			}
			if !reflect.DeepEqual(gotDirs, tt.wantDirs) {
				t.Errorf("filterIgnoredGroups dirs = %v, want %v", gotDirs, tt.wantDirs)
			}
		})
	}
}
