package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/status"
)

func enrichProjectLinkedDeps(workDir string, files []projectFileStatus) error {
	for i := range files {
		file := &files[i]
		file.LinkedDependencies = nil
		switch filepath.Base(file.Path) {
		case "go.mod", "go.work", "package.json", "pnpm-lock.yaml":
		default:
			continue
		}
		if file.State == status.StateConflict || file.WorkKind == status.KindDeleted ||
			(file.StagedKind == status.KindDeleted && file.WorkKind == status.KindUnknown) {
			continue
		}
		var violations []commit.LinkedDepViolation
		var err error
		if file.State == status.StateStaged {
			violations, err = commit.EvaluateLinkedDeps(workDir, workDir, []string{file.Path}, nil)
		} else {
			var blob []byte
			blob, err = os.ReadFile(filepath.Join(workDir, file.Path))
			if err == nil {
				violations, err = commit.EvaluateManifestLinkedDeps(workDir, file.Path, blob)
			}
		}
		if err != nil {
			return fmt.Errorf("check linked dependencies in %s: %w", file.Path, err)
		}
		for _, violation := range violations {
			file.LinkedDependencies = append(file.LinkedDependencies, projectLinkedDependency{Name: violation.Name, Target: violation.Target})
		}
	}
	return nil
}
