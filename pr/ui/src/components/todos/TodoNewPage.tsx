import { useMemo, useState } from 'react';
import { Field, Button } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { inputClass, priorities, statuses, statusLabel, todoQuery } from './format';

// firstParam reads the first present query value across a set of aliases so
// external callers can use the field name they have (e.g. ?body= or ?text=).
function firstParam(params: URLSearchParams, ...keys: string[]): string {
  for (const key of keys) {
    const value = params.get(key);
    if (value !== null && value.trim() !== '') return value.trim();
  }
  return '';
}

function parseBool(value: string): boolean {
  return /^(1|true|yes|on)$/i.test(value.trim());
}

// returnTarget resolves where the form returns to after a create or cancel: an
// explicit ?return= path wins, otherwise the (same-origin) referer that opened
// the page. Cross-origin referers and the new-todo page itself are ignored so a
// hard-load or external link falls through to the caller's default.
function returnTarget(params: URLSearchParams): string | null {
  const candidate = firstParam(params, 'return', 'returnTo') || (typeof document !== 'undefined' ? document.referrer : '');
  if (!candidate) return null;
  try {
    const url = new URL(candidate, window.location.origin);
    if (url.origin !== window.location.origin) return null;
    if (url.pathname.startsWith('/todos/new')) return null;
    return `${url.pathname}${url.search}${url.hash}`;
  } catch {
    return null;
  }
}

function oneOf<T extends string>(value: string, allowed: readonly T[], fallback: T): T {
  return (allowed as readonly string[]).includes(value) ? (value as T) : fallback;
}

// TodoNewPage is the focused, full-page todo form served at /todos/new. Unlike
// the in-dashboard CreateTodoDialog modal it is meant to be linked to (from the
// menubar, a bookmarklet, or another app): every field can be pre-filled from
// query params and, on submit or cancel, it navigates back to the referer (or an
// explicit ?return= path), falling back to the newly-created todo otherwise.
export function TodoNewPage({ projects }: { projects: Project[] }) {
  const workspaces = useMemo(() => projects.filter(p => !!p.dir), [projects]);
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const back = useMemo(() => returnTarget(params), [params]);
  const embed = useMemo(() => parseBool(firstParam(params, 'embed')), [params]);

  const queryDir = useMemo(() => firstParam(params, 'dir', 'workspace'), [params]);
  const queryProvider = useMemo(() => firstParam(params, 'provider', 'todoProvider'), [params]);
  const autoSave = useMemo(() => parseBool(firstParam(params, 'autoSave', 'autosave', 'auto_save')), [params]);

  // The workspace options are every configured workspace plus an explicit ?dir=
  // that isn't one of them (so external links can target any directory). An
  // empty value means the server's own work dir.
  const dirOptions = useMemo(() => {
    const opts = workspaces.map(w => ({ value: w.dir, label: w.name }));
    if (queryDir && !workspaces.some(w => w.dir === queryDir)) {
      opts.unshift({ value: queryDir, label: queryDir });
    }
    if (opts.length === 0) opts.push({ value: '', label: 'Default workspace' });
    return opts;
  }, [workspaces, queryDir]);

  const [dir, setDir] = useState(() => queryDir || workspaces[0]?.dir || '');
  const [title, setTitle] = useState(() => firstParam(params, 'title', 'name'));
  const [body, setBody] = useState(() => firstParam(params, 'body', 'description', 'text'));
  const [priority, setPriority] = useState<TodoPriority>(() => oneOf(firstParam(params, 'priority', 'severity'), priorities, 'medium'));
  const [status, setStatus] = useState<TodoStatus>(() => oneOf(firstParam(params, 'status'), statuses, autoSave ? 'pending' : 'draft'));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [created, setCreated] = useState(false);

  function providerForDir(target: string): string {
    const ws = workspaces.find(w => w.dir === target);
    return ws?.todoProvider || queryProvider || 'auto';
  }

  function leave(to: string) {
    window.location.href = to;
  }

  // In embed mode (rendered inside the React Grab dialog iframe) the form reports
  // its outcome to the parent window — which closes the dialog — instead of
  // navigating this iframe.
  function finish(type: 'todo-created' | 'cancel', ref?: string) {
    window.parent.postMessage({ source: 'gavel-react-grab', type, ref }, '*');
  }

  function cancel() {
    if (embed) finish('cancel');
    else leave(back ?? '/todos');
  }

  async function submit() {
    if (!title.trim() || busy) return;
    setBusy(true);
    setError('');
    try {
      const response = await fetch(`/api/todos/new?${todoQuery(dir, providerForDir(dir))}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, body, priority, status }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Create failed');
      const todo = data.todo as TodoItem | undefined;
      if (embed) {
        setCreated(true);
        finish('todo-created', todo?.ref);
        return;
      }
      leave(back ?? (todo?.ref ? `/todos/${encodeURIComponent(todo.ref)}` : '/todos'));
    } catch (err: any) {
      setError(err?.message || 'Create failed');
      setBusy(false);
    }
  }

  if (created) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="text-sm font-medium">Todo created ✓</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <img src="/brand/gavel-logo.svg" alt="gavel" className="h-7 shrink-0" />
          <span className="text-sm font-semibold">New todo</span>
        </div>
        <a
          href={back ?? '/todos'}
          onClick={e => {
            if (embed) {
              e.preventDefault();
              finish('cancel');
            }
          }}
          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          title={back ? 'Back' : 'Back to todos'}
          aria-label={back ? 'Back' : 'Back to todos'}
        >
          <GavelIcon name="codicon:close" className="text-base" />
        </a>
      </header>

      <main className="mx-auto w-full max-w-2xl px-4 py-6">
        <form
          className="space-y-4"
          onSubmit={e => {
            e.preventDefault();
            void submit();
          }}
        >
          {error && <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{error}</div>}
          <Field label="Workspace">
            <select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace">
              {dirOptions.map(o => <option key={o.value || '(default)'} value={o.value}>{o.label}</option>)}
            </select>
          </Field>
          <Field label="Title" required>
            <input
              className={inputClass}
              value={title}
              placeholder="What needs doing?"
              onChange={e => setTitle(e.currentTarget.value)}
              autoFocus
            />
          </Field>
          <div className="flex gap-3">
            <div className="flex-1">
              <Field label="Priority">
                <select value={priority} onChange={e => setPriority(e.currentTarget.value as TodoPriority)} className={inputClass} aria-label="Priority">
                  {priorities.map(p => <option key={p} value={p}>{p}</option>)}
                </select>
              </Field>
            </div>
            <div className="flex-1">
              <Field label="Status">
                <select value={status} onChange={e => setStatus(e.currentTarget.value as TodoStatus)} className={inputClass} aria-label="Status">
                  {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
                </select>
              </Field>
            </div>
          </div>
          <Field label="Body">
            <textarea
              className={`${inputClass} h-40 resize-none`}
              value={body}
              placeholder="Details (optional)"
              onChange={e => setBody(e.currentTarget.value)}
            />
          </Field>
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" onClick={cancel}>Cancel</Button>
            <Button type="submit" loading={busy} disabled={!title.trim()}>Add todo</Button>
          </div>
        </form>
      </main>
    </div>
  );
}
