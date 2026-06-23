import { useEffect, useState, type ReactNode } from 'react';
import type { Project, TodoItem, TodoPriority, TodoRunOptions, TodoStatus } from '../../types';
import { Markdown } from '../Markdown';
import { GavelIcon } from '../GavelIcon';
import { TodoTimeline } from './TodoTimeline';
import { TodoSession } from './TodoSession';
import { TodoSessionTimer } from './TodoSessionTimer';
import { priorities, priorityClass, statusClass, statuses, statusLabel, todoQuery } from './format';
import { TodoRunAdvancedDialog, TodoRunSplitButton, defaultRunOptions, useTodoRun } from './run';
import { TodoCommentBox, TodoEditForm } from './TodoCompose';

export function TodoDetail({
  todo,
  loading,
  dir,
  provider,
  onChanged,
  onDeleted,
  workspaces = [],
  onTransferred,
}: {
  todo: TodoItem | null;
  loading: boolean;
  dir: string;
  provider: string;
  onChanged: (todo: TodoItem) => void;
  onDeleted: () => void;
  // workspaces/onTransferred are optional: the "Move to project" control only
  // renders where a caller wires them (the dashboard), not the compact menubar.
  workspaces?: Project[];
  onTransferred?: (toDir: string, todo: TodoItem) => void;
}) {
  const [busy, setBusy] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<'overview' | 'session'>('overview');
  const [editing, setEditing] = useState(false);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const { runBusy, runMessage, runError, reset: resetRun, run } = useTodoRun(dir, provider);
  const isGrite = todo?.provider === 'grite';
  // Projects this todo can move to: every configured workspace except its own.
  const transferTargets = workspaces.filter(ws => !!ws.dir && ws.dir !== dir);
  // A todo is "closed" when completed (Grite also reports providerState).
  const closed = todo?.status === 'completed' || todo?.providerState === 'closed';
  const body = todo?.body?.trim() ?? '';
  const events = todo?.events ?? [];

  useEffect(() => {
    setError('');
    resetRun();
    setAdvancedOpen(false);
    setTab('overview');
    setEditing(false);
  }, [todo?.ref, resetRun]);

  // patch sends a partial update (status, priority, title, body, and/or comment)
  // and adopts the server's returned todo so the view reflects provider-side side
  // effects (labels, state, rewritten body, new comment). Resolves true on success.
  async function patch(payload: {
    status?: TodoStatus;
    priority?: TodoPriority;
    title?: string;
    body?: string;
    comment?: string;
  }): Promise<boolean> {
    if (!todo || busy) return false;
    setBusy(true);
    setError('');
    try {
      const response = await fetch(`/api/todos/item?${todoQuery(dir, provider)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, ...payload }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Update failed');
      onChanged(data as TodoItem);
      return true;
    } catch (err: any) {
      setError(err?.message || 'Update failed');
      return false;
    } finally {
      setBusy(false);
    }
  }

  function startEditing() {
    if (!todo) return;
    setDraftTitle(todo.title);
    setDraftBody(todo.body ?? '');
    setTab('overview');
    setEditing(true);
  }

  async function saveEdit() {
    if (await patch({ title: draftTitle.trim(), body: draftBody })) {
      setEditing(false);
    }
  }

  async function transferTo(toDir: string) {
    if (!todo || busy || !toDir || !onTransferred) return;
    const target = transferTargets.find(ws => ws.dir === toDir);
    setBusy(true);
    setError('');
    try {
      const response = await fetch('/api/todos/transfer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ref: todo.ref,
          fromDir: dir,
          fromProvider: provider,
          toDir,
          toProvider: target?.todoProvider || 'auto',
        }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Move failed');
      onTransferred(toDir, data.todo as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Move failed');
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

  async function runTodo(options?: TodoRunOptions) {
    if (!todo) return;
    const result = await run([todo.ref], options);
    if (result?.status === 'started') {
      onChanged({
        ...todo,
        status: 'in_progress',
        lastRun: new Date().toISOString(),
        // Adopt the run's session id so the Session tab follows the new run.
        sessionId: result.sessionId || todo.sessionId,
      });
      // Surface the live session as soon as a run starts.
      setTab('session');
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
    <div className="flex h-full flex-col">
      <div className="shrink-0 border-b border-border bg-background px-4 py-3">
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
            {onTransferred && transferTargets.length > 0 && (
              <select
                value=""
                disabled={busy}
                onChange={e => transferTo((e.target as HTMLSelectElement).value)}
                className="h-8 rounded-md border border-border bg-background px-2 text-xs"
                aria-label="Move todo to another project"
                title="Move to project"
              >
                <option value="" disabled>Move to…</option>
                {transferTargets.map(ws => (
                  <option key={ws.dir} value={ws.dir}>{ws.name || ws.dir}</option>
                ))}
              </select>
            )}
            {todo.sessionId && (
              <button
                type="button"
                onClick={() => runTodo({ ...defaultRunOptions, resume: true })}
                disabled={busy || runBusy}
                title="Resume prior agent session"
                className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
                aria-label="Resume prior agent session"
              >
                <GavelIcon name={runBusy ? 'svg-spinners:ring-resize' : 'codicon:debug-continue'} className="text-sm" />
              </button>
            )}
            <TodoRunSplitButton
              disabled={busy || runBusy}
              loading={runBusy}
              onRun={runTodo}
              onAdvanced={() => setAdvancedOpen(true)}
            />
            <button
              type="button"
              onClick={() => (editing ? setEditing(false) : startEditing())}
              disabled={busy}
              title={editing ? 'Cancel edit' : 'Edit title & body'}
              className={`inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-muted hover:text-foreground disabled:opacity-50 ${editing ? 'text-foreground' : 'text-muted-foreground'}`}
              aria-label={editing ? 'Cancel edit' : 'Edit title & body'}
              aria-pressed={editing}
            >
              <GavelIcon name={editing ? 'codicon:close' : 'codicon:edit'} className="text-sm" />
            </button>
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
        {(error || runError) && <div className="mt-2 text-xs text-red-600">{error || runError}</div>}
        {runMessage && !error && !runError && <div className="mt-2 text-xs text-emerald-600">{runMessage}</div>}
        {todo.sessionId && <TodoSessionTimer dir={dir} provider={provider} sessionId={todo.sessionId} />}
      </div>
      <TodoRunAdvancedDialog
        open={advancedOpen}
        onClose={() => setAdvancedOpen(false)}
        onRun={options => {
          setAdvancedOpen(false);
          runTodo(options);
        }}
        loading={runBusy}
        dir={dir}
        provider={provider}
        refs={[todo.ref]}
      />
      <div className="flex shrink-0 gap-1 border-b border-border px-4 pt-2">
        <DetailTab active={tab === 'overview'} onClick={() => setTab('overview')} icon="codicon:list-flat" label="Overview" />
        <DetailTab active={tab === 'session'} onClick={() => setTab('session')} icon="codicon:comment-discussion" label="Session" />
      </div>
      {tab === 'session' ? (
        <TodoSession dir={dir} provider={provider} sessionId={todo.sessionId} active={tab === 'session'} />
      ) : (
        <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
          {loading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <GavelIcon name="svg-spinners:ring-resize" />
              Loading
            </div>
          ) : editing ? (
            <TodoEditForm
              title={draftTitle}
              body={draftBody}
              busy={busy}
              onTitle={setDraftTitle}
              onBody={setDraftBody}
              onSave={saveEdit}
              onCancel={() => setEditing(false)}
            />
          ) : (
            <div className="space-y-3">
              {body ? (
                <TodoSection title="Body" icon="codicon:markdown" defaultOpen resetKey={`${todo.ref}:body`}>
                  <Markdown text={body} className="text-sm" />
                </TodoSection>
              ) : (
                <div className="text-sm text-muted-foreground">No body</div>
              )}
              <TodoCommentBox
                closed={closed}
                busy={busy}
                onComment={(text, reopen) => patch(reopen ? { status: 'pending', comment: text } : { comment: text })}
              />
              {events.length > 0 && <TodoTimeline events={events} />}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function DetailTab({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: string; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`-mb-px inline-flex items-center gap-1.5 border-b-2 px-2.5 py-1.5 text-xs font-medium transition-colors ${
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      }`}
    >
      <GavelIcon name={icon} className="text-sm" />
      {label}
    </button>
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

