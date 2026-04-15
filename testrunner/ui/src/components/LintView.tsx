import type { Test, LinterResult } from '../types';
import { TestNode } from './TestNode';

interface Props {
  lint: LinterResult[] | undefined;
  tree: Test[];
  expandAll: boolean | null;
  selected: Test | null;
  onSelect: (t: Test) => void;
}

export function LintView({ lint, tree, expandAll, selected, onSelect }: Props) {
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
