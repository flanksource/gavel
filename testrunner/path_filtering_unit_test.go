package testrunner

import (
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
