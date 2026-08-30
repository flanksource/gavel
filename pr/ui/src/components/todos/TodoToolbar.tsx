import { Button, FilterBar } from '@flanksource/clicky-ui/components';
import type { FilterBarFilter } from '@flanksource/clicky-ui/components';
// No `icon` on the facets: clicky's comboboxLabelProps renders the icon
// INSTEAD of the text label, and three unlabelled glyphs read as broken.
import { UiClose, UiRefresh } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import { TODO_ACTIVITY_FILTER_LABEL, useTodoFilterBar } from './todoFilterBar';
import { TodoSelectionBar } from './TodoSelectionBar';
import { TodoGroupByMenu } from './TodoGroupByMenu';
import { TodoSortByMenu } from './TodoSortByMenu';
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
export function TodoToolbar({ todos }: { todos: WorkspaceTodos }) {
  const { groupBy, setGroupBy, sortBy, setSortBy, timeRange, setTimeRange, loadingList, refresh, selection } = todos;
  const { facets, range } = useTodoFilterBar(todos);

  // The activity range is a filter here rather than the bar's trailing range
  // slot: the trailing slot never collapses, and in a sidebar this narrow the
  // ~100px it holds is the difference between a facet being inline and being
  // hidden behind the overflow button. Last in the list, so it is also the
  // first thing to fold away.
  const bar: FilterBarFilter[] = [
    ...facets,
    { ...range, key: 'activity', kind: 'date-range', label: TODO_ACTIVITY_FILTER_LABEL },
  ];

  const RefreshIcon = loadingList ? Spinner : UiRefresh;

  return (
    <FilterBar
      className="shrink-0 flex-nowrap border-b border-border bg-card px-2"
      filters={bar}
      leading={
        <>
          <TodoGroupByMenu groupBy={groupBy} onChange={setGroupBy} />
          <TodoSortByMenu sortBy={sortBy} onChange={setSortBy} />
        </>
      }
      trailing={
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
      }
      // Selecting a todo takes the row rather than adding a second strip below
      // it, so the list underneath never shifts. The filters stay mounted
      // behind the bar and come back when the selection clears.
      {...(selection.selection.size > 0
        ? { overlay: <TodoSelectionBar todos={todos} onApplied={refresh} /> }
        : {})}
    />
  );
}
