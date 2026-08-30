import { useMemo } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { DataTable } from '@flanksource/clicky-ui/data';
import type { DataTableColumn, DataTableGroupingMode } from '@flanksource/clicky-ui/data';
import { UiClose, UiFolder, UiRefresh } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import type { Project, TodoGroupBy, TodoItem, TodoPhase } from '../../types';
import { TODO_PHASES } from '../../types';
import { TodoPhaseCell } from './TodoPhaseCell';
import { phase as phaseMeta } from './phaseMachine';
import type { TagIndex } from './tagResolve';
import { todoVisibleLabels } from './tagResolve';
import {
  SessionBadge,
  StatusIcon,
  TodoAges,
  TodoCountsBar,
  TodoDiffBadge,
  TodoPlanIndicator,
  TodoVerificationIndicator,
  countsFromItems,
  isLiveRun,
  priorityBadgeClass,
  priorityIcon,
  statusLabel,
} from './format';
import { isEntryVisible } from './todoFilter';
import { TODO_ACTIVITY_FILTER_LABEL, useTodoFilterBar } from './todoFilterBar';
import {
  AGE_BUCKETS,
  SEVERITY_BUCKETS,
  ageKey,
  flattenTodos,
  severityKey,
  type TodoEntry,
} from './todoGroup';
import { defaultTodoSort, todoComparator, type TodoSortColumn } from './todoSort';
import { resolveRange } from './todoTimeRange';
import { selectionKey } from './todoSelection';
import { useTodoBulkContext, useTodoBulkToolbar } from './todoActions';
import { TodoTagRow } from './TodoTag';
import type { WorkspaceTodos } from './useWorkspaceTodos';

// A DataTable row is the same {todo, workspace} pair the list groups on. The
// index signature is DataTable's `T extends Record<string, unknown>` constraint;
// every column reads through an `accessor`, so no key is ever resolved as a path.
export type TodoTableRow = TodoEntry & Record<string, unknown>;

// The row id doubles as the bulk-selection key, so a todo checked in the table
// is the same todo the bulk bar acts on.
export const todoTableRowId = (row: TodoTableRow): string =>
  selectionKey({ dir: row.workspace.dir, ref: row.todo.ref });

// The phase's own label, so a column header reads the same word the detail
// pane's phase strip and the session viewer use.
const phaseLabel = (phase: TodoPhase): string => phaseMeta(phase).label;

const AGE_ORDER = [...AGE_BUCKETS.map(b => b.key), 'none'];
const AGE_LABELS: Record<string, string> = {
  ...Object.fromEntries(AGE_BUCKETS.map(b => [b.key, b.label])),
  none: 'No activity',
};

// byDefinedOrder ranks a group key against a fixed list, so severity reads
// high→medium→low and age reads today→older however the rows happen to be
// sorted. DataTable keeps first-appearance order otherwise, which follows the
// sort column and would shuffle the group headers with it.
function byDefinedOrder(order: string[]) {
  return (a: { key: string }, b: { key: string }) => {
    const rank = (key: string) => {
      const at = order.indexOf(key);
      return at === -1 ? order.length : at;
    };
    return rank(a.key) - rank(b.key);
  };
}

// todoGroupingModes are the strategies behind DataTable's native grouping
// picker, driving the same groupBy preference the split layout's Group menu
// does. They bucket with todoGroup's own severityKey/ageKey rather than a copy.
export function todoGroupingModes(
  { workspaces, now }: { workspaces: Project[]; now: number },
): DataTableGroupingMode<TodoTableRow>[] {
  const workspaceNames = new Map(workspaces.map(ws => [ws.dir, ws.name || ws.dir]));
  const workspaceOrder = workspaces.map(ws => ws.dir);
  const meta = (rows: TodoTableRow[]) => <TodoCountsBar counts={countsFromItems(rows.map(r => r.todo))} />;

  return [
    {
      type: 'custom',
      value: 'workspace',
      label: 'Workspace',
      getGroupKey: row => row.workspace.dir,
      getGroupLabel: key => workspaceNames.get(key) ?? key,
      getGroupMeta: (_key, rows) => meta(rows),
      metaAlign: 'end',
      compareGroups: byDefinedOrder(workspaceOrder),
    },
    {
      type: 'custom',
      value: 'severity',
      label: 'Severity',
      getGroupKey: row => severityKey(row),
      getGroupLabel: key => SEVERITY_BUCKETS.find(b => b.key === key)?.label ?? key,
      getGroupMeta: (_key, rows) => meta(rows),
      metaAlign: 'end',
      compareGroups: byDefinedOrder(SEVERITY_BUCKETS.map(b => b.key)),
    },
    {
      type: 'custom',
      value: 'age',
      label: 'Age',
      getGroupKey: row => ageKey(row, now),
      getGroupLabel: key => AGE_LABELS[key] ?? key,
      getGroupMeta: (_key, rows) => meta(rows),
      metaAlign: 'end',
      compareGroups: byDefinedOrder(AGE_ORDER),
    },
    { type: 'none', value: 'none', label: 'None' },
  ];
}

// todoTableColumns are the full-width layout's columns. Every cell reuses the
// renderer the list row already uses, so the two layouts show the same todo the
// same way — only the arrangement differs.
//
// Only the four columns whose keys are TodoSortColumn are sortable, and the
// rest say so explicitly: DataTable treats a column as sortable unless
// `sortable: false`, and the sort state here IS the shared TodoSort preference
// — a header for any other key would write a column loadTodoSort then rejects,
// silently resetting the user's sort on the next reload.
//
// Workspace is deliberately among the unsortable ones: TodoSort sorts TodoItem,
// which carries no workspace, and grouping by workspace already organises by it.
export function todoTableColumns(
  { groupBy, tagsByDir }: { groupBy: TodoGroupBy; tagsByDir?: Map<string, TagIndex> },
): DataTableColumn<TodoTableRow>[] {
  const columns: DataTableColumn<TodoTableRow>[] = [
    {
      key: 'status',
      label: 'Status',
      sortable: false,
      shrink: true,
      accessor: row => row.todo.status,
      render: (_value, row) => <TodoStatusCell todo={row.todo} dir={row.workspace.dir} />,
    },
    {
      key: 'title',
      label: 'Title',
      sortable: true,
      grow: true,
      minWidth: 240,
      accessor: row => row.todo.title,
      render: (_value, row) => (
        <span className="block truncate font-medium text-foreground" title={row.todo.title}>
          {row.todo.title}
        </span>
      ),
    },
  ];

  // Under workspace grouping the header already names the workspace, so the
  // column would repeat it on every row.
  if (groupBy !== 'workspace') {
    columns.push({
      key: 'workspace',
      label: 'Workspace',
      sortable: false,
      shrink: true,
      maxWidth: 200,
      accessor: row => row.workspace.name || row.workspace.dir,
      render: (value, row) => (
        <span className="inline-flex min-w-0 items-center gap-1 text-muted-foreground" title={row.workspace.dir}>
          <UiFolder className="shrink-0 text-[11px]" />
          <span className="truncate">{String(value ?? '')}</span>
        </span>
      ),
    });
  }

  columns.push(
    {
      key: 'priority',
      label: 'Priority',
      sortable: true,
      shrink: true,
      accessor: row => row.todo.priority,
      render: (_value, row) => <TodoPriorityCell priority={row.todo.priority} />,
    },
    // One column per lifecycle phase, in pipeline order, each carrying that
    // phase's own status, progress and elapsed time. They come from the list
    // payload, so none of them costs a request per row.
    ...TODO_PHASES.map((phase): DataTableColumn<TodoTableRow> => ({
      key: `phase.${phase}`,
      label: phaseLabel(phase),
      sortable: false,
      shrink: true,
      accessor: row => row.todo.phases?.[phase]?.state ?? '',
      render: (_value, row) => <TodoPhaseCell todo={row.todo} phase={phase} />,
    })),
    {
      key: 'tags',
      label: 'Tags',
      sortable: false,
      accessor: row => todoVisibleLabels(row.todo),
      render: (_value, row) => {
        const index = tagsByDir?.get(row.workspace.dir);
        const labels = index ? todoVisibleLabels(row.todo) : [];
        if (!index || labels.length === 0) return null;
        return <TodoTagRow labels={labels} index={index} max={4} size="xxs" showKey={false} />;
      },
    },
    {
      key: 'created',
      label: 'Created',
      sortable: true,
      shrink: true,
      accessor: row => row.todo.created,
      kind: 'timestamp',
    },
    {
      key: 'updated',
      label: 'Last activity',
      sortable: true,
      shrink: true,
      accessor: row => row.todo.lastRun,
      kind: 'timestamp',
    },
    {
      key: 'signals',
      label: '',
      sortable: false,
      shrink: true,
      align: 'right',
      accessor: row => row.todo.ref,
      render: (_value, row) => <TodoSignalsCell todo={row.todo} />,
    },
  );

  return columns;
}

// TodoStatusCell is the list row's status affordance: a live run reports its own
// state and elapsed time, anything else falls back to the lifecycle icon.
function TodoStatusCell({ todo, dir }: { todo: TodoItem; dir: string }) {
  return (
    <span className="inline-flex items-center gap-1.5 whitespace-nowrap">
      {todo.sessionId ? (
        <SessionBadge dir={dir} sessionId={todo.sessionId} status={todo.status} live={isLiveRun(todo)} />
      ) : (
        <StatusIcon status={todo.status} />
      )}
      <span className="text-xs text-muted-foreground">{statusLabel(todo.status)}</span>
    </span>
  );
}

// Severity is only the row's left border in the list layout. A table has the
// room to make it a column, which is also what makes it sortable.
function TodoPriorityCell({ priority }: { priority: TodoItem['priority'] }) {
  const Icon = priorityIcon(priority);
  return (
    <span
      className={`inline-flex items-center gap-1 whitespace-nowrap rounded border px-1.5 py-0.5 text-[11px] ${priorityBadgeClass(priority)}`}
    >
      <Icon className="text-[11px]" />
      {priority}
    </span>
  );
}

function TodoSignalsCell({ todo }: { todo: TodoItem }) {
  return (
    <span className="inline-flex items-center justify-end gap-2 text-xs text-muted-foreground">
      {todo.hasPlan && <TodoPlanIndicator />}
      {todo.hasVerification && <TodoVerificationIndicator />}
      {todo.diff && <TodoDiffBadge diff={todo.diff} />}
    </span>
  );
}

// TodoTable is the full-width layout's list: every workspace's todos in one
// sortable, groupable table under a filter bar wide enough to show its facets
// inline. It reuses the split layout's filter/sort/group preferences and its
// data pipeline, so switching layouts changes the arrangement, not the contents.
export function TodoTable({ todos, projectsLoaded }: {
  todos: WorkspaceTodos;
  projectsLoaded: boolean;
}) {
  const {
    workspaces, byDir, filters, timeRange, setTimeRange, density, groupBy, setGroupBy,
    sortBy, setSortBy, select, selected, loadingList, refresh, error, selection, tagsByDir,
  } = todos;
  const { facets, range } = useTodoFilterBar(todos);
  // The same descriptors the split layout renders, so a bulk action added on
  // the server appears in both without either layout being touched.
  // The label menu needs what the selection currently holds and the taxonomy to
  // offer it from; both are already loaded for the rows themselves.
  const bulkContext = useTodoBulkContext(todos);
  const selectionActions = useTodoBulkToolbar({
    selection: selection.selection,
    ...bulkContext,
    onApplied: refresh,
  });

  // Filtering and ordering stay with the helpers the list layout uses, so both
  // layouts show the same todos in the same order. manualSort below keeps
  // DataTable from re-sorting and losing todoComparator's title tie-break.
  const rows = useMemo<TodoTableRow[]>(() => {
    const resolved = resolveRange(timeRange, Date.now());
    const cmp = todoComparator(sortBy);
    return flattenTodos(workspaces, byDir)
      .filter(entry => isEntryVisible(entry, filters, resolved))
      .sort((a, b) => cmp(a.todo, b.todo))
      // The index signature is DataTable's row constraint, not a shape any
      // column reads: every column resolves its value through an accessor.
      .map(entry => entry as TodoTableRow);
  }, [workspaces, byDir, filters, timeRange, sortBy]);

  const columns = useMemo(() => todoTableColumns({ groupBy, tagsByDir }), [groupBy, tagsByDir]);
  // Re-derived whenever the rows are, so the age buckets track the clock the way
  // the list layout's per-render bucketTodos(…, Date.now()) does. Memoised only
  // on `now` staying put between refreshes would leave a dashboard left open
  // overnight still calling yesterday's todos "Today".
  const groupingModes = useMemo(
    () => todoGroupingModes({ workspaces, now: Date.now() }),
    [workspaces, rows],
  );

  // projectError is not surfaced here: the Todos body already renders it as a
  // banner above both layouts, so repeating it inside the table would double it.
  const RefreshIcon = loadingList ? Spinner : UiRefresh;

  return (
    <div className="flex h-full min-h-0 flex-col">
      <DataTable<TodoTableRow>
        className="min-h-0 flex-1"
        data={rows}
        columns={columns}
        getRowId={todoTableRowId}
        // The facets are caller-owned and already applied to `data` above; no
        // column filters are generated, so the bar carries exactly the controls
        // the sidebar toolbar has, with room to show them inline.
        autoFilter={false}
        externalFilters={facets}
        externalTimeRange={range}
        showGlobalFilter
        globalFilterPlaceholder="Search todos"
        filterBarProps={{
          className: 'shrink-0 border-b border-border bg-card px-2',
          // This row carries search, the workspace facet, the selection notice
          // and its actions, the grouping picker, sort and the column menu.
          // FilterBar's default refuses to wrap above md, so on a laptop the
          // right-hand controls ran off the viewport and the selection count
          // was clipped — unreachable, not merely cramped. Wrapping only takes
          // effect when it genuinely does not fit.
          overflowMode: 'wrap',
          trailing: (
            <>
              {timeRange && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => setTimeRange(null)}
                  title={`Clear ${TODO_ACTIVITY_FILTER_LABEL.toLowerCase()} filter`}
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
          ),
        }}
        groupingModes={groupingModes}
        groupingMode={groupBy}
        onGroupingModeChange={value => setGroupBy(value as TodoGroupBy)}
        sort={{ key: sortBy.column, dir: sortBy.dir }}
        onSortChange={next => setSortBy(next ? { column: next.key as TodoSortColumn, dir: next.dir } : defaultTodoSort())}
        manualSort
        // Density is the navbar picker's preference. Deliberately not persisted
        // here: DataTable would write its own key and the two would drift.
        density={density}
        showDensityControl={false}
        // Selection is unconditional: the checkbox column is what advertises
        // that bulk actions exist at all, and gating it behind a mode meant a
        // user had to find the mode before discovering the capability.
        rowSelection={{
          selectedRowIds: [...selection.selection],
          onSelectionChange: ids => selection.replaceSelection(ids),
          // The header checkbox reaches only the revealed window (see
          // clientReveal below), so on a workspace of hundreds "select all"
          // silently stops at the batch size. This offers the rest explicitly
          // rather than letting the user believe they have the whole match.
          selectAllPages: {
            noun: 'todos',
            scopes: [{
              total: rows.length,
              // Every filtered row is already in memory — the reveal window
              // limits rendering, not the data — so widening is local.
              onSelectAll: () => selection.replaceSelection(rows.map(todoTableRowId)),
            }],
          },
        }}
        selectionActions={selectionActions}
        // Selecting a todo swaps the filter row for the bulk bar in place, so
        // the rows being acted on never move under the cursor.
        selectionBar="takeover"
        // onRowClick only, no getRowHref: DataTable's row link routes
        // client-side through clicky-ui's RouterAdapter, and pr/ui mounts no
        // RouterProvider — the link would fall back to a plain <a> and hard
        // reload the app on every row click. Selection still reaches the URL
        // through select() -> navigateTodo.
        onRowClick={row => select({ dir: row.workspace.dir, ref: row.todo.ref })}
        getRowClassName={row => (
          selected?.dir === row.workspace.dir && selected?.ref === row.todo.ref ? 'bg-primary/5' : undefined
        )}
        // clicky has no virtualized table, and a single workspace already carries
        // hundreds of todos — reveal them in batches instead of paginating.
        clientReveal={{ batchSize: 100 }}
        loading={loadingList && rows.length === 0}
        loadingMessage="Loading todos"
        error={error || undefined}
        emptyMessage={projectsLoaded ? 'No todos match these filters' : 'Loading workspaces'}
      />
    </div>
  );
}
