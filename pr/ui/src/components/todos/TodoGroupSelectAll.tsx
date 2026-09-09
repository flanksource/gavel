import { useEffect, useRef } from 'react';
import type { SelectedTodo, TodoSelection } from './todoSelection';

// GroupSelectAll is the tri-state checkbox in a group header: unchecked when the
// group holds nothing selected, indeterminate when it holds some, checked when it
// holds all. `targets` is the group's *visible* rows, so checking it never reaches
// a row the filters are hiding.
//
// `indeterminate` is a DOM property with no HTML attribute, so it is set through a
// ref rather than JSX.
export function GroupSelectAll({ label, targets, selection }: {
  label: string;
  targets: SelectedTodo[];
  selection: TodoSelection;
}) {
  const ref = useRef<HTMLInputElement>(null);
  const selectedCount = targets.filter(target => selection.isSelected(target)).length;
  const all = targets.length > 0 && selectedCount === targets.length;
  const some = selectedCount > 0 && !all;

  useEffect(() => {
    if (ref.current) ref.current.indeterminate = some;
  }, [some]);

  return (
    <label className="flex shrink-0 cursor-pointer items-center pr-1" title={`Select every visible todo in ${label}`}>
      <input
        ref={ref}
        type="checkbox"
        checked={all}
        disabled={targets.length === 0}
        onChange={() => selection.setGroupSelected(targets, !all)}
        aria-label={`Select every visible todo in ${label}`}
        className="h-3.5 w-3.5 cursor-pointer accent-primary disabled:cursor-not-allowed disabled:opacity-40"
      />
    </label>
  );
}
