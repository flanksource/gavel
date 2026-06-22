import { useEffect, useState } from 'react';
import type { TodoItem, TodoStatus } from '../../types';
import { Markdown } from '../Markdown';
import { GavelIcon } from '../GavelIcon';
import { priorityClass, statusClass, statuses, statusLabel, todoQuery } from './format';

export function TodoDetail({
  todo,
  loading,
  dir,
  provider,
  onChanged,
  onDeleted,
}: {
  todo: TodoItem | null;
  loading: boolean;
  dir: string;
  provider: string;
  onChanged: (todo: TodoItem) => void;
  onDeleted: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const isGrite = todo?.provider === 'grite';

  useEffect(() => {
    setError('');
  }, [todo?.ref]);

  async function updateStatus(status: TodoStatus) {
    if (!todo || busy) return;
    setBusy(true);
    setError('');
    try {
      const response = await fetch(`/api/todos/item?${todoQuery(dir, provider)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, status }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Update failed');
      onChanged(data as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Update failed');
    } finally {
      setBusy(false);
    }
  }

  async function archiveTodo() {
    if (!todo || busy) return;
    if (!window.confirm(isGrite ? 'Archive this Grite issue?' : 'Delete this TODO file?')) return;
    setBusy(true);
    setError('');
    try {
      const params = new URLSearchParams(todoQuery(dir, provider));
      params.set('ref', todo.ref);
      const response = await fetch(`/api/todos/item?${params.toString()}`, { method: 'DELETE' });
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || 'Archive failed');
      }
      onDeleted();
    } catch (err: any) {
      setError(err?.message || 'Archive failed');
    } finally {
      setBusy(false);
    }
  }

  if (!todo) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        <div className="text-center">
          <GavelIcon name="codicon:check" className="mb-2 text-4xl" />
          <p>Select a todo</p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="sticky top-0 z-10 border-b border-border bg-background px-4 py-3">
        <div className="flex min-w-0 items-start gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <span className={`inline-flex rounded border px-1.5 py-0.5 text-[10px] font-medium uppercase ${statusClass(todo.status)}`}>
                {statusLabel(todo.status)}
              </span>
              <span className={`text-xs font-medium ${priorityClass(todo.priority)}`}>{todo.priority}</span>
              {todo.shortId && <span className="font-mono text-xs text-muted-foreground">{todo.shortId}</span>}
            </div>
            <h2 className="mt-1 truncate text-base font-semibold text-foreground">{todo.title}</h2>
            <div className="mt-0.5 truncate text-xs text-muted-foreground">{todo.filePath || todo.cwd || todo.provider}</div>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <select
              value={todo.status}
              disabled={busy}
              onChange={e => updateStatus((e.target as HTMLSelectElement).value as TodoStatus)}
              className="h-8 rounded-md border border-border bg-background px-2 text-xs"
              aria-label="Update todo status"
            >
              {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
            </select>
            <button
              type="button"
              onClick={archiveTodo}
              disabled={busy}
              title={isGrite ? 'Archive issue' : 'Delete file'}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
              aria-label={isGrite ? 'Archive issue' : 'Delete file'}
            >
              <GavelIcon name={busy ? 'svg-spinners:ring-resize' : 'codicon:trash'} className="text-sm" />
            </button>
          </div>
        </div>
        {error && <div className="mt-2 text-xs text-red-600">{error}</div>}
      </div>
      <div className="px-4 py-3">
        {loading ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <GavelIcon name="svg-spinners:ring-resize" />
            Loading
          </div>
        ) : (
          <>
            {todo.body ? (
              <Markdown text={todo.body} className="text-sm" />
            ) : (
              <div className="text-sm text-muted-foreground">No body</div>
            )}
            {todo.events && todo.events.length > 0 && (
              <div className="mt-4 border-t border-border pt-3">
                <div className="mb-2 text-xs font-semibold uppercase text-muted-foreground">History</div>
                <div className="space-y-2">
                  {todo.events.map((event, i) => (
                    <div key={`${event.id || i}`} className="text-xs">
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <span className="font-medium text-foreground">{event.kind || 'Event'}</span>
                        {event.short_id && <span className="font-mono">{event.short_id}</span>}
                        {event.actor && <span>{event.actor}</span>}
                        {event.timestamp && <span>{new Date(event.timestamp).toLocaleString()}</span>}
                      </div>
                      {event.body && <div className="mt-1 whitespace-pre-wrap text-muted-foreground">{event.body}</div>}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
