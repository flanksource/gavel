import { SplitPane } from '@flanksource/clicky-ui/components';
import type { Project } from '../types';
import { GavelIcon } from './GavelIcon';
import { TodoCountsBar } from './todos/format';
import { useWorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoDetail } from './todos/TodoDetail';
import { CreateTodoDialog } from './todos/CreateTodoDialog';
import { TodoFilterBar } from './todos/TodoFilterBar';

export function TodoView({ projects }: { projects: Project[] }) {
  const {
    workspaces, byDir, loadingList, error, aggregate,
    selected, setSelected, detail, loadingDetail,
    refresh, showCreate, setShowCreate, created, updateItem, deleted,
    hiddenStatuses, toggleStatus,
  } = useWorkspaceTodos(projects);

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex items-center justify-between gap-2 border-b border-border bg-background px-3 py-2">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">Todos</span>
          <span className="text-xs text-muted-foreground">{workspaces.length} workspace{workspaces.length === 1 ? '' : 's'}</span>
        </div>
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
      </div>
      {aggregate.total > 0 && (
        <div className="flex items-center gap-2 border-b border-border bg-background px-3 py-1.5">
          <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Filter</span>
          <TodoFilterBar counts={aggregate} hidden={hiddenStatuses} onToggle={toggleStatus} />
        </div>
      )}
      {error && <div className="border-b border-border px-3 py-1 text-xs text-red-600">{error}</div>}
      <CreateTodoDialog open={showCreate} onClose={() => setShowCreate(false)} workspaces={workspaces} onCreated={created} />
      <div className="min-h-0 flex-1 overflow-hidden">
        <SplitPane
          left={
            workspaces.length > 0 ? (
              <div className="h-full overflow-y-auto">
                {workspaces.map(ws => (
                  <WorkspaceTodoGroup
                    key={ws.dir}
                    workspace={ws}
                    data={byDir[ws.dir]}
                    hiddenStatuses={hiddenStatuses}
                    selectedRef={selected?.dir === ws.dir ? selected.ref : ''}
                    onSelect={ref => setSelected({ dir: ws.dir, ref, provider: ws.todoProvider || 'auto' })}
                  />
                ))}
              </div>
            ) : (
              <div className="p-6 text-center text-sm text-muted-foreground">
                <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-3xl" />
                <p>{loadingList ? 'Loading' : 'No workspaces configured'}</p>
              </div>
            )
          }
          right={
            <TodoDetail
              todo={detail}
              loading={loadingDetail}
              dir={selected?.dir ?? ''}
              provider={selected?.provider ?? 'auto'}
              onChanged={updateItem}
              onDeleted={deleted}
            />
          }
        />
      </div>
    </div>
  );
}
