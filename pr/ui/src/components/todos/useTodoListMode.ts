import { useCallback, useState } from 'react';
import type { TodoStatus } from '../../types';
import { useIsMobile } from '../../useIsMobile';
import { focusStatusFilter } from './todoFilter';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// useTodoListMode is how a todo list behaves on a phone, shared by the dashboard
// sidebar and the menubar. A phone has no room for a checkbox on every row, so
// selection is a mode the toolbar's overflow menu enters; anything already
// checked keeps it on. The desktop keeps its unconditional checkboxes. The
// group header counts shrink to the statuses that need action, each tap
// applying that status as the filter.
export function useTodoListMode({ filters, setFilters, selection }: Pick<WorkspaceTodos, 'filters' | 'setFilters' | 'selection'>) {
  const mobile = useIsMobile();
  const [selecting, setSelecting] = useState(false);
  const selectMode = !mobile || selecting || selection.selection.size > 0;
  const focusStatus = useCallback(
    (status: TodoStatus) => setFilters({ ...filters, statuses: focusStatusFilter(filters.statuses, status) }),
    [filters, setFilters],
  );
  return {
    selecting,
    setSelecting,
    rowSelection: selectMode ? selection : undefined,
    onFocusStatus: mobile ? focusStatus : undefined,
  };
}
