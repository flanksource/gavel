import { useMemo, useState } from 'react';
import {
  LintList,
  TestRunner,
  annotateRoutePaths,
  emptyLintFilters,
  emptyTestFilters,
  findNodeByRoutePath,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { detailAdapter } from '@flanksource/gavel/testrunner';
import type { LinterResult, Test } from '../types';

// Artifact failures are gavel's own test nodes, so they get gavel's detail body
// (Go and Ginkgo context, attempt metrics, output) like the Tests tab.
const adapters = [detailAdapter()];

/** How many findings the PR view previews; the rest are one link away on the results page. */
export const PREVIEW_LIMIT = 5;

const NO_LINT_FILTERS = emptyLintFilters();

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

/**
 * The artifact's first failing tests as clicky's test list, each opening its
 * detail beneath it. A preview this short gets no view toolbar: the tree and
 * filters live on the results page its "Show all" link opens.
 */
export function ShardTestFailures({ failures }: { failures: Test[] }) {
  const tests = useMemo(() => annotateRoutePaths(failures.slice(0, PREVIEW_LIMIT)), [failures]);
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const selected = useMemo(() => findNodeByRoutePath(tests, selectedPath ?? ''), [tests, selectedPath]);
  return (
    <TestRunner
      adapters={adapters}
      tests={tests}
      selected={selected}
      filters={filters}
      expandAll={expandAll}
      done
      hideHeader
      title={null}
      view="list"
      onSelect={node => setSelectedPath(node?.route_path ?? null)}
      onFiltersChange={setFilters}
      onExpandAllChange={setExpandAll}
    />
  );
}

/** The artifact's first lint findings listed like `gavel lint` prints them, without the results view's stats and toolbar. */
export function ShardLintFindings({ lint }: { lint: LinterResult[] }) {
  const preview = useMemo(() => previewLint(lint, PREVIEW_LIMIT), [lint]);
  return <LintList results={preview} filters={NO_LINT_FILTERS} className="p-2" />;
}
