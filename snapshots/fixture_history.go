package snapshots

import (
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func historyIdentity(snap testui.Snapshot) ([]string, []string) {
	frameworks, files := map[string]bool{}, map[string]bool{}
	workDir := ""
	if snap.Metadata != nil {
		for _, framework := range snap.Metadata.Frameworks {
			frameworks[framework] = true
		}
		workDir, _ = snap.Metadata.Args["work_dir"].(string)
		switch paths := snap.Metadata.Args["fixture_files"].(type) {
		case []string:
			for _, path := range paths {
				files[path] = true
			}
		case []any:
			for _, path := range paths {
				if value, ok := path.(string); ok {
					files[value] = true
				}
			}
		}
		if snap.Metadata.Kind == "fixtures" || snap.Metadata.Args["fixtures"] == true {
			frameworks["fixture"] = true
		}
	}
	collectHistoryIdentity(snap.Tests, frameworks, files)
	delete(files, "")
	if len(files) > 0 {
		frameworks["fixture"] = true
	}
	normalized := map[string]bool{}
	for path := range files {
		if filepath.IsAbs(path) && filepath.IsAbs(workDir) {
			if relative, err := filepath.Rel(workDir, path); err == nil {
				path = relative
			}
		}
		normalized[filepath.ToSlash(filepath.Clean(path))] = true
	}
	delete(frameworks, "")
	return sortedHistoryKeys(frameworks), sortedHistoryKeys(normalized)
}

func collectHistoryIdentity(tests []parsers.Test, frameworks, files map[string]bool) {
	for _, test := range tests {
		frameworks[string(test.Framework)] = true
		if test.Framework == parsers.Framework("fixture") {
			files[test.File] = true
		}
		collectHistoryIdentity(test.Children, frameworks, files)
	}
}

func sortedHistoryKeys(values map[string]bool) []string {
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
