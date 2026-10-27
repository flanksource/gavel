package lint

import (
	"context"
	"fmt"
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

func (o Options) WithChanges() (Options, bool, error) {
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
	if opts.Selection.Active() {
		if trigger := projectWideChange(inv, opts); trigger != "" {
			logger.Infof("%s: %s changed; analysing %s, reporting changed files only", inv.linter.Name(), trigger, inv.projectRoot)
			if inv.configDir == "" {
				inv.files = nil
			}
			return inv, slices.Clone(opts.Files), nil
		}
	}
	scopes := resultScopes(&linters.LinterResult{Linter: inv.linter.Name(), WorkDir: inv.projectRoot}, opts.WorkDir, opts.Files)
	if slices.Contains(scopes, inv.projectRoot) {
		if inv.configDir == "" {
			inv.files = nil
		}
		return inv, scopes, nil
	}
	if !opts.Selection.Active() {
		applicable, err := anyRequestedFileApplies(inv)
		inv.skip = !applicable
		return inv, scopes, err
	}
	files := make([]string, 0, len(inv.files))
	applicable := false
	for _, file := range inv.files {
		matches, err := matchesLinterIncludes(inv.linter, file)
		if err != nil {
			return inv, nil, err
		}
		if !matches {
			continue
		}
		applicable = true
		_, err = os.Lstat(resolveLintPath(inv.projectRoot, file))
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
	if inv.linter.Name() == "golangci-lint" && inv.configDir == "" {
		inv.files = nil
	}
	return inv, scopes, nil
}

// projectWideChange returns the changed file (config, root marker or Go
// module file) that requires the linter to analyse its whole project. The
// reported violations stay limited to the changed files.
func projectWideChange(inv linterInvocation, opts Options) string {
	probe := &linters.LinterResult{Linter: inv.linter.Name(), WorkDir: inv.projectRoot}
	for _, file := range opts.Files {
		absolute := resolveLintPath(opts.WorkDir, file)
		if slices.Contains(resultScopes(probe, opts.WorkDir, []string{file}), inv.projectRoot) {
			return file
		}
		if rooted, ok := inv.linter.(linters.ProjectRooted); ok && filepath.Dir(absolute) == inv.projectRoot && slices.Contains(rooted.ProjectRootMarkers(), filepath.Base(absolute)) {
			return file
		}
		if (inv.linter.Name() == "golangci-lint" || inv.linter.Name() == "uir") && utils.IsWithin(inv.projectRoot, filepath.Dir(absolute)) {
			switch filepath.Base(absolute) {
			case "go.mod", "go.sum", "go.work", "go.work.sum":
				return file
			}
		}
	}
	return ""
}

// anyRequestedFileApplies reports whether an explicit file list gives the
// linter something to lint: a directory, or a file matching its includes.
func anyRequestedFileApplies(inv linterInvocation) (bool, error) {
	if len(inv.files) == 0 {
		return true, nil
	}
	for _, file := range inv.files {
		absolute := resolveLintPath(inv.projectRoot, file)
		if info, err := os.Stat(absolute); err == nil && info.IsDir() {
			return true, nil
		}
		if rel, err := filepath.Rel(inv.projectRoot, absolute); err == nil {
			file = rel
		}
		matches, err := matchesLinterIncludes(inv.linter, file)
		if err != nil || matches {
			return matches, err
		}
	}
	return false, nil
}

func matchesLinterIncludes(linter linters.Linter, file string) (bool, error) {
	includes := linter.DefaultIncludes()
	if len(includes) == 0 {
		return true, nil
	}
	for _, pattern := range includes {
		match, err := doublestar.Match(pattern, filepath.ToSlash(file))
		if err != nil {
			return false, fmt.Errorf("match %s include %q against %s: %w", linter.Name(), pattern, file, err)
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}
