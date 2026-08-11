import { useMemo } from 'react';
import { Button, FilterBar } from '@flanksource/clicky-ui/components';
import type { FilterBarFilter, MultiSelectOption } from '@flanksource/clicky-ui/components';
// No `icon` on the facets below: clicky's comboboxLabelProps renders the icon
// INSTEAD of the text label, and three unlabelled glyphs read as broken.
import { UiClose, UiRefresh } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import type { FacetModes } from '../../utils';
import { flattenTodos } from './todoGroup';
import { EXTERNAL_FILTER_DEFS, PRIORITY_FILTER_DEFS, STATUS_FILTER_DEFS, externalKey, priorityKey } from './todoFilter';
import { TodoGroupByMenu } from './TodoGroupByMenu';
import { TodoSortByMenu } from './TodoSortByMenu';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// TodoToolbar is the single filter row for the todos list, on both the desktop
// sidebar and the compact menubar/mobile view. It is clicky-ui's FilterBar, so
// the todos facets behave exactly like the PRs tab's: tri-state include/exclude
// chips with counts in the option labels.
//
// The row never wraps. `flex-nowrap` in className overrides clicky's default
// `flex-wrap md:flex-nowrap`, and `overflowMode="responsive"` (the default) is
// what makes that safe: the bar measures itself and moves whatever no longer
// fits into its "More filters" popover, folding every facet in below 768px. So
// the narrow bar is Group · Sort · ⧩ · range · refresh on one line, and the same
// controls stay reachable instead of a second, weaker mobile bar.
export function TodoToolbar({ todos }: { todos: WorkspaceTodos }) {
  const {
    workspaces, byDir, aggregate, filters, setFilters,
    groupBy, setGroupBy, sortBy, setSortBy, timeRange, setTimeRange, loadingList, refresh,
  } = todos;

  // Priority and external-link counts are derived from the loaded items — the
  // server's TodoCounts only aggregates status, and the items are already here.
  const derived = useMemo(() => {
    const priorities: Record<string, number> = {};
    const external: Record<string, number> = {};
    for (const { todo } of flattenTodos(workspaces, byDir)) {
      priorities[priorityKey(todo)] = (priorities[priorityKey(todo)] ?? 0) + 1;
      external[externalKey(todo)] = (external[externalKey(todo)] ?? 0) + 1;
    }
    return { priorities, external };
  }, [workspaces, byDir]);

  const setFacet = (key: 'statuses' | 'priorities' | 'external', value: FacetModes) =>
    setFilters({ ...filters, [key]: value });

  const statusOptions: MultiSelectOption[] = STATUS_FILTER_DEFS
    .filter(def => aggregate[def.countKey] > 0)
    .map(def => ({ value: def.status, label: `${def.label} (${aggregate[def.countKey]})` }));
  const priorityOptions: MultiSelectOption[] = PRIORITY_FILTER_DEFS
    .filter(def => (derived.priorities[def.priority] ?? 0) > 0)
    .map(def => ({ value: def.priority, label: `${def.label} (${derived.priorities[def.priority]})` }));
  const externalOptions: MultiSelectOption[] = EXTERNAL_FILTER_DEFS
    .map(def => ({ value: def.key, label: `${def.label} (${derived.external[def.key] ?? 0})` }));

  const bar: FilterBarFilter[] = [];
  if (statusOptions.length) {
    bar.push({
      key: 'status', kind: 'multi', label: 'Status', options: statusOptions,
      value: filters.statuses, onChange: value => setFacet('statuses', value),
    });
  }
  if (priorityOptions.length) {
    bar.push({
      key: 'priority', kind: 'multi', label: 'Priority', options: priorityOptions,
      value: filters.priorities, onChange: value => setFacet('priorities', value),
    });
  }
  // A workspace that has never pushed a todo to GitHub gets no external facet —
  // "Not linked (all of them)" is chrome with nothing behind it. It appears as
  // soon as one todo is linked, or while a stale exclusion is still applied.
  if ((derived.external.linked ?? 0) > 0 || Object.keys(filters.external).length > 0) {
    bar.push({
      key: 'external', kind: 'multi', label: 'Issue',
      options: externalOptions, value: filters.external, onChange: value => setFacet('external', value),
    });
  }

  // The activity range is a filter here rather than the bar's trailing range
  // slot: the trailing slot never collapses, and in a sidebar this narrow the
  // ~100px it holds is the difference between a facet being inline and being
  // hidden behind the overflow button. Last in the list, so it is also the
  // first thing to fold away.
  bar.push({
    key: 'activity', kind: 'date-range', label: 'Active',
    from: timeRange?.from,
    to: timeRange?.to,
    // An applied range with both bounds empty is the user clearing it.
    onApply: (from, to) => setTimeRange(from || to ? { from, to } : null),
    presets: ['hr', 'day', 'wk+'],
    emptyLabel: 'Any time',
  });

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
    />
  );
}
