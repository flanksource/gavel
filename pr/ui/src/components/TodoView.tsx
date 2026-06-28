import type { ReactNode } from 'react';
import { Button, ListMenu, TimeRange } from '@flanksource/clicky-ui/components';
import { GavelIcon } from './GavelIcon';
import { TodoDensityPicker } from './todos/format';
import type { WorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoBucketGroup } from './todos/TodoBucketGroup';
import { bucketTodos, flattenTodos } from './todos/todoGroup';
import { resolveRange } from './todos/todoTimeRange';
import { TodoDetail } from './todos/TodoDetail';
import { TodoFilterBar } from './todos/TodoFilterBar';
import { TodoGroupByMenu } from './todos/TodoGroupByMenu';

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
      <GavelIcon name="codicon:add" className="text-xs" />
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

// TodoSidebarActions sits above the todo tree in the AppShell bodySidebar. The
// filter pills are also the count surface, so the sidebar has one compact row:
// grouping, status filters, time filtering, and refresh.
export function TodoSidebarActions({ todos }: { todos: WorkspaceTodos }) {
  const { aggregate, hiddenStatuses, toggleStatus, groupBy, setGroupBy, timeRange, setTimeRange, loadingList, refresh } = todos;
  return (
    <div className="flex shrink-0 flex-wrap items-center gap-1.5 border-b border-border bg-card px-2 py-1.5">
      {aggregate.total > 0 && (
        <>
          <TodoGroupByMenu groupBy={groupBy} onChange={setGroupBy} />
          <TodoFilterBar counts={aggregate} hidden={hiddenStatuses} onToggle={toggleStatus} />
        </>
      )}
      <div className="min-w-0 flex-1" />
      {aggregate.total > 0 && (
        <div className="flex items-center gap-1.5">
          <TimeRange
            kind="date"
            label="Active"
            emptyLabel="Any time"
            from={timeRange?.from}
            to={timeRange?.to}
            onApply={(from, to) => setTimeRange({ from, to })}
            presets={['hr', 'day', 'wk+']}
            align="right"
          />
          {timeRange && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setTimeRange(null)}
              title="Clear time filter"
              aria-label="Clear time filter"
              className="h-8 w-8 text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <GavelIcon name="codicon:close" className="text-xs" />
            </Button>
          )}
        </div>
      )}
      <Button
        type="button"
        variant="ghost"
        size="icon"
        onClick={refresh}
        disabled={loadingList}
        title="Refresh todos"
        className="h-8 w-8 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
        aria-label="Refresh todos"
      >
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:refresh'} className="text-sm" />
      </Button>
    </div>
  );
}

// TodoWorkspaceList is the AppShell bodySidebar: every configured workspace's
// todos, grouped and independently scrollable beside the detail pane. The
// group-by preference picks the grouping: workspace (the default, with batch-run
// controls) or severity/age buckets that span workspaces.
export function TodoWorkspaceList({ todos }: { todos: WorkspaceTodos }) {
  const { workspaces, byDir, hiddenStatuses, toggleStatus, density, groupBy, timeRange, selected, select, refresh, loadingList } = todos;
  // Resolve the activity range to absolute bounds once per render so every group
  // filters against the same instant.
  const range = resolveRange(timeRange, Date.now());
  let content: ReactNode;
  if (workspaces.length === 0) {
    content = (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-3xl" />
        <p>{loadingList ? 'Loading' : 'No workspaces configured'}</p>
      </div>
    );
  } else if (groupBy === 'workspace') {
    content = (
      <ListMenu>
        {workspaces.map(ws => (
          <WorkspaceTodoGroup
            key={ws.dir}
            workspace={ws}
            data={byDir[ws.dir]}
            hiddenStatuses={hiddenStatuses}
            onToggleStatus={toggleStatus}
            range={range}
            density={density}
            selectedRef={selected?.dir === ws.dir ? selected.ref : ''}
            onSelect={ref => select({ dir: ws.dir, ref, provider: ws.todoProvider || 'auto' })}
            multiSelect
            onRunStarted={refresh}
          />
        ))}
      </ListMenu>
    );
  } else {
    const buckets = bucketTodos(flattenTodos(workspaces, byDir), groupBy, Date.now());
    content = buckets.length === 0 ? (
      <div className="p-6 text-center text-sm text-muted-foreground">
        <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-3xl" />
        <p>{loadingList ? 'Loading' : 'No todos'}</p>
      </div>
    ) : (
      <ListMenu>
        {buckets.map(bucket => (
          <TodoBucketGroup
            key={bucket.key}
            bucket={bucket}
            selected={selected}
            onSelect={entry => select({ dir: entry.workspace.dir, ref: entry.todo.ref, provider: entry.workspace.todoProvider || 'auto' })}
            hiddenStatuses={hiddenStatuses}
            range={range}
            density={density}
          />
        ))}
      </ListMenu>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <TodoSidebarActions todos={todos} />
      <div className="min-h-0 flex-1 overflow-y-auto">
        {content}
      </div>
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
