import { useCallback, useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Field, Button, Select } from '@flanksource/clicky-ui/components';
import { UiClose } from '@flanksource/clicky-ui/icons';
import { useProcStatus } from '../../procStatusQuery';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { ScreenshotPicker, todoCommentFormData, todoFormData, useAttachments } from './attachments';
import { inputClass, priorities, statuses, statusLabel } from './format';
import { RecentTodoChips } from './RecentTodoChips';
import { loadRecentTodos, recentTodoHost, rememberRecentTodo, type RecentTodo } from './recentTodos';
import { useTodoTagCounts, useTodoTagIndex } from './tagQueries';
import { TagListField } from './TodoTag';
import { TodoBodyField } from './TodoBodyField';
import { useCreateTodoMutation, useUpdateTodoMutation } from './todoMutations';
import {
  closedStatuses, compareRecentTodos, detectProjectDir, firstParam, listParam, modeFromParams,
  oneOf, parseBool, returnTarget, todoMatchesSearch, type TodoNewMode,
} from './todoNewParams';
import { todoListQueryOptions } from './todoQueries';

// TodoNewPage is the focused, full-page todo form served at /todos/new. Unlike
// the in-dashboard CreateTodoDialog modal it is meant to be linked to (from the
// menubar, a bookmarklet, or another app): every field can be pre-filled from
// query params and, on submit or cancel, it navigates back to the referer (or an
// explicit ?return= path), falling back to the newly-created todo otherwise.
export function TodoNewPage({ projects, projectError = '' }: { projects: Project[]; projectError?: string }) {
  const procStatus = useProcStatus();
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
  const recentHost = useMemo(() => recentTodoHost(sourceUrl), [sourceUrl]);
  const [recent, setRecent] = useState(() => loadRecentTodos(recentHost));
  // The workspace last used from this host:port. Port detection needs
  // /api/proc/status; until it answers this is known synchronously (and shows
  // its chips). With neither, nothing is guessed: the user picks.
  const recentDir = useMemo(
    () => recent.find(entry => workspaces.some(w => w.dir === entry.dir))?.dir || '',
    [recent, workspaces],
  );
  const preferredDir = queryDir || queryProjectDir || detectedDir || recentDir;

  // The workspace options are every configured workspace plus an explicit ?dir=
  // that isn't one of them (so external links can target any directory). An
  // empty value means the server's own work dir — but only when the catalog
  // actually loaded: a failed projects request looks exactly like "nothing is
  // configured", so it says so instead of quietly filing the todo elsewhere.
  // With workspaces to choose from, empty is instead the unpicked placeholder.
  const dirOptions = useMemo(() => {
    const opts = workspaces.map(w => ({ value: w.dir, label: w.name, disabled: false }));
    if (queryDir && !workspaces.some(w => w.dir === queryDir)) {
      opts.unshift({ value: queryDir, label: queryDir, disabled: false });
    }
    if (opts.length === 0) return [{ value: '', label: projectError ? 'Workspaces unavailable' : 'Default workspace', disabled: false }];
    return [{ value: '', label: 'Select workspace', disabled: true }, ...opts];
  }, [workspaces, queryDir, projectError]);
  const workspaceRequired = dirOptions.some(o => o.value !== '');

  const [mode, setMode] = useState<TodoNewMode>(initialMode);
  const [dir, setDirState] = useState(() => preferredDir);
  const [dirTouched, setDirTouched] = useState(() => !!queryDir || !!queryProjectDir);
  const [title, setTitle] = useState(() => firstParam(params, 'title', 'name'));
  const [body, setBody] = useState(() => firstParam(params, 'body', 'description', 'text'));
  const [priority, setPriority] = useState<TodoPriority>(() => oneOf(firstParam(params, 'priority', 'severity'), priorities, 'medium'));
  const [status, setStatus] = useState<TodoStatus>(() => oneOf(firstParam(params, 'status'), statuses, autoSave ? 'pending' : 'draft'));
  const [labels, setLabels] = useState(() => listParam(params, 'labels', 'label'));
  const [todoSearch, setTodoSearch] = useState('');
  const [selectedRef, setSelectedRef] = useState(queryRef);
  const [error, setError] = useState('');
  const [completeMessage, setCompleteMessage] = useState('');
  const recentForDir = useMemo(() => recent.filter(entry => entry.dir === dir), [recent, dir]);
  const { attachments, previews, add, remove } = useAttachments({ pasteAnywhere: true });
  const workspaceMissing = workspaceRequired && !dir;
  const tagIndex = useTodoTagIndex(dir, mode === 'new' && !workspaceMissing);
  const tagCounts = useTodoTagCounts(dir, mode === 'new' && !workspaceMissing);
  // The list also backs the recent chips, so it loads whenever there are chips
  // to refresh — their titles stay current and closed or deleted todos drop out.
  const existingTodosQuery = useQuery({
    ...todoListQueryOptions(dir),
    enabled: (mode === 'existing' || recentForDir.length > 0) && !!dir,
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

  // Until the list loads the chips show what was stored; once it has, only
  // todos that are still open remain, under their current title.
  const recentChips = useMemo((): RecentTodo[] => {
    if (!existingTodosQuery.data) return recentForDir;
    const open = new Map(existingTodos.filter(todo => !closedStatuses.has(todo.status)).map(todo => [todo.ref, todo]));
    return recentForDir.flatMap(entry => {
      const todo = open.get(entry.ref);
      return todo ? [{ ...entry, title: todo.title, shortId: todo.shortId || entry.shortId }] : [];
    });
  }, [existingTodosQuery.data, existingTodos, recentForDir]);

  function pickRecent(ref: string) {
    setMode('existing');
    setTodoSearch('');
    setSelectedRef(ref);
  }

  function remember(todo: Pick<TodoItem, 'ref' | 'title'> & { shortId?: string }) {
    setRecent(rememberRecentTodo(recentHost, { ref: todo.ref, shortId: todo.shortId, title: todo.title, dir, at: Date.now() }));
  }

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
    if (workspaceMissing) return;
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
        ? { body: todoFormData({ title, body, priority, status, labels }, attachments) }
        : {
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, body, priority, status, labels }),
          });
      const todo = data.todo as TodoItem | undefined;
      if (todo?.ref) remember(todo);
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
      const known = todo?.title ? todo : existingTodos.find(t => t.ref === ref);
      if (known) remember({ ...known, ref });
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

  const submitDisabled = workspaceMissing || (mode === 'existing'
    ? !selectedRef || busy || (!body.trim() && attachments.length === 0)
    : !title.trim() || busy);

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
              {dirOptions.map(o => <option key={o.value || '(default)'} value={o.value} disabled={o.disabled}>{o.label}</option>)}
            </Select>
          </Field>
          {mode === 'existing' ? (
            <>
              <Field label="Issue">
                <div className="space-y-2">
                  <RecentTodoChips label="Recent" entries={recentChips} selectedRef={selectedRef} disabled={busy} onPick={pickRecent} />
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
              {showModeSwitch && (
                <RecentTodoChips label="Add to recent" entries={recentChips} disabled={busy} onPick={pickRecent} />
              )}
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
              <Field label="Labels">
                <TagListField labels={labels} index={tagIndex} counts={tagCounts} disabled={busy} onChange={setLabels} />
              </Field>
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
