import { Button, DropdownMenu, FilterBar } from '@flanksource/clicky-ui/components';
import type { DropdownMenuItem, FilterBarFilter } from '@flanksource/clicky-ui/components';
// No `icon` on the facets: clicky's comboboxLabelProps renders the icon
// INSTEAD of the text label, and three unlabelled glyphs read as broken.
import { UiCheck, UiClose, UiDotsVertical, UiRefresh } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import type { TodoGroupBy } from '../../types';
import { useIsMobile } from '../../useIsMobile';
import { TODO_ACTIVITY_FILTER_LABEL, useTodoFilterBar } from './todoFilterBar';
import { TodoSelectionBar } from './TodoSelectionBar';
import { TodoGroupByMenu } from './TodoGroupByMenu';
import { TodoSortByMenu } from './TodoSortByMenu';
import { GROUP_BY_OPTIONS, defaultGroupBy } from './todoGroup';
import { TODO_SORT_COLUMN_OPTIONS, defaultTodoSort, type TodoSort, type TodoSortColumn } from './todoSort';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// TodoToolbar is the filter row for the todos list in the split (master-detail)
// layout and the compact menubar/mobile view. It is clicky-ui's FilterBar over
// the shared useTodoFilterBar facets, so the todos facets behave exactly like
// the PRs tab's: tri-state include/exclude chips with counts in the option
// labels. The full-width layout renders the same facets inside DataTable's own
// filter bar instead — see TodoTable.
//
// The row never wraps. `flex-nowrap` in className overrides clicky's default
// `flex-wrap md:flex-nowrap`, and `overflowMode="responsive"` (the default) is
// what makes that safe: the bar measures itself and moves whatever no longer
// fits into its "More filters" popover, folding every facet in below 768px. So
// the narrow bar is Group · Sort · ⧩ · range · refresh on one line, and the same
// controls stay reachable instead of a second, weaker mobile bar.
//
// On a phone the row is Filters · ⋮: Group and Sort join the Filters sheet as
// single choices (their bare glyphs read as two unlabelled dropdowns), and
// refresh and select mode move to the overflow menu.
export function TodoToolbar({ todos, selecting = false, onSelectingChange }: {
  todos: WorkspaceTodos;
  // Phone select mode, owned by the list whose checkboxes it reveals. A host
  // without it (the menubar) offers no Select action.
  selecting?: boolean;
  onSelectingChange?: (selecting: boolean) => void;
}) {
  const { groupBy, setGroupBy, sortBy, setSortBy, timeRange, setTimeRange, loadingList, refresh, selection } = todos;
  const { facets, range } = useTodoFilterBar(todos);
  const mobile = useIsMobile();

  // The activity range is a filter here rather than the bar's trailing range
  // slot: the trailing slot never collapses, and in a sidebar this narrow the
  // ~100px it holds is the difference between a facet being inline and being
  // hidden behind the overflow button. Last in the list, so it is also the
  // first thing to fold away.
  const bar: FilterBarFilter[] = [
    ...facets,
    { ...range, key: 'activity', kind: 'date-range', label: TODO_ACTIVITY_FILTER_LABEL },
    ...(mobile ? [groupByFilter(groupBy, setGroupBy), sortByFilter(sortBy, setSortBy)] : []),
  ];

  const RefreshIcon = loadingList ? Spinner : UiRefresh;

  const overlay = selection.selection.size > 0
    ? <TodoSelectionBar todos={todos} onApplied={refresh} />
    : mobile && selecting && onSelectingChange
      ? <SelectModeHint onDone={() => onSelectingChange(false)} />
      : undefined;

  const menu: DropdownMenuItem[] = [
    ...(onSelectingChange ? [{ label: 'Select todos', icon: UiCheck, onSelect: () => onSelectingChange(true) }] : []),
    { label: loadingList ? 'Refreshing todos' : 'Refresh todos', icon: UiRefresh, disabled: loadingList, onSelect: refresh },
  ];

  return (
    <FilterBar
      className="shrink-0 flex-nowrap border-b border-border bg-card px-2"
      filters={bar}
      leading={mobile ? undefined : (
        <>
          <TodoGroupByMenu groupBy={groupBy} onChange={setGroupBy} />
          <TodoSortByMenu sortBy={sortBy} onChange={setSortBy} />
        </>
      )}
      trailing={mobile ? (
        <DropdownMenu
          align="right"
          menuLabel="Todo actions"
          trigger={(
            <Button type="button" variant="ghost" size="icon" aria-label="More todo actions" title="More todo actions" className="size-9 shrink-0 text-muted-foreground">
              <UiDotsVertical />
            </Button>
          )}
          items={menu}
        />
      ) : (
        <>
          {timeRange && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setTimeRange(null)}
              title="Clear time filter"
              aria-label="Clear time filter"
              className="h-8 w-8 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <UiClose className="text-xs" />
            </Button>
          )}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={refresh}
            disabled={loadingList}
            title="Refresh todos"
            aria-label="Refresh todos"
            className="h-8 w-8 shrink-0 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
          >
            <RefreshIcon className="text-sm" />
          </Button>
        </>
      )}
      // Selecting a todo takes the row rather than adding a second strip below
      // it, so the list underneath never shifts. The filters stay mounted
      // behind the bar and come back when the selection clears.
      {...(overlay ? { overlay } : {})}
    />
  );
}

// Select mode before anything is checked: the row says what to do next and how
// to leave, rather than showing filters that no longer describe the task.
function SelectModeHint({ onDone }: { onDone: () => void }) {
  return (
    <div className="flex min-w-0 flex-1 items-center justify-between gap-2 px-2 text-xs text-muted-foreground">
      <span>Tap todos to select</span>
      <Button type="button" variant="ghost" size="sm" onClick={onDone}>Done</Button>
    </div>
  );
}

// Grouping and sorting always hold a value, so each is offered with its default
// as the empty choice: a non-default arrangement then counts on the Filters
// badge, and clearing it goes back to the default rather than to nothing.
function groupByFilter(groupBy: TodoGroupBy, setGroupBy: (groupBy: TodoGroupBy) => void): FilterBarFilter {
  const fallback = defaultGroupBy();
  return {
    key: 'groupBy',
    kind: 'enum',
    label: 'Group by',
    value: groupBy === fallback ? '' : groupBy,
    placeholder: groupOption(fallback).label,
    options: GROUP_BY_OPTIONS.map(opt => ({ value: opt.value, label: opt.label })),
    onChange: value => setGroupBy(value === '' ? fallback : groupOption(value).value),
  };
}

function groupOption(value: string) {
  const option = GROUP_BY_OPTIONS.find(opt => opt.value === value);
  if (!option) throw new Error(`unknown todo grouping "${value}"`);
  return option;
}

const SORT_DIRECTION_WORDS: Record<TodoSortColumn, Record<TodoSort['dir'], string>> = {
  priority: { desc: 'high first', asc: 'low first' },
  updated: { desc: 'newest first', asc: 'oldest first' },
  created: { desc: 'newest first', asc: 'oldest first' },
  title: { asc: 'A to Z', desc: 'Z to A' },
};

const SORT_OPTIONS = TODO_SORT_COLUMN_OPTIONS.flatMap(opt => (['desc', 'asc'] as const).map(dir => ({
  value: `${opt.value}:${dir}`,
  label: `${opt.label}, ${SORT_DIRECTION_WORDS[opt.value][dir]}`,
  sort: { column: opt.value, dir } satisfies TodoSort,
})));

function sortByFilter(sortBy: TodoSort, setSortBy: (sortBy: TodoSort) => void): FilterBarFilter {
  const fallback = defaultTodoSort();
  const isDefault = sortBy.column === fallback.column && sortBy.dir === fallback.dir;
  return {
    key: 'sortBy',
    kind: 'enum',
    label: 'Sort by',
    value: isDefault ? '' : `${sortBy.column}:${sortBy.dir}`,
    placeholder: sortOption(`${fallback.column}:${fallback.dir}`).label,
    options: SORT_OPTIONS.map(({ value, label }) => ({ value, label })),
    onChange: value => setSortBy(value === '' ? fallback : sortOption(value).sort),
  };
}

function sortOption(value: string) {
  const option = SORT_OPTIONS.find(opt => opt.value === value);
  if (!option) throw new Error(`unknown todo sort "${value}"`);
  return option;
}
