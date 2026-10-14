import { useQuery } from '@tanstack/react-query';
import { StatusIcon } from './todos/format';
import { TodoRouteLink } from './todos/TodoChildren';
import { prLinkedTodosQueryOptions } from './todos/todoLinkQueries';

// The todos that point at this pull request, whether they implement it or only
// relate to it, and who linked them. Draws nothing when there are none, so a PR
// nobody tracks keeps its page unchanged.
export function PRLinkedTodos({ prUrl, onOpenTodo }: { prUrl: string; onOpenTodo?: (ref: string) => void }) {
  const { data: todos = [], error } = useQuery(prLinkedTodosQueryOptions(prUrl));
  if (error) return <p className="mt-4 text-xs text-red-600">{error.message}</p>;
  if (todos.length === 0) return null;
  return (
    <section className="mt-4">
      <div className="mb-2 flex items-center justify-between border-b border-border pb-1">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Linked TODOs</h3>
      </div>
      <ul className="divide-y divide-border">
        {todos.map(todo => (
          <li key={todo.id} className="flex flex-wrap items-center gap-x-2 gap-y-1 py-1.5 text-xs">
            <StatusIcon status={todo.status} />
            <TodoRouteLink
              todoRef={todo.shortId || todo.id}
              onOpen={onOpenTodo}
              title={todo.title}
              className="min-w-0 flex-1 basis-40 truncate text-sm font-medium text-foreground hover:underline"
            >
              {todo.title}
            </TodoRouteLink>
            <span className="shrink-0 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wide text-muted-foreground">{todo.relation}</span>
            <span className="text-[11px] text-muted-foreground" title={`Linked by ${todo.source}`}>{todo.source}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
