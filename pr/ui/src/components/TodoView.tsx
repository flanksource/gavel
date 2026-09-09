import { useMemo, useState, type ReactNode } from 'react';
import { Button, ListMenu } from '@flanksource/clicky-ui/components';
import { UiAdd, UiCheck } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../icons/Spinner';
import { TodoDensityPicker, TodoLayoutPicker } from './todos/format';
import type { WorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoBucketGroup } from './todos/TodoBucketGroup';
import { bucketTodos, flattenTodos } from './todos/todoGroup';
import { isWorkspaceShown } from './todos/todoFilter';
import { resolveRange } from './todos/todoTimeRange';
import { TodoDetail } from './todos/TodoDetail';
import { TodoDetailStack } from './todos/TodoDetailStack';
import { TodoTable, todoTableColumns, type TodoTableRow } from './todos/TodoTable';
import type { TodoNavigationControlsProps } from './todos/TodoNavigationControls';
import { filterTodoNavigationEntries, orderedTodoNavigationEntries, todoNavigationState } from './todos/todoNavigation';
import { TodoToolbar } from './todos/TodoToolbar';

// The Todos tab renders its chrome into the shared AppShell's body slots: top-bar
// actions and an independently-scrolling bodySidebar (the workspace list) beside
// the detail pane. Each piece below is one slot, driven by the shared
// useWorkspaceTodos data layer the App owns.

// TodoNewButton is the primary "create todo" action. It lives in the AppShell's
// top-bar actions cluster (the action header) alongside the other global
// controls, not the body row. Disabled until at least one workspace exists.
export function TodoNewButton({ todos }: { todos: WorkspaceTodos }) {
  const { workspaces, setShowCreate } = todos;
  return (
    <Button
      type="button"
      variant="ghost"
      onClick={() => setShowCreate(true)}
      disabled={workspaces.length === 0}
      title="New todo"
      className="inline-flex h-8 items-center justify-start gap-1 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted disabled:opacity-50"
    >
      <UiAdd className="text-xs" />
      New
    </Button>
  );
}

// TodoNavbarDensityPicker is a top-bar list display control for the Todos tab.
// Filter controls live with the sidebar tree they affect; density stays in the
// navbar because it is a display preference rather than a list filter.
export function TodoNavbarDensityPicker({ todos }: { todos: WorkspaceTodos }) {
  const { aggregate, density, setDensity } = todos;
  if (aggregate.total === 0) return null;
  return <TodoDensityPicker density={density} onChange={setDensity} />;
}

// TodoNavbarLayoutPicker switches the tab between the master-detail split and
// the full-width table. It sits beside the density picker for the same reason:
// both are display preferences, not list filters. Unlike density it renders
// even with no todos loaded — the layout is how you look at the tab, so it must
// not disappear exactly when an empty full-width table would need explaining.
export function TodoNavbarLayoutPicker({ todos }: { todos: WorkspaceTodos }) {
  const { layout, setLayout } = todos;
  return <TodoLayoutPicker layout={layout} onChange={setLayout} />;
}

// TodoSidebarActions sits above the todo tree in the AppShell bodySidebar. It is
// the shared TodoToolbar — the same row the menubar and mobile layouts render —
// so there is one filter surface rather than a weaker duplicate per layout.
export function TodoSidebarActions({ todos }: { todos: WorkspaceTodos }) {
  return <TodoToolbar todos={todos} />;
}

// TodoWorkspaceList is the AppShell bodySidebar: every configured workspace's
// todos, grouped and independently scrollable beside the detail pane. The
// group-by preference picks the grouping: workspace (the default) or
// severity/age buckets that span workspaces.
//
// The workspaces are derived from /api/projects, so an empty sidebar means one of
// three things and must not conflate them: projectsLoaded is false while the
// first fetch is in flight, projectError holds the reason it failed, and only
// with neither set is "no workspaces configured" the truth.
export function TodoWorkspaceList({ todos, projectsLoaded, projectError }: {
  todos: WorkspaceTodos;
  projectsLoaded: boolean;
  projectError?: string;
}) {
  const { workspaces, byDir, filters, toggleStatus, density, groupBy, sortBy, timeRange, selected, select, loadingList, error, selection, tagsByDir } = todos;
  // Resolve the activity range to absolute bounds so every group filters against
  // the same instant. Keyed on a minute-granularity clock rather than a raw
  // Date.now(): a relative range ("last 7 days") only needs minute accuracy, and
  // reading the raw clock in the render body produced a new bounds object every
  // render, which was passed to every group and re-filtered and re-sorted every
  // workspace underneath.
  const minute = Math.floor(Date.now() / 60_000);
  const range = useMemo(() => resolveRange(timeRange, minute * 60_000), [timeRange, minute]);
  // Bucketing is only used by the non-workspace groupings, but it is computed
  // here because hooks cannot live inside the branch below. Same
  // minute-granularity clock as `range`: age buckets do not need sub-minute
  // precision, and re-bucketing every todo on every render rebuilt every
  // bucket's entry array and re-rendered every row in the list.
  const buckets = useMemo(
    () => (groupBy === 'workspace' ? [] : bucketTodos(flattenTodos(workspaces, byDir), groupBy, minute * 60_000, sortBy)),
    [groupBy, workspaces, byDir, minute, sortBy],
  );
  // Waiting on either fetch is "loading": the workspaces come from
  // /api/projects and the todos from the per-workspace lists.
  const pending = loadingList || !projectsLoaded;
  const EmptyIcon = pending ? Spinner : UiCheck;
  let content: ReactNode;
  if (workspaces.length === 0) {
    content = projectError ? (
      <div role="alert" className="p-6 text-center text-sm text-destructive">{projectError}</div>
    ) : (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <EmptyIcon className="mb-2 text-3xl" />
        <p>{pending ? 'Loading' : 'No workspaces configured'}</p>
      </div>
    );
  } else if (groupBy === 'workspace') {
    // An excluded workspace drops out whole rather than rendering an empty
    // section: its header carries counts, and a section reading "0 open" would
    // look like a workspace that had gone quiet rather than one filtered away.
    const shown = workspaces.filter(ws => isWorkspaceShown(filters, ws.dir));
    content = shown.length === 0 ? (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <EmptyIcon className="mb-2 text-3xl" />
        <p>{pending ? 'Loading' : 'No workspaces match the filter'}</p>
      </div>
    ) : (
      <ListMenu>
        {shown.map(ws => (
          <WorkspaceTodoGroup
            key={ws.dir}
            workspace={ws}
            data={byDir[ws.dir]}
            filters={filters}
            onToggleStatus={toggleStatus}
            range={range}
            density={density}
            sortBy={sortBy}
            selectedRef={selected?.dir === ws.dir ? selected.ref : ''}
            onSelect={select}
            selection={selection}
            tags={tagsByDir?.get(ws.dir)}
          />
        ))}
      </ListMenu>
    );
  } else {
    content = buckets.length === 0 ? (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <EmptyIcon className="mb-2 text-3xl" />
        <p>{loadingList ? 'Loading' : 'No todos'}</p>
      </div>
    ) : (
      <ListMenu>
        {buckets.map(bucket => (
          <TodoBucketGroup
            key={bucket.key}
            bucket={bucket}
            selected={selected}
            onSelect={select}
            filters={filters}
            range={range}
            density={density}
            selection={selection}
            tagsByDir={tagsByDir}
          />
        ))}
      </ListMenu>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <TodoSidebarActions todos={todos} />
      {error && (
        <div role="alert" className="shrink-0 border-b border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          {error}
        </div>
      )}
      <div className="min-h-0 flex-1 overflow-y-auto">
        {content}
      </div>
    </div>
  );
}

// TodoFullPane is the AppShell body-main in the full-width layout, where there
// is no body sidebar to hold the list. The table owns the viewport until a todo
// is selected, then the detail covers it behind a back arrow — the same shape
// MenubarTodos uses, through the same TodoDetailStack, which keeps the table
// mounted underneath so Back returns to the row it left from. Selection is
// route-backed (/todos/{ref}), so back/forward and deep links work without
// either half knowing about the other.
//
// Unlike the menubar it keeps passing workspaces/onTransferred, so "Move to
// project" does not silently vanish when the layout changes.
function useTodoNavigator(todos: WorkspaceTodos, query: string, enabled: boolean) {
  const { workspaces, byDir, filters, groupBy, sortBy, timeRange, selected, select, tagsByDir } = todos;
  const columns = useMemo(() => todoTableColumns({ groupBy, tagsByDir }), [groupBy, tagsByDir]);
  const entries = useMemo(() => orderedTodoNavigationEntries({
    workspaces,
    byDir,
    filters,
    groupBy,
    sortBy,
    timeRange,
    now: Date.now(),
  }), [workspaces, byDir, filters, groupBy, sortBy, timeRange]);
  const matched = useMemo(
    () => filterTodoNavigationEntries(entries, columns, query),
    [entries, columns, query],
  );
  const state = enabled ? todoNavigationState(matched, selected) : null;
  const navigation: TodoNavigationControlsProps | undefined = state ? {
    position: state.position,
    total: state.total,
    canPrevious: !!state.previous,
    canNext: !!state.next,
    onPrevious: () => state.previous && select(state.previous),
    onNext: () => state.next && select(state.next),
  } : undefined;
  // `matched` is handed to DataTable as its `data`. Pass the memoised array
  // through by reference — the previous `matched.map(e => e as TodoTableRow)`
  // allocated a fresh array on every render for a cast that changes nothing at
  // runtime, and DataTable keys its row window on `data` identity, so every
  // keystroke in the search box and every todos state change threw away every
  // row past the first batch and re-revealed them.
  return { columns, rows: matched as TodoTableRow[], navigation };
}

export function TodoFullPane({ todos, projectsLoaded, navigationEnabled = true }: {
  todos: WorkspaceTodos;
  projectsLoaded: boolean;
  navigationEnabled?: boolean;
}) {
  const [query, setQuery] = useState('');
  const { detail, loadingDetail, detailError, selected, select, updateItem, deleted, workspaces, transferred } = todos;
  const navigator = useTodoNavigator(todos, query, navigationEnabled);
  return (
    <TodoDetailStack
      list={(
        <TodoTable
          todos={todos}
          projectsLoaded={projectsLoaded}
          rows={navigator.rows}
          columns={navigator.columns}
          query={query}
          onQueryChange={setQuery}
        />
      )}
      detail={selected && (
        <TodoDetail
          todo={detail}
          loading={loadingDetail}
          loadError={detailError}
          dir={selected.dir}
          onChanged={updateItem}
          onDeleted={deleted}
          onBack={() => select(null)}
          workspaces={workspaces}
          onTransferred={transferred}
          navigation={navigator.navigation}
        />
      )}
    />
  );
}

// TodoDetailPane is the AppShell body-main in the split layout: the selected
// todo's detail (or the empty "Select a todo" prompt).
export function TodoDetailPane({ todos, navigationEnabled = true }: { todos: WorkspaceTodos; navigationEnabled?: boolean }) {
  const { detail, loadingDetail, detailError, selected, updateItem, deleted, workspaces, transferred } = todos;
  const { navigation } = useTodoNavigator(todos, '', navigationEnabled);
  return (
    <TodoDetail
      todo={detail}
      loading={loadingDetail}
      loadError={detailError}
      dir={selected?.dir ?? ''}
      onChanged={updateItem}
      onDeleted={deleted}
      workspaces={workspaces}
      onTransferred={transferred}
      navigation={navigation}
    />
  );
}
