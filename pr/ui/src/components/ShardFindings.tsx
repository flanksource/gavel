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
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { detailAdapter } from '@flanksource/gavel/testrunner';
import type { LinterResult, Test } from '../types';

// Artifact failures are gavel's own test nodes, so they get gavel's detail body
// (Go and Ginkgo context, attempt metrics, output) like the Tests tab.
const adapters = [detailAdapter()];

// Both lists are a bounded sample of the artifact, shown in a fixed-height pane:
// the tree on the left, the selected finding's detail on the right.
const PANE_CLASS = 'h-96';

/** The artifact's failing tests in clicky's test runner: tree beside the message, output and failure detail. */
export function ShardTestFailures({ failures }: { failures: Test[] }) {
  const tests = useMemo(() => annotateRoutePaths(failures), [failures]);
  // Start on the first failure so triage sees the reason without a click.
  const [selectedPath, setSelectedPath] = useState<string | null>(tests[0]?.route_path ?? null);
  const [filters, setFilters] = useState<TestFilters>(emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const selected = useMemo(() => findNodeByRoutePath(tests, selectedPath ?? ''), [tests, selectedPath]);
  return (
    <div className={PANE_CLASS}>
      <TestRunner
        adapters={adapters}
        tests={tests}
        selected={selected}
        filters={filters}
        expandAll={expandAll}
        done
        hideHeader
        title={null}
        onSelect={node => setSelectedPath(node?.route_path ?? null)}
        onFiltersChange={setFilters}
        onExpandAllChange={setExpandAll}
      />
    </div>
  );
}

/** The artifact's lint findings in clicky's lint results: linter, rule and file groups beside the finding. */
export function ShardLintFindings({ lint }: { lint: LinterResult[] }) {
  const [grouping, setGrouping] = useState<LintGrouping>(DEFAULT_LINT_GROUPING);
  const [filters, setFilters] = useState<LintFilters>(emptyLintFilters());
  const [selected, setSelected] = useState<string | null>(null);
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  return (
    <div className={PANE_CLASS}>
      <LintResults
        results={lint}
        title={null}
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
