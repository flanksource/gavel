import type { Project } from '../types';
import { GavelIcon } from './GavelIcon';
import { useWorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoDetail } from './todos/TodoDetail';

// MenubarTodos is the compact, single-column todos view for the menubar popover.
// It mirrors the PRs tab's master-detail idiom: a workspace-grouped list, and
// tapping a todo swaps in its detail behind a back button. It shares the data
// layer with the dashboard TodoView via useWorkspaceTodos, so both stay in sync.
export function MenubarTodos({ projects }: { projects: Project[] }) {
  const {
    workspaces, byDir, loadingList,
    selected, setSelected, detail, loadingDetail,
    updateItem, deleted,
  } = useWorkspaceTodos(projects);

  if (selected) {
    return (
      <div className="flex h-full min-h-0 flex-col">
        <div className="flex h-10 shrink-0 items-center gap-2 border-b border-border px-2">
          <button
            type="button"
            onClick={() => setSelected(null)}
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            title="Back to todos"
            aria-label="Back to todos"
          >
            <GavelIcon name="codicon:arrow-left" className="text-base" />
          </button>
          <span className="min-w-0 flex-1 truncate text-sm font-semibold">{detail?.title ?? 'Todo'}</span>
        </div>
        <div className="min-h-0 flex-1">
          <TodoDetail
            todo={detail}
            loading={loadingDetail}
            dir={selected.dir}
            provider={selected.provider}
            onChanged={updateItem}
            onDeleted={deleted}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="min-h-0 flex-1 overflow-y-auto">
        {workspaces.length > 0 ? (
          workspaces.map(ws => (
            <WorkspaceTodoGroup
              key={ws.dir}
              workspace={ws}
              data={byDir[ws.dir]}
              selectedRef=""
              onSelect={ref => setSelected({ dir: ws.dir, ref, provider: ws.todoProvider || 'auto' })}
            />
          ))
        ) : (
          <div className="px-3 py-6 text-center text-xs text-muted-foreground">
            <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-2xl" />
            <p>{loadingList ? 'Loading' : 'No workspaces configured'}</p>
          </div>
        )}
      </div>
    </div>
  );
}
