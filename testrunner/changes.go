package testrunner

import (
	"context"
	"fmt"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func (opts RunOptions) ChangeOptions() changeset.Options {
	return changeset.Options{Dir: opts.WorkDir, Changed: opts.Changed, Since: opts.Since, Files: opts.ChangedFiles}
}

func (opts RunOptions) ChangeSelection() *changeset.Selection {
	if opts.Selection != nil {
		return opts.Selection
	}
	return changeset.NewSelection(opts.ChangeOptions())
}

func (o *TestOrchestrator) selectChanges(packages map[Framework][]string) error {
	o.Selection = o.ChangeSelection()
	ctx := o.Context
	if ctx == nil {
		ctx = context.Background()
	}
	selected, err := o.Selection.Resolve(ctx, o.WorkDir)
	if err != nil {
		return err
	}
	if selected != nil {
		logger.Infof("%s", selected.Summary())
		if len(selected.Changes) == 0 {
			for framework := range packages {
				packages[framework] = nil
			}
			logger.Infof("Change selection: no applicable test work")
			return nil
		}
	}
	if (selected != nil || o.Cache) && (len(packages[parsers.GoTest]) > 0 || len(packages[parsers.Ginkgo]) > 0) {
		o.selector, err = newSelectorContext(o.WorkDir, o.Tags)
		if err != nil {
			return fmt.Errorf("initialize selector: %w", err)
		}
	}
	if selected != nil && o.selector != nil {
		for _, framework := range []Framework{parsers.GoTest, parsers.Ginkgo} {
			before := len(packages[framework])
			packages[framework], err = o.selector.filterByChangeGraph(packages[framework], selected.Paths(o.WorkDir))
			if err != nil {
				return err
			}
			logger.Infof("[%s] change-graph: %d → %d packages", framework, before, len(packages[framework]))
		}
	}
	return nil
}
