import { Button, ListMenu } from '@flanksource/clicky-ui/components';
import { UiAdd, UiCheck } from '@flanksource/clicky-ui/icons';
import type { Project } from '../types';
import { Spinner } from '../icons/Spinner';
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
// The workspace list is derived from /api/projects, so an empty `projects` means
// one of three things and this view must not conflate them: projectsLoaded is
// false while the first fetch is in flight, projectError holds the reason it
// failed, and only with neither set is "no workspaces configured" the truth.
export function MenubarTodos({ projects, projectsLoaded, projectError }: {
  projects: Project[];
  projectsLoaded: boolean;
  projectError?: string;
}) {
  const {
    workspaces, byDir, loadingList, aggregate,
    selected, select, detail, loadingDetail, detailError, error,
    updateItem, deleted, hiddenStatuses, toggleStatus,
    groupBy, setGroupBy, showCreate, setShowCreate, created,
  } = useWorkspaceTodos(projects);

  if (selected) {
    return (
      <div className="flex h-full min-h-0 flex-col">
        <TodoDetail
          todo={detail}
          loading={loadingDetail}
          loadError={detailError}
          dir={selected.dir}
          onChanged={updateItem}
          onDeleted={deleted}
          onBack={() => select(null)}
        />
      </div>
    );
  }

  // Severity/age grouping flattens todos across workspaces into buckets; the
  // default 'workspace' grouping keeps the per-workspace sections (the only mode
  // that supports batch runs on the dashboard).
  const buckets = groupBy === 'workspace' ? null : bucketTodos(flattenTodos(workspaces, byDir), groupBy, Date.now());
  // Waiting on either fetch is "loading": the workspaces come from /api/projects
  // and the todos from the per-workspace lists.
  const pending = loadingList || !projectsLoaded;
  const StatusIcon = pending ? Spinner : UiCheck;

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
            <UiAdd className="text-xs" />
            New
          </Button>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-y-auto">
        {error && (
          <div role="alert" className="border-b border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive">
            {error}
          </div>
        )}
        {workspaces.length === 0 ? (
          projectError ? (
            <div role="alert" className="px-3 py-6 text-center text-xs text-destructive">{projectError}</div>
          ) : (
            <div className="px-3 py-6 text-center text-xs text-muted-foreground">
              <StatusIcon className="mb-2 text-2xl" />
              <p>{pending ? 'Loading' : 'No workspaces configured'}</p>
            </div>
          )
        ) : buckets ? (
          buckets.length > 0 ? (
            <ListMenu>
              {buckets.map(bucket => (
                <TodoBucketGroup
                  key={bucket.key}
                  bucket={bucket}
                  selected={selected}
                  onSelect={entry => select({ dir: entry.workspace.dir, ref: entry.todo.ref })}
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
                onSelect={ref => select({ dir: ws.dir, ref })}
              />
            ))}
          </ListMenu>
        )}
      </div>

      <CreateTodoDialog open={showCreate} onClose={() => setShowCreate(false)} workspaces={workspaces} onCreated={created} />
    </div>
  );
}
