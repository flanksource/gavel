package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// expandGlobInRows walks the fixture tree and rewrites any TestNode
// whose Expected.Stdout or Expected.Stderr cell is an "@<glob>"
// reference into N rows, one per glob match. Template variables
// (file, filename, dir, absfile, absdir, basename, ext) are set on
// each expanded row exactly like expandFixturesForFiles does for the
// frontmatter `files:` field.
//
// Only the first glob cell drives expansion (stdout, then stderr).
// Other cells that reference the same stem use {{.filename}} style
// templating and are filled by gomplate at exec time. Non-glob @file
// references are left untouched; they are loaded later by
// ResolveFileRef during Expectations.Evaluate.
func expandGlobInRows(root *FixtureNode, sourceDir string) error {
	if root == nil {
		return nil
	}
	for i := 0; i < len(root.Children); i++ {
		child := root.Children[i]
		if child.Test != nil {
			expanded, err := expandTestNode(child, sourceDir)
			if err != nil {
				return err
			}
			if expanded != nil {
				root.Children = append(root.Children[:i], append(expanded, root.Children[i+1:]...)...)
				i += len(expanded) - 1
			}
			continue
		}
		if err := expandGlobInRows(child, sourceDir); err != nil {
			return err
		}
	}
	return nil
}

func expandTestNode(child *FixtureNode, sourceDir string) ([]*FixtureNode, error) {
	pattern, _ := pickGlobCell(child.Test)
	if pattern == "" {
		return nil, nil
	}
	matches, err := resolveRowGlob(sourceDir, pattern)
	if err != nil {
		return nil, fmt.Errorf("row %q: %w", child.Test.Name, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("row %q: glob %q matched no files", child.Test.Name, pattern)
	}

	out := make([]*FixtureNode, 0, len(matches))
	for _, absFile := range matches {
		clone := *child
		test := *child.Test
		vars := globTemplateVars(absFile, sourceDir)
		test.Name = fmt.Sprintf("%s [%s]", test.Name, vars["filename"])
		if test.TemplateVars == nil {
			test.TemplateVars = map[string]any{}
		}
		for k, v := range vars {
			test.TemplateVars[k] = v
		}
		test.Expected.Stdout = applyGlobMatch(test.Expected.Stdout, vars)
		test.Expected.Stderr = applyGlobMatch(test.Expected.Stderr, vars)
		clone.Test = &test
		out = append(out, &clone)
	}
	return out, nil
}

// pickGlobCell returns the first @-glob cell on a row, in priority
// order: Stdout, Stderr. Properties are templated later by gomplate.
func pickGlobCell(t *FixtureTest) (pattern, cell string) {
	if isGlobRef(t.Expected.Stdout) {
		return t.Expected.Stdout, "stdout"
	}
	if isGlobRef(t.Expected.Stderr) {
		return t.Expected.Stderr, "stderr"
	}
	return "", ""
}

func isGlobRef(v string) bool {
	if !strings.HasPrefix(v, "@") || strings.HasPrefix(v, `\@`) {
		return false
	}
	rest := strings.TrimPrefix(v, "@")
	return strings.ContainsAny(rest, "*?[")
}

func resolveRowGlob(sourceDir, ref string) ([]string, error) {
	pattern := strings.TrimPrefix(ref, "@")
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(sourceDir, pattern)
	}
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob %q: %w", ref, err)
	}
	out := matches[:0]
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || info.IsDir() {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func globTemplateVars(absFile, sourceDir string) map[string]any {
	fileDir := filepath.Dir(absFile)
	fileName := filepath.Base(absFile)
	stem := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	relFile, _ := filepath.Rel(sourceDir, absFile)
	relDir, _ := filepath.Rel(sourceDir, fileDir)
	return map[string]any{
		"file":     relFile,
		"filename": stem,
		"dir":      relDir,
		"absfile":  absFile,
		"absdir":   fileDir,
		"basename": fileName,
		"ext":      filepath.Ext(fileName),
	}
}

// applyGlobMatch rewrites an @-glob cell to @<absFile> after
// expansion so ResolveFileRef gets a concrete path. Non-glob cells
// pass through unchanged.
func applyGlobMatch(value string, vars map[string]any) string {
	if !isGlobRef(value) {
		return value
	}
	abs, _ := vars["absfile"].(string)
	return "@" + abs
}
