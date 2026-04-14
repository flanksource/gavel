import { useMemo } from 'preact/hooks';
import type { Test, LinterResult } from '../types';
import { groupLintByLinterFile, groupLintByFileLinterRule } from '../utils';
import { TestNode } from './TestNode';
import type { LintGrouping, LintFilters } from './LintFilterBar';

interface Props {
  lint: LinterResult[] | undefined;
  grouping: LintGrouping;
  filters: LintFilters;
  expandAll: boolean | null;
  selected: Test | null;
  onSelect: (t: Test) => void;
}

export function LintView({ lint, grouping, filters, expandAll, selected, onSelect }: Props) {
  const tree = useMemo(() => {
    if (grouping === 'linter-file') return groupLintByLinterFile(lint, filters);
    return groupLintByFileLinterRule(lint, filters);
  }, [lint, grouping, filters]);

  if (tree.length === 0) {
    return (
      <div class="p-8 text-center text-gray-400 text-sm">
        {(lint || []).length === 0 ? 'No lint results' : 'No violations match the current filters'}
      </div>
    );
  }

  return (
    <>
      {tree.map((t, i) => (
        <TestNode
          key={i}
          test={t}
          depth={0}
          expandAll={expandAll}
          selected={selected}
          onSelect={onSelect}
        />
      ))}
    </>
  );
}
