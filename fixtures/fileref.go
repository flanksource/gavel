package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileRef represents the resolution of a fixture cell value that may
// optionally reference an external file via an "@" prefix.
//
//	"@testdata/golden.md"  → IsFile=true,  Contents=<file body>
//	"some literal value"   → IsFile=false, Raw="some literal value"
//
// A literal "@" character is preserved by escaping with a leading
// backslash: "\@literal" becomes Raw="@literal".
type FileRef struct {
	Raw      string
	Path     string
	Contents string
	IsFile   bool
}

// ResolveFileRef parses value and, when it begins with "@", reads the
// referenced file. Relative paths are joined with sourceDir (the
// directory of the fixture markdown file).
//
// Globs are NOT expanded here — per-row glob expansion happens earlier
// in the parser (see expand.go); by the time ResolveFileRef runs each
// row has a concrete file path.
func ResolveFileRef(sourceDir, value string) (FileRef, error) {
	if strings.HasPrefix(value, `\@`) {
		return FileRef{Raw: value[1:]}, nil
	}
	if !strings.HasPrefix(value, "@") {
		return FileRef{Raw: value}, nil
	}

	path := strings.TrimPrefix(value, "@")
	if !filepath.IsAbs(path) {
		path = filepath.Join(sourceDir, path)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return FileRef{}, fmt.Errorf("read file ref %q: %w", value, err)
	}

	return FileRef{
		Raw:      value,
		Path:     path,
		Contents: string(contents),
		IsFile:   true,
	}, nil
}
