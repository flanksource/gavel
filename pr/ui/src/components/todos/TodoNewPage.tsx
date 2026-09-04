import { useCallback, useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Field, Button, Select } from '@flanksource/clicky-ui/components';
import { UiClose } from '@flanksource/clicky-ui/icons';
import type { ProcStatus, Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { ScreenshotPicker, todoCommentFormData, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { TodoBodyField } from './TodoBodyField';
import { useCreateTodoMutation, useUpdateTodoMutation } from './todoMutations';
import { todoListQueryOptions } from './todoQueries';

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

type TodoNewMode = 'new' | 'existing';

const closedStatuses = new Set<TodoStatus>(['completed', 'skipped']);

function modeFromParams(params: URLSearchParams): TodoNewMode {
  const mode = firstParam(params, 'mode', 'target');
  if (mode === 'existing' || firstParam(params, 'ref', 'todo', 'issue')) return 'existing';
  return 'new';
}

function sourcePort(raw: string): number | null {
  if (!raw) return null;
  try {
    const url = new URL(raw);
    if (url.port) return Number(url.port);
    if (url.protocol === 'http:') return 80;
    if (url.protocol === 'https:') return 443;
  } catch {
    return null;
  }
  return null;
}

function statusForProject(project: Project, procStatus: Record<string, ProcStatus>): ProcStatus | undefined {
  if (procStatus[project.name]) return procStatus[project.name];
  for (const repo of project.repos || []) {
    if (procStatus[repo]) return procStatus[repo];
  }
  return undefined;
}

function detectProjectDir(sourceUrl: string, workspaces: Project[], procStatus: Record<string, ProcStatus>): string {
  const port = sourcePort(sourceUrl);
  if (!port) return '';
  for (const workspace of workspaces) {
    const status = statusForProject(workspace, procStatus);
    if (status?.processes?.some(proc => proc.ports?.includes(port))) return workspace.dir;
  }
  return '';
}

function todoMatchesSearch(todo: TodoItem, query: string): boolean {
  const q = query.trim().toLowerCase();
  if (!q) return true;
  return todo.title.toLowerCase().includes(q) ||
    todo.ref.toLowerCase().includes(q) ||
    (todo.shortId || '').toLowerCase().includes(q) ||
    (todo.id || '').toLowerCase().includes(q);
}

function todoActivityMs(todo: TodoItem): number | null {
  const raw = todo.lastRun || todo.created;
  if (!raw) return null;
  const ms = Date.parse(raw);
  return Number.isNaN(ms) ? null : ms;
}

function compareRecentTodos(a: TodoItem, b: TodoItem): number {
  const am = todoActivityMs(a);
  const bm = todoActivityMs(b);
  if (am !== bm) {
    if (am === null) return 1;
    if (bm === null) return -1;
    return bm - am;
  }
  return (a.title || '').localeCompare(b.title || '');
}

// TodoNewPage is the focused, full-page todo form served at /todos/new. Unlike
// the in-dashboard CreateTodoDialog modal it is meant to be linked to (from the
// menubar, a bookmarklet, or another app): every field can be pre-filled from
// query params and, on submit or cancel, it navigates back to the referer (or an
// explicit ?return= path), falling back to the newly-created todo otherwise.
export function TodoNewPage({ projects, procStatus = {}, projectError = '' }: { projects: Project[]; procStatus?: Record<string, ProcStatus>; projectError?: string }) {
  const workspaces = useMemo(() => projects.filter(p => !!p.dir), [projects]);
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const back = useMemo(() => returnTarget(params), [params]);
  const embed = useMemo(() => parseBool(firstParam(params, 'embed')), [params]);

  const queryDir = useMemo(() => firstParam(params, 'dir', 'workspace'), [params]);
  const queryProject = useMemo(() => firstParam(params, 'project'), [params]);
  const queryRef = useMemo(() => firstParam(params, 'ref', 'todo', 'issue'), [params]);
  const sourceUrl = useMemo(() => firstParam(params, 'sourceUrl', 'sourceURL', 'url'), [params]);
  const autoSave = useMemo(() => parseBool(firstParam(params, 'autoSave', 'autosave', 'auto_save')), [params]);
  const initialMode = useMemo(() => modeFromParams(params), [params]);
  const queryProjectDir = useMemo(() => {
    if (!queryProject) return '';
    return workspaces.find(w => w.name === queryProject)?.dir || '';
  }, [queryProject, workspaces]);
  const detectedDir = useMemo(() => detectProjectDir(sourceUrl, workspaces, procStatus), [sourceUrl, workspaces, procStatus]);
  const preferredDir = queryDir || queryProjectDir || detectedDir || workspaces[0]?.dir || '';

  // The workspace options are every configured workspace plus an explicit ?dir=
  // that isn't one of them (so external links can target any directory). An
  // empty value means the server's own work dir — but only when the catalog
  // actually loaded: a failed projects request looks exactly like "nothing is
  // configured", so it says so instead of quietly filing the todo elsewhere.
  const dirOptions = useMemo(() => {
    const opts = workspaces.map(w => ({ value: w.dir, label: w.name }));
    if (queryDir && !workspaces.some(w => w.dir === queryDir)) {
      opts.unshift({ value: queryDir, label: queryDir });
    }
    if (opts.length === 0) opts.push({ value: '', label: projectError ? 'Workspaces unavailable' : 'Default workspace' });
    return opts;
  }, [workspaces, queryDir, projectError]);

  const [mode, setMode] = useState<TodoNewMode>(initialMode);
  const [dir, setDirState] = useState(() => preferredDir);
  const [dirTouched, setDirTouched] = useState(() => !!queryDir || !!queryProjectDir);
  const [title, setTitle] = useState(() => firstParam(params, 'title', 'name'));
  const [body, setBody] = useState(() => firstParam(params, 'body', 'description', 'text'));
  const [priority, setPriority] = useState<TodoPriority>(() => oneOf(firstParam(params, 'priority', 'severity'), priorities, 'medium'));
  const [status, setStatus] = useState<TodoStatus>(() => oneOf(firstParam(params, 'status'), statuses, autoSave ? 'pending' : 'draft'));
  const [todoSearch, setTodoSearch] = useState('');
  const [selectedRef, setSelectedRef] = useState(queryRef);
  const [error, setError] = useState('');
  const [completeMessage, setCompleteMessage] = useState('');
  const { attachments, previews, add, remove } = useAttachments({ pasteAnywhere: true });
  const existingTodosQuery = useQuery({
    ...todoListQueryOptions(dir),
    enabled: mode === 'existing' && !!dir,
  });
  const createTodo = useCreateTodoMutation(dir);
  const commentTodo = useUpdateTodoMutation(dir, `Failed to add comment to todo ${selectedRef}`);
  const existingTodos = existingTodosQuery.data?.items ?? [];
  const loadingTodos = existingTodosQuery.isFetching;
  const busy = createTodo.isPending || commentTodo.isPending;
  const visibleError = error || projectError || (existingTodosQuery.error instanceof Error ? existingTodosQuery.error.message : '');
  const showModeSwitch = embed || initialMode === 'existing';

  const setDir = useCallback((next: string) => {
    setDirTouched(true);
    setDirState(next);
  }, []);

  useEffect(() => {
    if (!dirTouched && preferredDir && dir !== preferredDir) {
      setDirState(preferredDir);
    }
  }, [dir, dirTouched, preferredDir]);

  // In embed mode the React Grab plugin hands us a captured screenshot: we tell
  // the parent we're ready, then receive the image Blob over postMessage (a Blob
  // survives the structured clone even cross-origin) and attach it to the create.
  useEffect(() => {
    if (!embed) return;
    function onMessage(e: MessageEvent) {
      const d = e.data;
      if (!d || d.source !== 'gavel-react-grab' || d.type !== 'attachment' || !d.blob) return;
      add([{ blob: d.blob as Blob, name: (d.name as string) || 'attachment' }]);
    }
    window.addEventListener('message', onMessage);
    window.parent.postMessage({ source: 'gavel-react-grab', type: 'embed-ready' }, '*');
    return () => window.removeEventListener('message', onMessage);
  }, [embed, add]);

  const selectableTodos = useMemo(
    () => existingTodos
      .filter(todo => !closedStatuses.has(todo.status))
      .filter(todo => todoMatchesSearch(todo, todoSearch))
      .sort(compareRecentTodos),
    [existingTodos, todoSearch],
  );

  useEffect(() => {
    if (mode !== 'existing' || !selectedRef || loadingTodos) return;
    if (existingTodos.length === 0) return;
    if (!existingTodos.some(todo => todo.ref === selectedRef)) setSelectedRef('');
  }, [existingTodos, loadingTodos, mode, selectedRef]);

  function leave(to: string) {
    window.location.href = to;
  }

  // In embed mode (rendered inside the React Grab dialog iframe) the form reports
  // its outcome to the parent window — which closes the dialog — instead of
  // navigating this iframe.
  function finish(type: 'todo-created' | 'todo-commented' | 'cancel', ref?: string) {
    window.parent.postMessage({ source: 'gavel-react-grab', type, ref }, '*');
  }

  function cancel() {
    if (embed) finish('cancel');
    else leave(back ?? '/todos');
  }

  async function submit() {
    if (mode === 'existing') {
      await submitExisting();
      return;
    }
    await submitNew();
  }

  async function submitNew() {
    if (!title.trim() || busy) return;
    setError('');
    try {
      // With attachments, post multipart so the image bytes ride along and the
      // server persists them; otherwise keep the lighter JSON path.
      const data = await createTodo.mutateAsync(attachments.length
        ? { body: todoFormData({ title, body, priority, status }, attachments) }
        : {
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, body, priority, status }),
          });
      const todo = data.todo as TodoItem | undefined;
      if (embed) {
        setCompleteMessage('Todo created');
        finish('todo-created', todo?.ref);
        return;
      }
      leave(back ?? (todo?.ref ? `/todos/${encodeURIComponent(todo.ref)}` : '/todos'));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create todo');
    }
  }

  async function submitExisting() {
    if (!selectedRef || busy || (!body.trim() && attachments.length === 0)) return;
    setError('');
    try {
      const data = await commentTodo.mutateAsync(attachments.length
        ? {
            ref: selectedRef,
            body: todoCommentFormData({ ref: selectedRef, comment: body }, attachments),
          }
        : {
            ref: selectedRef,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ref: selectedRef, comment: body }),
          });
      const todo = data as TodoItem | undefined;
      const ref = todo?.ref || selectedRef;
      if (embed) {
        setCompleteMessage('Comment added');
        finish('todo-commented', ref);
        return;
      }
      leave(back ?? `/todos/${encodeURIComponent(ref)}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add comment');
    }
  }

  const submitDisabled = mode === 'existing'
    ? !selectedRef || busy || (!body.trim() && attachments.length === 0)
    : !title.trim() || busy;

  if (completeMessage) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-foreground">
        <div className="text-sm font-medium">{completeMessage}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <img src="/brand/gavel-logo.svg" alt="gavel" className="h-7 shrink-0" />
          <span className="text-sm font-semibold">{mode === 'existing' ? 'Add to issue' : 'New todo'}</span>
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
          <UiClose className="text-base" />
        </a>
      </header>

      <main className="mx-auto w-full max-w-4xl px-4 py-6">
        <form
          className="space-y-4"
          onSubmit={e => {
            e.preventDefault();
            void submit();
          }}
        >
          {visibleError && <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{visibleError}</div>}
          {showModeSwitch && (
            <div className="grid grid-cols-2 rounded-md border border-border bg-muted p-1">
              {(['new', 'existing'] as const).map(next => (
                <Button
                  key={next}
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setMode(next)}
                  className={`h-8 rounded px-3 text-sm font-medium ${mode === next ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                >
                  {next === 'new' ? 'New issue' : 'Existing issue'}
                </Button>
              ))}
            </div>
          )}
          <Field label="Workspace">
            <Select value={dir} onChange={e => setDir(e.currentTarget.value)} className={inputClass} aria-label="Workspace" disabled={!!projectError}>
              {dirOptions.map(o => <option key={o.value || '(default)'} value={o.value}>{o.label}</option>)}
            </Select>
          </Field>
          {mode === 'existing' ? (
            <>
              <Field label="Issue">
                <div className="space-y-2">
                  <input
                    className={inputClass}
                    value={todoSearch}
                    placeholder="Search issues"
                    onChange={e => setTodoSearch(e.currentTarget.value)}
                    autoFocus
                  />
                  <Select
                    value={selectedRef}
                    onChange={e => setSelectedRef(e.currentTarget.value)}
                    className={inputClass}
                    aria-label="Issue"
                    disabled={loadingTodos}
                  >
                    <option value="">{loadingTodos ? 'Loading issues...' : 'Select issue'}</option>
                    {selectableTodos.map(todo => (
                      <option key={todo.ref} value={todo.ref}>
                        {todo.shortId || todo.ref} - {todo.title}
                      </option>
                    ))}
                  </Select>
                </div>
              </Field>
              <TodoBodyField
                label="Comment"
                value={body}
                onChange={setBody}
                placeholder="Comment"
                disabled={busy}
              />
            </>
          ) : (
            <>
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
                    <Select value={priority} onChange={e => setPriority(e.currentTarget.value as TodoPriority)} className={inputClass} aria-label="Priority">
                      {priorities.map(p => <option key={p} value={p}>{p}</option>)}
                    </Select>
                  </Field>
                </div>
                <div className="flex-1">
                  <Field label="Status">
                    <Select value={status} onChange={e => setStatus(e.currentTarget.value as TodoStatus)} className={inputClass} aria-label="Status">
                      {statuses.map(s => <option key={s} value={s}>{statusLabel(s)}</option>)}
                    </Select>
                  </Field>
                </div>
              </div>
              <TodoBodyField
                label="Body"
                value={body}
                onChange={setBody}
                placeholder="Details (optional)"
                disabled={busy}
              />
            </>
          )}
          <Field label="Screenshot">
            <ScreenshotPicker previews={previews} onAdd={add} onRemove={remove} disabled={busy} />
          </Field>
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" onClick={cancel}>Cancel</Button>
            <Button type="submit" loading={busy} disabled={submitDisabled}>{mode === 'existing' ? 'Add comment' : 'Add todo'}</Button>
          </div>
        </form>
      </main>
    </div>
  );
}
