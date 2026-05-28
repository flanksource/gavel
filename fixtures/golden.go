package fixtures

import (
	"fmt"
	"os"

	"github.com/pmezard/go-difflib/difflib"
)

// UnifiedDiff returns a 3-line-context unified diff between want and
// got. An empty result means the two are identical.
func UnifiedDiff(want, got, wantLabel, gotLabel string) string {
	if want == got {
		return ""
	}
	out, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(want),
		B:        difflib.SplitLines(got),
		FromFile: wantLabel,
		ToFile:   gotLabel,
		Context:  3,
	})
	if err != nil {
		return fmt.Sprintf("(diff failed: %v)\n--- want\n%s\n+++ got\n%s", err, want, got)
	}
	return out
}

// WriteGolden replaces the contents of path with got. Used by
// --update-golden mode when an @file expectation does not match.
func WriteGolden(path, got string) error {
	return os.WriteFile(path, []byte(got), 0o644)
}
