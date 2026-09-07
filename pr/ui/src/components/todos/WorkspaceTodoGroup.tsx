import { useCallback, useMemo, useState } from 'react';
import { Button, ListMenuHeader, ListMenuSection } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiChevronRight, UiFolder } from '@flanksource/clicky-ui/icons';
import type { Project, TodoDensity, TodoListResponse, TodoStatus } from '../../types';
import { RepoIcon } from '../RepoIcon';
import { emptyCounts, TodoCountsBar, TodoRow, type TodoRowTarget } from './format';
import type { TagIndex } from './tagResolve';
import { defaultTodoFilters, isTodoVisible, type TodoFilters } from './todoFilter';
import { GroupSelectAll } from './TodoGroupSelectAll';
import type { TodoSelection } from './todoSelection';
import type { TodoSort } from './todoSort';
import { defaultTodoSort, todoComparator } from './todoSort';
import type { ResolvedRange } from './todoTimeRange';

// The shared "this workspace has no todos yet" list. A fresh `[]` per render
// would change `allItems` identity and defeat the memo over filter+sort below.
const emptyItems: TodoListResponse['items'] = [];

// WorkspaceTodoGroup is one collapsible workspace section, mirroring the PR
// tab's per-repo grouping: a sticky header with the workspace name and its
// open/failed/total counts, with the workspace's todos listed beneath. The
// Closed/Status filter hides matching rows but leaves the header counts whole.
export function WorkspaceTodoGroup({ workspace, data, selectedRef, onSelect, filters, onToggleStatus, range, density = 'comfortable', sortBy = defaultTodoSort(), selection, tags }: {
  workspace: Project;
  data?: TodoListResponse;
  selectedRef: string;
  // Takes the row's identity rather than a bare ref so callers can pass their
  // stable `select` straight down. A caller closing over the workspace dir
  // (`ref => select({ dir, ref })`) would build a new function per workspace per
  // render and re-render every row underneath it.
  onSelect: (target: TodoRowTarget) => void;
  filters?: TodoFilters;
  onToggleStatus?: (status: TodoStatus) => void;
  range?: ResolvedRange | null;
  density?: TodoDensity;
  sortBy?: TodoSort;
  selection?: TodoSelection;
  tags?: TagIndex;
}) {
  const [open, setOpen] = useState(true);

  const active = filters ?? defaultTodoFilters();
  const allItems = data?.items ?? emptyItems;
  // Filtering and sorting the workspace's todos is O(n log n) over every todo it
  // owns. Unmemoised it ran on every render of this group — and this group
  // re-renders whenever anything above it does — so an idle dashboard re-sorted
  // every workspace several times a second. Sorting also allocates a new array,
  // which changed `items` identity and cascaded into every row below.
  const items = useMemo(
    () => allItems.filter(item => isTodoVisible(item, active, range)).sort(todoComparator(sortBy)),
    [allItems, active, range, sortBy],
  );
  const hiddenCount = allItems.length - items.length;
  const counts = data?.counts ?? workspace.todoCounts ?? emptyCounts;

  // Match the PR tab's per-repo header: when the workspace maps to a GitHub repo,
  // show that repo's icon and short name in place of a generic folder + dir name.
  const repo = workspace.repos?.[0];
  const repoShort = repo ? repo.split('/').pop() || repo : '';

  // In bulk-edit mode the header's checkbox checks or clears every row the
  // filters currently show — never the hidden ones, so what is checked is always
  // what is on screen.
  const bulkTargets = useMemo(
    () => items.map(item => ({ dir: workspace.dir, ref: item.ref })),
    [items, workspace.dir],
  );

  // Bound once per group rather than once per row: TodoRow calls back with the
  // row's own identity, so these two keep one reference across renders and the
  // rows' memo holds.
  const toggleSelected = selection?.toggleSelected;
  const handleToggleSelect = useCallback(
    (target: TodoRowTarget) => toggleSelected?.(target),
    [toggleSelected],
  );

  return (
    <ListMenuSection>
      <ListMenuHeader>
        {selection && (
          <GroupSelectAll label={workspace.name} targets={bulkTargets} selection={selection} />
        )}
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex h-auto min-w-0 flex-1 items-center justify-start gap-2 p-0 text-left hover:opacity-80"
        >
          {open ? <UiChevronDown className="text-muted-foreground text-xs" /> : <UiChevronRight className="text-muted-foreground text-xs" />}
          {repo ? (
            <>
              <RepoIcon repo={repo} size={16} />
              <span className="min-w-0 flex-1 truncate text-sm font-medium text-foreground" title={workspace.dir}>{repoShort}</span>
            </>
          ) : (
            <>
              <UiFolder className="text-muted-foreground text-xs" />
              <span className="min-w-0 flex-1 truncate text-sm font-semibold text-foreground" title={workspace.dir}>{workspace.name}</span>
            </>
          )}
        </Button>
        <TodoCountsBar counts={counts} statusFilter={active.statuses} onToggle={onToggleStatus} />
      </ListMenuHeader>
      {/* Rows and the empty-state note are siblings rather than two branches of
          a ternary. As a ternary, a workspace going from "no rows yet" to "rows
          loaded" swapped one subtree for another, so React unmounted and
          remounted the whole group instead of reconciling it — the source of
          both the mount churn during load and the layout shift as late data
          landed. */}
      {open && items.map(item => (
        <TodoRow
          key={item.ref}
          todo={item}
          active={item.ref === selectedRef}
          onSelect={onSelect}
          density={density}
          dir={workspace.dir}
          selectable={Boolean(selection)}
          selected={selection?.isSelected({ dir: workspace.dir, ref: item.ref })}
          onToggleSelect={handleToggleSelect}
          tags={tags}
        />
      ))}
      {open && items.length === 0 && (
        <div className="px-3 py-2 text-xs text-muted-foreground">
          {hiddenCount > 0 ? `${hiddenCount} todo${hiddenCount === 1 ? '' : 's'} hidden by filter` : 'No todos'}
        </div>
      )}
    </ListMenuSection>
  );
}
