import { Button, ListMenu } from '@flanksource/clicky-ui/components';
import type { Project } from '../types';
import { GavelIcon } from './GavelIcon';
import { useWorkspaceTodos } from './todos/useWorkspaceTodos';
import { WorkspaceTodoGroup } from './todos/WorkspaceTodoGroup';
import { TodoBucketGroup } from './todos/TodoBucketGroup';
import { TodoDetail } from './todos/TodoDetail';
import { TodoGroupByMenu } from './todos/TodoGroupByMenu';
import { TodoFilterMenu } from './todos/TodoFilterMenu';
import { CreateTodoDialog } from './todos/CreateTodoDialog';
import { bucketTodos, flattenTodos } from './todos/todoGroup';

// MenubarTodos is the compact, single-column todos view for the menubar popover.
// It mirrors the PRs tab's master-detail idiom: a workspace-grouped list, and
// tapping a todo swaps in its detail behind a back button. It shares the data
// layer with the dashboard TodoView via useWorkspaceTodos, so both stay in sync.
//
// A thin tab strip above the list carries the same grouping/filter controls as
// the dashboard (folded into dropdowns to fit the narrow popover) plus a New
// control, so todos can be regrouped, filtered, and created without leaving the
// menubar.
export function MenubarTodos({ projects }: { projects: Project[] }) {
  const {
    workspaces, byDir, loadingList, aggregate,
    selected, setSelected, detail, loadingDetail,
    updateItem, deleted, hiddenStatuses, toggleStatus,
    groupBy, setGroupBy, showCreate, setShowCreate, created,
  } = useWorkspaceTodos(projects);

  if (selected) {
    return (
      <div className="flex h-full min-h-0 flex-col">
        <div className="flex h-10 shrink-0 items-center gap-2 border-b border-border px-2">
          <Button
            variant="ghost"
            size="icon"
            type="button"
            onClick={() => setSelected(null)}
            className="h-8 w-8 rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            title="Back to todos"
            aria-label="Back to todos"
          >
            <GavelIcon name="codicon:arrow-left" className="text-base" />
          </Button>
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

  // Severity/age grouping flattens todos across workspaces into buckets; the
  // default 'workspace' grouping keeps the per-workspace sections (the only mode
  // that supports batch runs on the dashboard).
  const buckets = groupBy === 'workspace' ? null : bucketTodos(flattenTodos(workspaces, byDir), groupBy, Date.now());

  return (
    <div className="flex h-full min-h-0 flex-col">
      {workspaces.length > 0 && (
        <div className="flex shrink-0 items-center gap-1.5 border-b border-border px-2 py-1.5">
          {aggregate.total > 0 && (
            <>
              <TodoGroupByMenu groupBy={groupBy} onChange={setGroupBy} />
              <TodoFilterMenu counts={aggregate} hidden={hiddenStatuses} onToggle={toggleStatus} />
            </>
          )}
          <Button
            variant="ghost"
            type="button"
            onClick={() => setShowCreate(true)}
            title="New todo"
            className="ml-auto inline-flex h-8 items-center justify-start gap-1 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <GavelIcon name="codicon:add" className="text-xs" />
            New
          </Button>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-y-auto">
        {workspaces.length === 0 ? (
          <div className="px-3 py-6 text-center text-xs text-muted-foreground">
            <GavelIcon name={loadingList ? 'svg-spinners:ring-resize' : 'codicon:check'} className="mb-2 text-2xl" />
            <p>{loadingList ? 'Loading' : 'No workspaces configured'}</p>
          </div>
        ) : buckets ? (
          buckets.length > 0 ? (
            <ListMenu>
              {buckets.map(bucket => (
                <TodoBucketGroup
                  key={bucket.key}
                  bucket={bucket}
                  selected={selected}
                  onSelect={entry => setSelected({ dir: entry.workspace.dir, ref: entry.todo.ref, provider: entry.workspace.todoProvider || 'auto' })}
                  hiddenStatuses={hiddenStatuses}
                />
              ))}
            </ListMenu>
          ) : (
            <div className="px-3 py-6 text-center text-xs text-muted-foreground">No todos</div>
          )
        ) : (
          <ListMenu>
            {workspaces.map(ws => (
              <WorkspaceTodoGroup
                key={ws.dir}
                workspace={ws}
                data={byDir[ws.dir]}
                hiddenStatuses={hiddenStatuses}
                onToggleStatus={toggleStatus}
                selectedRef=""
                onSelect={ref => setSelected({ dir: ws.dir, ref, provider: ws.todoProvider || 'auto' })}
              />
            ))}
          </ListMenu>
        )}
      </div>

      <CreateTodoDialog open={showCreate} onClose={() => setShowCreate(false)} workspaces={workspaces} onCreated={created} />
    </div>
  );
}
