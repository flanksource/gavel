import { GavelIcon } from './GavelIcon';
import { ReactGrabHelp } from './ReactGrabHelp';
import { TodoCountsBar, TodoDensityPicker, TodoGroupByPicker } from './todos/format';
import type { WorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoBucketGroup } from './todos/TodoBucketGroup';
import { bucketTodos, flattenTodos } from './todos/todoGroup';
import { TodoDetail } from './todos/TodoDetail';
import { TodoFilterBar } from './todos/TodoFilterBar';

// The Todos tab renders its chrome into the shared AppShell's body slots — a
// bodyHeader/bodyActions row, a filter toolbar, and an independently-scrolling
// bodySidebar (the workspace list) beside the detail pane — rather than nesting
// its own header + SplitPane inside the shell's content. Each piece below is one
// slot, driven by the shared useWorkspaceTodos data layer the App owns.

// TodoBodyHeader is the AppShell bodyHeader (left): the title, workspace count,
// and any list-load error.
export function TodoBodyHeader({ todos }: { todos: WorkspaceTodos }) {
  const { workspaces, error } = todos;
  return (
    <div className="flex items-center gap-2">
      <span className="text-sm font-semibold text-foreground">Todos</span>
      <span className="text-xs text-muted-foreground">
        {workspaces.length} workspace{workspaces.length === 1 ? '' : 's'}
      </span>
      {error && <span className="text-xs text-red-600">{error}</span>}
    </div>
  );
}

// TodoBodyActions is the AppShell bodyActions (right): aggregate counts plus the
// New and Refresh controls.
export function TodoBodyActions({ todos }: { todos: WorkspaceTodos }) {
  const { workspaces, aggregate, loadingList, refresh, setShowCreate } = todos;
  return (
    <div className="flex items-center gap-2">
      <TodoCountsBar counts={aggregate} />
      <button
        type="button"
        onClick={() => setShowCreate(true)}
        disabled={workspaces.length === 0}
        title="New todo"
        className="inline-flex h-8 items-center gap-1 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted disabled:opacity-50"
      >
        <GavelIcon name="codicon:add" className="text-xs" />
        New
      </button>
      <ReactGrabHelp />
      <button
        type="button"
        onClick={refresh}
        disabled={loadingList}
        title="Refresh todos"
        className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        aria-label="Refresh todos"
      >
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:refresh'} className="text-sm" />
      </button>
    </div>
  );
}

// TodoFilterToolbar is the AppShell toolbar content for the Todos tab. The App
// only mounts it when there are todos to filter, so the toolbar row stays hidden
// on an empty list.
export function TodoFilterToolbar({ todos }: { todos: WorkspaceTodos }) {
  const { aggregate, hiddenStatuses, toggleStatus, density, setDensity, groupBy, setGroupBy } = todos;
  return (
    <>
      <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Filter</span>
      <TodoFilterBar counts={aggregate} hidden={hiddenStatuses} onToggle={toggleStatus} />
      <div className="ml-auto flex items-center gap-2">
        <TodoGroupByPicker groupBy={groupBy} onChange={setGroupBy} />
        <TodoDensityPicker density={density} onChange={setDensity} />
      </div>
    </>
  );
}

// TodoWorkspaceList is the AppShell bodySidebar: every configured workspace's
// todos, grouped and independently scrollable beside the detail pane. The
// group-by preference picks the grouping: workspace (the default, with batch-run
// controls) or severity/age buckets that span workspaces.
export function TodoWorkspaceList({ todos }: { todos: WorkspaceTodos }) {
  const { workspaces, byDir, hiddenStatuses, density, groupBy, selected, select, refresh, loadingList } = todos;
  if (workspaces.length === 0) {
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-3xl" />
        <p>{loadingList ? 'Loading' : 'No workspaces configured'}</p>
      </div>
    );
  }
  if (groupBy === 'workspace') {
    return (
      <div>
        {workspaces.map(ws => (
          <WorkspaceTodoGroup
            key={ws.dir}
            workspace={ws}
            data={byDir[ws.dir]}
            hiddenStatuses={hiddenStatuses}
            density={density}
            selectedRef={selected?.dir === ws.dir ? selected.ref : ''}
            onSelect={ref => select({ dir: ws.dir, ref, provider: ws.todoProvider || 'auto' })}
            multiSelect
            onRunStarted={refresh}
          />
        ))}
      </div>
    );
  }
  const buckets = bucketTodos(flattenTodos(workspaces, byDir), groupBy, Date.now());
  if (buckets.length === 0) {
    return (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-3xl" />
        <p>{loadingList ? 'Loading' : 'No todos'}</p>
      </div>
    );
  }
  return (
    <div>
      {buckets.map(bucket => (
        <TodoBucketGroup
          key={bucket.key}
          bucket={bucket}
          selected={selected}
          onSelect={entry => select({ dir: entry.workspace.dir, ref: entry.todo.ref, provider: entry.workspace.todoProvider || 'auto' })}
          hiddenStatuses={hiddenStatuses}
          density={density}
        />
      ))}
    </div>
  );
}

// TodoDetailPane is the AppShell body-main: the selected todo's detail (or the
// empty "Select a todo" prompt).
export function TodoDetailPane({ todos }: { todos: WorkspaceTodos }) {
  const { detail, loadingDetail, selected, updateItem, deleted, workspaces, transferred } = todos;
  return (
    <TodoDetail
      todo={detail}
      loading={loadingDetail}
      dir={selected?.dir ?? ''}
      provider={selected?.provider ?? 'auto'}
      onChanged={updateItem}
      onDeleted={deleted}
      workspaces={workspaces}
      onTransferred={transferred}
    />
  );
}
