package lint

import (
	"context"
	"os"
	"path/filepath"
	"slices"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
)

func (o Options) ChangeSelection() *changeset.Selection {
	if o.Selection != nil {
		return o.Selection
	}
	return changeset.NewSelection(changeset.Options{Dir: o.WorkDir, Changed: o.Changed, Since: o.Since, Files: o.ChangedFiles})
}

func (o Options) withChanges() (Options, bool, error) {
	o.Selection = o.ChangeSelection()
	ctx := o.Context
	if ctx == nil {
		ctx = context.Background()
	}
	selected, err := o.Selection.Resolve(ctx, o.WorkDir)
	if err != nil || selected == nil {
		return o, false, err
	}
	logger.Infof("%s", selected.Summary())
	paths := &linters.LinterResult{WorkDir: o.WorkDir}
	for _, path := range selected.Paths(o.WorkDir) {
		paths.Violations = append(paths.Violations, models.Violation{File: filepath.Join(o.WorkDir, filepath.FromSlash(path))})
	}
	scopes := o.Files
	if len(scopes) == 0 {
		scopes = []string{o.WorkDir}
	}
	linters.FilterViolationsByUserScope([]*linters.LinterResult{paths}, o.WorkDir, scopes)
	o.Files = make([]string, 0, len(paths.Violations))
	for _, path := range paths.Violations {
		o.Files = append(o.Files, path.File)
	}
	if len(o.Files) == 0 {
		logger.Infof("Change selection: no applicable lint work")
		return o, true, nil
	}
	return o, false, nil
}

func scopeInvocation(inv linterInvocation, opts Options) (linterInvocation, []string, error) {
	scopes := resultScopes(&linters.LinterResult{Linter: inv.linter.Name(), WorkDir: inv.projectRoot}, opts.WorkDir, opts.Files)
	if opts.Selection.Active() {
		for _, file := range opts.Files {
			absolute := resolveLintPath(opts.WorkDir, file)
			if rooted, ok := inv.linter.(linters.ProjectRooted); ok && filepath.Dir(absolute) == inv.projectRoot && slices.Contains(rooted.ProjectRootMarkers(), filepath.Base(absolute)) {
				scopes = append(scopes, inv.projectRoot)
			}
			if (inv.linter.Name() == "golangci-lint" || inv.linter.Name() == "uir") && utils.IsWithin(inv.projectRoot, filepath.Dir(absolute)) {
				switch filepath.Base(absolute) {
				case "go.mod", "go.sum", "go.work", "go.work.sum":
					scopes = append(scopes, inv.projectRoot)
				}
			}
		}
	}
	if slices.Contains(scopes, inv.projectRoot) {
		inv.files = nil
		return inv, scopes, nil
	}
	if !opts.Selection.Active() {
		return inv, scopes, nil
	}
	files := make([]string, 0, len(inv.files))
	applicable := false
	for _, file := range inv.files {
		matches := len(inv.linter.DefaultIncludes()) == 0
		for _, pattern := range inv.linter.DefaultIncludes() {
			match, err := doublestar.Match(pattern, filepath.ToSlash(file))
			if err != nil {
				return inv, nil, err
			}
			matches = matches || match
		}
		if !matches {
			continue
		}
		applicable = true
		_, err := os.Lstat(resolveLintPath(inv.projectRoot, file))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return inv, nil, err
		}
		files = append(files, file)
	}
	inv.files = files
	inv.skip = !applicable
	if inv.linter.Name() == "golangci-lint" {
		inv.files = nil
	}
	return inv, scopes, nil
}
