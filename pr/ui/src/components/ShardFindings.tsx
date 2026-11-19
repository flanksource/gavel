import { useMemo, useState } from 'react';
import {
  DEFAULT_LINT_GROUPING,
  LintResults,
  TestRunner,
  annotateRoutePaths,
  emptyLintFilters,
  emptyTestFilters,
  findNodeByRoutePath,
  type LintFilters,
  type LintGrouping,
  type LintView,
  type TestFilters,
  type TestRunnerView,
} from '@flanksource/clicky-ui/data';
import { detailAdapter } from '@flanksource/gavel/testrunner';
import type { LinterResult, Test } from '../types';

// Artifact failures are gavel's own test nodes, so they get gavel's detail body
// (Go and Ginkgo context, attempt metrics, output) like the Tests tab.
const adapters = [detailAdapter()];

/** How many findings the PR view previews; the rest are one link away on the results page. */
export const PREVIEW_LIMIT = 5;

// The tree sits beside the selected finding's detail, so it needs a fixed-height
// pane; the list grows with its rows, each opening its detail beneath it.
const TREE_PANE_CLASS = 'h-96';

/** The first `limit` violations across the linters, dropping linters left without any. */
export function previewLint(lint: LinterResult[], limit: number): LinterResult[] {
  let left = limit;
  return lint.flatMap(result => {
    if (left === 0 || !result.violations?.length) return [];
    const violations = result.violations.slice(0, left);
    left -= violations.length;
    return [{ ...result, violations }];
  });
}

/** The artifact's first failing tests in clicky's test runner, as a list by default or a tree beside the detail. */
export function ShardTestFailures({ failures }: { failures: Test[] }) {
  const tests = useMemo(() => annotateRoutePaths(failures.slice(0, PREVIEW_LIMIT)), [failures]);
  const [view, setView] = useState<TestRunnerView>('list');
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const selected = useMemo(() => findNodeByRoutePath(tests, selectedPath ?? ''), [tests, selectedPath]);
  const changeView = (next: TestRunnerView) => {
    setView(next);
    // The tree's detail pane would otherwise open empty: start it on the first failure.
    if (next === 'tree' && !selectedPath) setSelectedPath(tests[0]?.route_path ?? null);
  };
  return (
    <div className={view === 'tree' ? TREE_PANE_CLASS : undefined}>
      <TestRunner
        adapters={adapters}
        tests={tests}
        selected={selected}
        filters={filters}
        expandAll={expandAll}
        done
        hideHeader
        title={null}
        view={view}
        onViewChange={changeView}
        onSelect={node => setSelectedPath(node?.route_path ?? null)}
        onFiltersChange={setFilters}
        onExpandAllChange={setExpandAll}
      />
    </div>
  );
}

/** The artifact's first lint findings in clicky's lint results, listed like `gavel lint` prints them by default. */
export function ShardLintFindings({ lint }: { lint: LinterResult[] }) {
  const preview = useMemo(() => previewLint(lint, PREVIEW_LIMIT), [lint]);
  const [view, setView] = useState<LintView>('list');
  const [grouping, setGrouping] = useState<LintGrouping>(DEFAULT_LINT_GROUPING);
  const [filters, setFilters] = useState<LintFilters>(emptyLintFilters());
  const [selected, setSelected] = useState<string | null>(null);
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  return (
    <div className={view === 'tree' ? TREE_PANE_CLASS : undefined}>
      <LintResults
        results={preview}
        title={null}
        view={view}
        onViewChange={setView}
        grouping={grouping}
        onGroupingChange={setGrouping}
        filters={filters}
        onFiltersChange={setFilters}
        selected={selected}
        onSelect={setSelected}
        expandAll={expandAll}
        onExpandAllChange={setExpandAll}
      />
    </div>
  );
}
