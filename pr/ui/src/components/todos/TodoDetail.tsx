import { useEffect, useState, type ReactNode } from 'react';
import type { TodoEvent, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { Markdown } from '../Markdown';
import { GavelIcon } from '../GavelIcon';
import { priorities, priorityClass, statusClass, statuses, statusLabel, todoQuery } from './format';

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
  // A todo is "closed" when completed (Grite also reports providerState).
  const closed = todo?.status === 'completed' || todo?.providerState === 'closed';
  const body = todo?.body?.trim() ?? '';
  const events = todo?.events ?? [];

  useEffect(() => {
    setError('');
  }, [todo?.ref]);

  // patch sends a partial update (status and/or priority) and adopts the server's
  // returned todo so the view reflects provider-side side effects (labels, state).
  async function patch(body: { status?: TodoStatus; priority?: TodoPriority }) {
    if (!todo || busy) return;
    setBusy(true);
    setError('');
    try {
      const response = await fetch(`/api/todos/item?${todoQuery(dir, provider)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, ...body }),
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
              onChange={e => patch({ status: (e.target as HTMLSelectElement).value as TodoStatus })}
              className="h-8 rounded-md border border-border bg-background px-2 text-xs"
              aria-label="Update todo status"
            >
              {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
            </select>
            <select
              value={todo.priority}
              disabled={busy}
              onChange={e => patch({ priority: (e.target as HTMLSelectElement).value as TodoPriority })}
              className="h-8 rounded-md border border-border bg-background px-2 text-xs"
              aria-label="Update todo severity"
              title="Severity"
            >
              {priorities.map(p => <option key={p} value={p}>{p}</option>)}
            </select>
            <button
              type="button"
              onClick={() => patch({ status: closed ? 'pending' : 'completed' })}
              disabled={busy}
              title={closed ? 'Reopen todo' : 'Close todo'}
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
              aria-label={closed ? 'Reopen todo' : 'Close todo'}
            >
              <GavelIcon name={busy ? 'svg-spinners:ring-resize' : closed ? 'codicon:debug-restart' : 'codicon:pass'} className="text-sm" />
            </button>
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
          <div className="space-y-3">
            {body ? (
              <TodoSection title="Body" icon="codicon:markdown" defaultOpen resetKey={`${todo.ref}:body`}>
                <Markdown text={body} className="text-sm" />
              </TodoSection>
            ) : (
              <div className="text-sm text-muted-foreground">No body</div>
            )}
            {events.length > 0 && (
              <TodoSection
                title="Context"
                icon="codicon:history"
                count={events.length}
                defaultOpen={!body}
                resetKey={`${todo.ref}:context`}
              >
                <div className="divide-y divide-border">
                  {events.map((event, i) => (
                    <TodoEventBlock key={`${event.id || i}`} event={event} index={i} />
                  ))}
                </div>
              </TodoSection>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function TodoSection({
  title,
  icon,
  count,
  defaultOpen = false,
  resetKey,
  children,
}: {
  title: string;
  icon: string;
  count?: number;
  defaultOpen?: boolean;
  resetKey: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(defaultOpen);

  useEffect(() => {
    setOpen(defaultOpen);
  }, [defaultOpen, resetKey]);

  return (
    <section className="rounded-md border border-border bg-background">
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="flex w-full min-w-0 items-center gap-2 px-3 py-2 text-left hover:bg-muted"
        aria-expanded={open}
      >
        <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="shrink-0 text-xs text-muted-foreground" />
        <GavelIcon name={icon} className="shrink-0 text-xs text-muted-foreground" />
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">{title}</span>
        {typeof count === 'number' && <span className="text-xs tabular-nums text-muted-foreground">{count}</span>}
      </button>
      {open && <div className="border-t border-border px-3 py-3">{children}</div>}
    </section>
  );
}

function TodoEventBlock({ event, index }: { event: TodoEvent; index: number }) {
  const body = event.body?.trim() ?? '';
  const timestamp = formatEventTimestamp(event.timestamp);
  const title = event.title?.trim();

  return (
    <div className="py-3 first:pt-0 last:pb-0">
      <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{event.kind || `Event ${index + 1}`}</span>
        {event.short_id && <span className="font-mono">{event.short_id}</span>}
        {event.actor && <span>{event.actor}</span>}
        {timestamp && <span>{timestamp}</span>}
      </div>
      {title && <div className="mt-1 text-xs font-medium text-foreground">{title}</div>}
      {body && <Markdown text={body} className="mt-2 text-xs" />}
    </div>
  );
}

function formatEventTimestamp(timestamp?: string) {
  if (!timestamp || timestamp.startsWith('0001-01-01')) return '';
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) return '';
  return parsed.toLocaleString();
}
